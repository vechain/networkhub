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
	"github.com/docker/docker/api/types/strslice"
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

	if n.hostDataDir == "" {
		n.hostDataDir = os.TempDir()
		slog.Info("host data directory not specified, using temporary directory", "dir", n.hostDataDir)
	}
	baseHostDir := fmt.Sprintf("%s/%s", n.hostDataDir, n.cfg.GetID())

	// write the genesis to the config directory on the host
	genesisBytes, err := nodegenesis.Marshal(n.cfg.GetGenesis())
	if err != nil {
		return fmt.Errorf("unable to marshal genesis - %w", err)
	}
	configDir := fmt.Sprintf("%s/config", baseHostDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}
	genesisPath := fmt.Sprintf("%s/genesis.json", configDir)
	if err := os.WriteFile(genesisPath, genesisBytes, 0644); err != nil {
		return fmt.Errorf("failed to write genesis file to %s: %w", genesisPath, err)
	}

	// write the master key to the config directory on the host
	masterKeyPath := fmt.Sprintf("%s/master.key", configDir)
	if err := os.WriteFile(masterKeyPath, []byte(n.cfg.GetKey()), 0600); err != nil {
		return fmt.Errorf("failed to write master key to %s: %w", masterKeyPath, err)
	}
	// write the p2p key to the config directory on the host
	p2pKeyPath := fmt.Sprintf("%s/p2p.key", configDir)
	if err := os.WriteFile(p2pKeyPath, []byte(n.cfg.GetKey()), 0600); err != nil {
		return fmt.Errorf("failed to write p2p key to %s: %w", p2pKeyPath, err)
	}

	containerConfigDir := n.cfg.GetConfigDir()
	if strings.HasSuffix(containerConfigDir, "/") {
		containerConfigDir = strings.TrimSuffix(containerConfigDir, "/")
	}

	var cmd strslice.StrSlice
	cmd = append(cmd, "--network", containerConfigDir+"/genesis.json")
	cmd = append(cmd, "--nat", "none")
	cmd = append(cmd, "--config-dir", n.cfg.GetConfigDir())
	cmd = append(cmd, "--api-addr", n.cfg.GetAPIAddr())
	cmd = append(cmd, "--api-cors", n.cfg.GetAPICORS())
	cmd = append(cmd, "--p2p-port", fmt.Sprintf("%d", n.cfg.GetP2PListenPort()))

	if enodeString != "" {
		cmd = append(cmd, "--bootnodes", enodeString)
	}

	for key, value := range n.cfg.GetAdditionalArgs() {
		cmd = append(cmd, "--"+key, value)
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
		Image: n.cfg.GetExecArtifact(),
		Cmd:   cmd,
		Entrypoint: []string{
			"thor",
		},
		ExposedPorts: exposedPorts,
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		LogConfig:    n.logConfig,
	}

	config.Volumes = make(map[string]struct{})
	var binds []string

	nodeID := n.cfg.GetID()

	// Mount config directory: eg /host/nodes/{nodeID}/config -> container's config dir
	if configDir := n.cfg.GetConfigDir(); configDir != "" {
		hostConfigPath := fmt.Sprintf("%s/config", baseHostDir)

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
		hostDataPath := fmt.Sprintf("%s/data", baseHostDir)

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
