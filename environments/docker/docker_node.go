package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
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
func NewDockerNode(
	cfg node.Config,
	enodes []string,
	networkID string,
	hostPort string,
	containerPort string,
	ipAddr string,
) *Node {
	return &Node{
		cfg:           cfg,
		enodes:        enodes,
		networkID:     networkID,
		hostPort:      hostPort,
		containerPort: containerPort,
		ipAddr:        ipAddr,
	}
}

// Node represents a Docker container node
type Node struct {
	cfg           node.Config
	enodes        []string
	id            string
	networkID     string
	hostPort      string
	containerPort string
	ipAddr        string
	hostDataDir   string
	logConfig     container.LogConfig
}

// SetHostVolume sets the base directory on the host where node data will be stored.
// The system will automatically create subdirectories: {hostDataDir}/{nodeID}/config and {hostDataDir}/{nodeID}/data
func (n *Node) SetHostVolume(hostDataDir string) {
	n.hostDataDir = hostDataDir
}

func (n *Node) SetLogConfig(logConfig container.LogConfig) {
	n.logConfig = logConfig
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

	var cleanEnode []string // todo theres a clever way of doing this
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

	args := fmt.Sprintf("cd %s; ", n.cfg.GetConfigDir()) +
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
		fmt.Sprintf("--bootnode=%s", enodeString)

	for key, value := range n.cfg.GetAdditionalArgs() {
		args += fmt.Sprintf(" --%s=%s", key, value)
	}

	cmd := []string{"sh", "-c", args}

	//serialize genesis
	genesisBytes, err := nodegenesis.Marshal(n.cfg.GetGenesis())
	if err != nil {
		return fmt.Errorf("unable to marshal genesis - %w", err)
	}

	exposedPorts := nat.PortSet{
		nat.Port(fmt.Sprintf("%s/tcp", n.containerPort)): struct{}{},
	}
	portBindings := map[nat.Port][]nat.PortBinding{
		nat.Port(fmt.Sprintf("%s/tcp", n.containerPort)): {
			{
				HostPort: n.hostPort,
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
		LogConfig:    n.logConfig,
	}

	// Set up automatic volume mounts if hostDataDir is specified
	if n.hostDataDir != "" {
		config.Volumes = make(map[string]struct{})
		var binds []string

		nodeID := n.cfg.GetID()

		// Mount config directory: eg /host/nodes/{nodeID}/config -> container's config dir
		if configDir := n.cfg.GetConfigDir(); configDir != "" {
			hostConfigPath := fmt.Sprintf("%s/%s/config", n.hostDataDir, nodeID)

			// Create the host directory if it doesn't exist
			if err := os.MkdirAll(hostConfigPath, 0755); err != nil {
				return fmt.Errorf("failed to create config directory %s: %w", hostConfigPath, err)
			}

			volumeBind := fmt.Sprintf("%s:%s", hostConfigPath, configDir)
			config.Volumes[configDir] = struct{}{}
			binds = append(binds, volumeBind)
		}

		// Mount data directory: eg /host/nodes/{nodeID}/data -> container's data dir
		if dataDir := n.cfg.GetDataDir(); dataDir != "" {
			hostDataPath := fmt.Sprintf("%s/%s/data", n.hostDataDir, nodeID)

			// Create the host directory if it doesn't exist
			if err := os.MkdirAll(hostDataPath, 0755); err != nil {
				return fmt.Errorf("failed to create data directory %s: %w", hostDataPath, err)
			}

			volumeBind := fmt.Sprintf("%s:%s", hostDataPath, dataDir)
			config.Volumes[dataDir] = struct{}{}
			binds = append(binds, volumeBind)
		}

		slog.Info("setting up volume binds for node", "nodeID", nodeID, "binds", strings.Join(binds, ", "))

		hostConfig.Binds = binds
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
