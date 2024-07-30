package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/vechain/networkhub/network/node"
)

// NewDockerNode initializes a new DockerNode
func NewDockerNode(cfg node.Node, enodes []string, networkID string, exposedPorts *exposedPort, ipAddr string) *Node {
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
	cfg          node.Node
	enodes       []string
	id           string
	networkID    string
	exposedPorts *exposedPort
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
	_, _, err = cli.ImageInspectWithRaw(ctx, n.cfg.GetExecArtifact())
	if err != nil {
		if client.IsErrNotFound(err) {
			// Pull the Docker image
			_, err = cli.ImagePull(ctx, n.cfg.GetExecArtifact(), image.PullOptions{})
			if err != nil {
				return fmt.Errorf("failed to pull Docker image: %w", err)
			}
		} else {
			return fmt.Errorf("failed to inspect Docker image: %w", err)
		}
	}

	cleanEnode := []string{} // todo theres a clever way of doing this
	for _, enode := range n.enodes {
		nodeEnode, err := n.cfg.Enode(n.ipAddr)
		if err != nil {
			return err
		}
		if nodeEnode != enode {
			cleanEnode = append(cleanEnode, enode)
		}
	}
	enodeString := strings.Join(cleanEnode, ",")

	cmd := []string{
		"sh", "-c",
		"cd /home/thor; " +
			"echo $GENESIS > genesis.json;" +
			"echo $PRIVATEKEY > master.key;" +
			"echo $PRIVATEKEY > p2p.key;" +
			"thor " +
			"--network genesis.json " +
			"--nat none " +
			fmt.Sprintf("--config-dir='%s' ", n.cfg.GetConfigDir()) +
			fmt.Sprintf("--api-addr='%s' ", n.cfg.GetAPIAddr()) +
			fmt.Sprintf("--api-cors='%s' ", n.cfg.GetAPICORS()) +
			fmt.Sprintf("--p2p-port=%d ", n.cfg.GetP2PListenPort()) +
			fmt.Sprintf("--bootnode=%s", enodeString),
	}

	//serialize genesis
	genesisBytes, err := json.Marshal(n.cfg.GetGenesis())
	if err != nil {
		return fmt.Errorf("unable to marshal genesis - %w", err)
	}

	exposedPorts := nat.PortSet{
		nat.Port(fmt.Sprintf("%s/tcp", n.exposedPorts.containerPort)): struct{}{},
	}
	portBindings := map[nat.Port][]nat.PortBinding{
		nat.Port(fmt.Sprintf("%s/tcp", n.exposedPorts.containerPort)): {
			{
				HostPort: n.exposedPorts.hostPort,
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
