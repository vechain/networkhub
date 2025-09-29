package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	dockertypes "github.com/docker/docker/api/types"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/thorbuilder"
)

// Manager handles Docker container management utilities
type Manager struct {
	ipManager   *IpManager
	dockerImage string
	networkName string
	mu          sync.Mutex
}

// NewManager creates a new Docker container manager
func NewManager() *Manager {
	return &Manager{
		ipManager: NewIPManagerRandom(),
	}
}

// Initialize sets up the Docker environment with network configuration
func (m *Manager) Initialize(networkCfg *network.Network) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Set network name and docker image
	m.networkName = fmt.Sprintf("docker%s-network", networkCfg.BaseID)

	// Use default docker image if not specified
	m.dockerImage = "vechain/thor:latest"

	// Create Docker network
	if err := m.createNetwork(); err != nil {
		return fmt.Errorf("failed to create Docker network: %w", err)
	}

	return nil
}

// StartNode starts a Docker container node
func (m *Manager) StartNode(nodeCfg node.Config, networkCfg *network.Network, enodes []string) (node.Lifecycle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get IP for the node (should already be allocated during enode generation)
	ipAddr := m.ipManager.GetNodeIP(nodeCfg.GetID())
	if ipAddr == "" {
		// Fallback: allocate IP if not already assigned
		var err error
		ipAddr, err = m.ipManager.NextIP(nodeCfg.GetID())
		if err != nil {
			return nil, fmt.Errorf("failed to allocate IP for node %s: %w", nodeCfg.GetID(), err)
		}
	}

	// Create exposed port configuration from node's API address
	exposedPort := m.getExposedPort(nodeCfg)
	if exposedPort == nil {
		return nil, fmt.Errorf("unable to determine API port for node %s", nodeCfg.GetID())
	}

	// Create and return the docker node
	dockerNode := NewDockerNode(nodeCfg, enodes, m.networkName, exposedPort, ipAddr)
	if err := dockerNode.Start(); err != nil {
		return nil, fmt.Errorf("failed to start docker node %s: %w", nodeCfg.GetID(), err)
	}

	return dockerNode, nil
}

// StopNode stops a Docker container node
func (m *Manager) StopNode(nodeInstance node.Lifecycle) error {
	return nodeInstance.Stop()
}

// GenerateEnodes creates enode strings for all nodes using Docker IPs
func (m *Manager) GenerateEnodes(networkCfg *network.Network) ([]string, error) {
	var enodes []string

	// Skip enode generation entirely for public networks (testnet/mainnet)
	if networkCfg.IsPublicNetwork() {
		return enodes, nil
	}

	for _, node := range networkCfg.Nodes {
		// Pre-allocate IP address if not already assigned
		ipAddr := m.ipManager.GetNodeIP(node.GetID())
		if ipAddr == "" {
			// Assign IP address now for enode generation
			var err error
			ipAddr, err = m.ipManager.NextIP(node.GetID())
			if err != nil {
				return nil, fmt.Errorf("failed to allocate IP for node %s: %w", node.GetID(), err)
			}
		}

		enode, err := node.Enode(ipAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to generate enode for node %s: %w", node.GetID(), err)
		}
		enodes = append(enodes, enode)
	}

	return enodes, nil
}

// BuildThorBinary builds the thor binary (Docker version) and returns the image name
func (m *Manager) BuildThorBinary(thorBuilder *thorbuilder.Config) (string, error) {
	if thorBuilder == nil {
		return "", nil // No thor builder configuration
	}

	builder := thorbuilder.New(thorBuilder)
	// For Docker, we need to build a Docker image instead of binary
	dockerImage, err := builder.BuildDockerImage()
	if err != nil {
		return "", fmt.Errorf("failed to build thor docker image: %w", err)
	}

	return dockerImage, nil
}

// ValidateNode validates a node configuration for Docker
func (m *Manager) ValidateNode(nodeCfg node.Config) error {
	if nodeCfg.GetExecArtifact() == "" {
		return fmt.Errorf("docker image cannot be empty")
	}

	// use preset dirs if not defined
	if nodeCfg.GetConfigDir() == "" {
		nodeCfg.SetConfigDir("/home/thor")
	}
	if nodeCfg.GetDataDir() == "" {
		nodeCfg.SetDataDir("/home/thor")
	}

	// ensure API ports are exposed to the localhost
	//split := strings.Split(nodeCfg.GetAPIAddr(), ":")
	//if len(split) != 2 {
	//	return fmt.Errorf("unable to parse API Addr")
	//}
	//
	//exposedAPIPort, err := strconv.Atoi(split[1])
	//if err != nil {
	//	return err
	//}

	//d.exposedPorts[node.GetID()] = &exposedPort{
	//	hostPort:      fmt.Sprintf("%d", exposedAPIPort+i),
	//	containerPort: split[1],
	//}
	return nil
}

// createNetwork creates a Docker network for the nodes
func (m *Manager) createNetwork() error {
	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("could not create Docker client: %v", err)
	}
	defer cli.Close()

	// List existing networks
	networks, err := cli.NetworkList(context.Background(), dockertypes.NetworkListOptions{})
	if err != nil {
		return fmt.Errorf("could not list Docker networks: %v", err)
	}

	// Check if the network already exists
	for _, net := range networks {
		if net.Name == m.networkName {
			slog.Info("Network already exists", "networkName", m.networkName)
			err := cli.NetworkRemove(context.Background(), m.networkName)
			if err != nil {
				return err
			}
			slog.Info("Removed existing network", "networkName", m.networkName)
		}
	}

	// Network does not exist, create it
	// Define the network configuration
	networkCreate := dockertypes.NetworkCreate{
		Driver: "bridge",
		IPAM: &dockernetwork.IPAM{
			Driver: "default",
			Config: []dockernetwork.IPAMConfig{
				{
					Subnet: m.ipManager.Subnet(),
				},
			},
		},
	}
	_, err = cli.NetworkCreate(context.Background(), m.networkName, networkCreate)
	if err != nil {
		return fmt.Errorf("could not create Docker network: %v", err)
	}

	slog.Info("Network created", "networkName", m.networkName, "subnet", m.ipManager.Subnet())
	return nil
}

// Cleanup removes Docker resources
func (m *Manager) Cleanup() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a Docker client to clean up network
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("could not create Docker client for cleanup: %v", err)
	}
	defer cli.Close()

	// Remove the Docker network if it exists
	if m.networkName != "" {
		err := cli.NetworkRemove(context.Background(), m.networkName)
		if err != nil {
			slog.Warn("Failed to remove Docker network", "networkName", m.networkName, "error", err)
			// Don't return error as cleanup should be best-effort
		} else {
			slog.Info("Cleaned up Docker network", "networkName", m.networkName)
		}
	}

	return nil
}

// getExposedPort extracts port configuration from node API address
func (m *Manager) getExposedPort(nodeCfg node.Config) *ExposedPort {
	apiAddr := nodeCfg.GetAPIAddr()
	if apiAddr == "" {
		return nil
	}

	// Split address into host:port
	parts := strings.Split(apiAddr, ":")
	if len(parts) != 2 {
		return nil
	}

	port := parts[1]
	return &ExposedPort{
		HostPort:      port,
		ContainerPort: port,
	}
}
