package docker

import (
	"fmt"
	"sync"

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

	// Allocate IP for the node
	ipAddr, err := m.ipManager.NextIP(nodeCfg.GetID())
	if err != nil {
		return nil, fmt.Errorf("failed to allocate IP for node %s: %w", nodeCfg.GetID(), err)
	}

	// Create exposed port configuration
	exposedPort := &ExposedPort{
		HostPort:      "8545", // Default API port
		ContainerPort: "8545",
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
	
	// For Docker, the exec artifact should be a Docker image name
	// Additional validation could be added here
	return nil
}

// createNetwork creates a Docker network for the nodes
func (m *Manager) createNetwork() error {
	// Implementation would use Docker API to create network
	// For now, this is a placeholder that would contain the logic
	// from the current docker environment's network creation
	return nil
}

// Cleanup removes Docker resources
func (m *Manager) Cleanup() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clean up Docker network and resources
	return nil
}