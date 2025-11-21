package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/vechain/networkhub/network/node"
	nodegenesis "github.com/vechain/networkhub/network/node/genesis"
)

// NewDockerNode initializes a new DockerNode
func NewDockerNode(cfg node.Config, enodes []string, networkID string, exposedPorts *ExposedPort, ipAddr string) *Node {
	return &Node{
		cfg:          cfg,
		enodes:       enodes,
		networkID:    networkID,
		exposedPorts: exposedPorts,
		ipAddr:       ipAddr,
	}
}

// Node represents a Docker container node
type Node struct {
	cfg          node.Config
	enodes       []string
	id           string
	networkID    string
	exposedPorts *ExposedPort
	ipAddr       string
}

// Start runs the node as a Docker container
func (n *Node) Start() error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Check if the Docker image is available locally
	_, err = cli.ImageInspect(ctx, n.cfg.GetExecArtifact())
	if err != nil {
		if client.IsErrNotFound(err) {
			// Pull the Docker image
			out, err := cli.ImagePull(ctx, n.cfg.GetExecArtifact(), image.PullOptions{})
			if err != nil {
				return fmt.Errorf("failed to pull Docker image: %w", err)
			}
			defer out.Close()

			// We wait for the image to be pulled
			decoder := json.NewDecoder(out)
			for {
				var event map[string]interface{}
				if err := decoder.Decode(&event); err == io.EOF {
					break
				} else if err != nil {
					return fmt.Errorf("failed to decode image pull event: %w", err)
				}
			}
		} else {
			return fmt.Errorf("failed to inspect Docker image: %w", err)
		}
	}

	cleanEnode := n.cleanEnodes()

	// Build thor command arguments
	thorArgs := n.buildThorArgs(cleanEnode)

	// Build the full shell command
	shellCommands := []string{
		"cd /home/thor",
		"echo $GENESIS > genesis.json",
		"echo $PRIVATEKEY > master.key",
		"echo $PRIVATEKEY > p2p.key",
		strings.Join(thorArgs, " "),
	}

	cmd := []string{"sh", "-c", strings.Join(shellCommands, "; ")}

	//serialize genesis
	genesisBytes, err := nodegenesis.Marshal(n.cfg.GetGenesis())
	if err != nil {
		return fmt.Errorf("unable to marshal genesis - %w", err)
	}

	exposedPorts := nat.PortSet{
		nat.Port(fmt.Sprintf("%s/tcp", n.exposedPorts.ContainerPort)): struct{}{},
	}
	portBindings := map[nat.Port][]nat.PortBinding{
		nat.Port(fmt.Sprintf("%s/tcp", n.exposedPorts.ContainerPort)): {
			{
				HostPort: n.exposedPorts.HostPort,
			},
		},
	}

	// Construct Docker container configuration
	config := &container.Config{
		Image:      n.cfg.GetExecArtifact(),
		Cmd:        cmd,
		Entrypoint: []string{},
		Env: []string{
			fmt.Sprintf("GENESIS=%s", string(genesisBytes)),
			fmt.Sprintf("PRIVATEKEY=%s", n.cfg.GetKey()),
		},
		ExposedPorts: exposedPorts,
		Hostname:     fmt.Sprintf("thor-%s", n.cfg.GetID()),
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
	}

	// Define the network configuration
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			n.networkID: {
				IPAMConfig: &network.EndpointIPAMConfig{
					IPv4Address: n.ipAddr,
				},
			},
		},
	}

	// Create the Docker container
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, networkConfig, nil, n.cfg.GetID())
	if err != nil {
		return fmt.Errorf("failed to create Docker container: %w", err)
	}

	n.id = resp.ID

	// Start the Docker container
	if err := cli.ContainerStart(ctx, n.id, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start Docker container: %w", err)
	}

	return nil
}

// Stop stops the Docker container
func (n *Node) Stop() error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Stop the Docker container
	if err := cli.ContainerStop(ctx, n.id, container.StopOptions{}); err != nil {
		return fmt.Errorf("failed to stop Docker container: %w", err)
	}

	// Remove the Docker container
	if err := cli.ContainerRemove(ctx, n.id, container.RemoveOptions{}); err != nil {
		return fmt.Errorf("failed to remove Docker container: %w", err)
	}

	return nil
}

// cleanEnodes filters out the current node's enode from the list
func (n *Node) cleanEnodes() []string {
	var cleanEnodes []string
	for _, enode := range n.enodes {
		nodeEnode, err := n.cfg.Enode(n.ipAddr)
		if err != nil {
			continue // Skip invalid enodes
		}
		if nodeEnode != enode {
			cleanEnodes = append(cleanEnodes, enode)
		}
	}
	return cleanEnodes
}

// buildThorArgs builds the thor command arguments array
func (n *Node) buildThorArgs(cleanEnodes []string) []string {
	args := []string{"thor"}

	// Add network parameter
	args = append(args, "--network", "genesis.json")

	// Add common arguments
	args = append(args,
		"--nat", "none",
		"--config-dir", n.cfg.GetConfigDir(),
		"--api-addr", n.cfg.GetAPIAddr(),
		fmt.Sprintf("--api-cors '%s' ", n.cfg.GetAPICORS()),
		"--verbosity", fmt.Sprintf("%d", n.cfg.GetVerbosity()),
		"--p2p-port", fmt.Sprintf("%d", n.cfg.GetP2PListenPort()),
	)

	// Add bootnodes if any
	if len(cleanEnodes) > 0 {
		enodeString := strings.Join(cleanEnodes, ",")
		args = append(args, "--bootnode", enodeString)
	}

	// Add additional arguments
	for key, value := range n.cfg.GetAdditionalArgs() {
		args = append(args, fmt.Sprintf("--%s", key), value)
	}

	return args
}
