package launcher

import (
	"fmt"
	"sync"

	"github.com/vechain/networkhub/internal/environments"
	"github.com/vechain/networkhub/internal/environments/docker"
	"github.com/vechain/networkhub/internal/environments/local"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
)

// Launcher centrally orchestrates all node management across environments
type Launcher struct {
	// Central state management
	networkCfg *network.Network
	nodes      map[string]node.Lifecycle
	started    bool

	// Infrastructure utilities
	dockerManager *docker.Manager
	localManager  *local.Manager

	mu sync.Mutex
}

// New creates a new launcher instance with the given network configuration
func New(cfg *network.Network) (*Launcher, error) {
	if cfg == nil {
		return nil, fmt.Errorf("network configuration cannot be nil")
	}

	launcher := &Launcher{
		networkCfg: cfg,
		nodes:      make(map[string]node.Lifecycle),
		started:    false,
	}

	// Initialize the appropriate managers based on environment type
	switch cfg.Environment {
	case environments.Local:
		launcher.localManager = local.NewManager()
	case environments.Docker:
		launcher.dockerManager = docker.NewManager()
		if err := launcher.dockerManager.Initialize(cfg); err != nil {
			return nil, fmt.Errorf("failed to initialize Docker manager: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported environment: %s", cfg.Environment)
	}

	return launcher, nil
}

// StartNetwork starts all nodes in the network
func (l *Launcher) StartNetwork() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.started {
		return fmt.Errorf("network is already running")
	}

	if len(l.networkCfg.Nodes) == 0 {
		// For public networks, it's valid to start with 0 nodes and add them later
		if !l.networkCfg.IsPublicNetwork() {
			return fmt.Errorf("no nodes defined in the network")
		}
		// Mark as started but don't actually start any nodes yet
		l.started = true
		return nil
	}

	// Build thor binary if needed
	if err := l.buildThorBinaryIfNeeded(); err != nil {
		return fmt.Errorf("failed to build thor binary: %w", err)
	}

	// Generate enodes for faster p2p bootstrap
	enodes, err := l.generateEnodes()
	if err != nil {
		return fmt.Errorf("failed to generate enodes: %w", err)
	}

	// Start all nodes
	for _, nodeCfg := range l.networkCfg.Nodes {
		// Validate node before starting
		if err := l.validateNode(nodeCfg); err != nil {
			return fmt.Errorf("failed to validate node %s: %w", nodeCfg.GetID(), err)
		}

		// Start the node using the appropriate manager
		var nodeInstance node.Lifecycle
		switch l.networkCfg.Environment {
		case environments.Local:
			nodeInstance, err = l.localManager.StartNode(nodeCfg, l.networkCfg, enodes)
		case environments.Docker:
			nodeInstance, err = l.dockerManager.StartNode(nodeCfg, l.networkCfg, enodes)
		default:
			return fmt.Errorf("unsupported environment: %s", l.networkCfg.Environment)
		}

		if err != nil {
			return fmt.Errorf("unable to start node %s: %w", nodeCfg.GetID(), err)
		}

		l.nodes[nodeCfg.GetID()] = nodeInstance
	}

	l.started = true
	return nil
}

// StopNetwork stops all nodes in the network
func (l *Launcher) StopNetwork() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.started {
		return nil // Already stopped
	}

	var lastErr error
	for nodeID, nodeInstance := range l.nodes {
		if err := l.stopNode(nodeInstance); err != nil {
			lastErr = fmt.Errorf("failed to stop node %s: %w", nodeID, err)
		}
	}

	l.nodes = make(map[string]node.Lifecycle)
	l.started = false

	// Clean up Docker resources if using Docker environment
	if l.networkCfg.Environment == environments.Docker && l.dockerManager != nil {
		if err := l.dockerManager.Cleanup(); err != nil {
			if lastErr == nil {
				lastErr = fmt.Errorf("failed to cleanup Docker resources: %w", err)
			}
		}
	}

	return lastErr
}

// AddNode adds a node to the existing network
func (l *Launcher) AddNode(nodeConfig node.Config) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.networkCfg == nil {
		return fmt.Errorf("network configuration is not loaded")
	}

	if _, exists := l.nodes[nodeConfig.GetID()]; exists {
		return fmt.Errorf("node with ID %s already exists", nodeConfig.GetID())
	}

	// Add to network configuration
	l.networkCfg.Nodes = append(l.networkCfg.Nodes, nodeConfig)

	// If network is running, start the new node immediately
	if l.started {
		// Build Thor binary if needed
		if err := l.buildThorBinaryIfNeeded(); err != nil {
			return fmt.Errorf("failed to build thor binary for new node: %w", err)
		}

		// Validate the node before starting
		if err := l.validateNode(nodeConfig); err != nil {
			return fmt.Errorf("failed to validate node %s: %w", nodeConfig.GetID(), err)
		}

		// Generate fresh enodes including the new node
		enodes, err := l.generateEnodes()
		if err != nil {
			return fmt.Errorf("failed to generate enodes: %w", err)
		}

		// Start the node using the appropriate manager
		var nodeInstance node.Lifecycle
		switch l.networkCfg.Environment {
		case environments.Local:
			nodeInstance, err = l.localManager.StartNode(nodeConfig, l.networkCfg, enodes)
		case environments.Docker:
			nodeInstance, err = l.dockerManager.StartNode(nodeConfig, l.networkCfg, enodes)
		default:
			return fmt.Errorf("unsupported environment: %s", l.networkCfg.Environment)
		}

		if err != nil {
			return fmt.Errorf("unable to start node %s after adding: %w", nodeConfig.GetID(), err)
		}

		l.nodes[nodeConfig.GetID()] = nodeInstance
	}

	return nil
}

// RemoveNode removes a node from the network
func (l *Launcher) RemoveNode(nodeID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	nodeInstance, exists := l.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node with ID %s does not exist", nodeID)
	}

	// Find and remove from network configuration
	index := -1
	for i, n := range l.networkCfg.Nodes {
		if n.GetID() == nodeID {
			index = i
			break
		}
	}
	if index == -1 {
		return fmt.Errorf("node with ID %s not found in network configuration", nodeID)
	}

	// Stop the node
	if err := l.stopNode(nodeInstance); err != nil {
		return fmt.Errorf("unable to stop node %s: %w", nodeID, err)
	}

	// Remove from configuration and tracking
	l.networkCfg.Nodes = append(l.networkCfg.Nodes[:index], l.networkCfg.Nodes[index+1:]...)
	delete(l.nodes, nodeID)

	return nil
}

// Nodes returns all nodes in the environment
func (l *Launcher) Nodes() map[string]node.Lifecycle {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Return a copy to prevent external modification
	nodes := make(map[string]node.Lifecycle, len(l.nodes))
	for id, node := range l.nodes {
		nodes[id] = node
	}
	return nodes
}

// Config returns the network configuration
func (l *Launcher) Config() *network.Network {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.networkCfg
}

// Helper methods for orchestration

// buildThorBinaryIfNeeded builds the thor binary if needed and sets exec artifact for nodes
func (l *Launcher) buildThorBinaryIfNeeded() error {
	var execPath string
	var err error

	switch l.networkCfg.Environment {
	case environments.Local:
		execPath, err = l.localManager.BuildThorBinary(l.networkCfg.ThorBuilder)
	case environments.Docker:
		execPath, err = l.dockerManager.BuildThorBinary(l.networkCfg.ThorBuilder)
	default:
		return fmt.Errorf("unsupported environment: %s", l.networkCfg.Environment)
	}

	if err != nil {
		return err
	}

	// Set exec artifact for nodes that don't have one configured
	if execPath != "" {
		for _, nodeConfig := range l.networkCfg.Nodes {
			if nodeConfig.GetExecArtifact() == "" {
				nodeConfig.SetExecArtifact(execPath)
			}
		}
	}

	return nil
}

// validateNode validates a node configuration
func (l *Launcher) validateNode(nodeCfg node.Config) error {
	switch l.networkCfg.Environment {
	case environments.Local:
		return l.localManager.ValidateNode(nodeCfg)
	case environments.Docker:
		return l.dockerManager.ValidateNode(nodeCfg)
	default:
		return fmt.Errorf("unsupported environment: %s", l.networkCfg.Environment)
	}
}

// generateEnodes generates enode strings for all nodes
func (l *Launcher) generateEnodes() ([]string, error) {
	switch l.networkCfg.Environment {
	case environments.Local:
		return l.localManager.GenerateEnodes(l.networkCfg)
	case environments.Docker:
		return l.dockerManager.GenerateEnodes(l.networkCfg)
	default:
		return nil, fmt.Errorf("unsupported environment: %s", l.networkCfg.Environment)
	}
}

// stopNode stops a node instance
func (l *Launcher) stopNode(nodeInstance node.Lifecycle) error {
	switch l.networkCfg.Environment {
	case environments.Local:
		return l.localManager.StopNode(nodeInstance)
	case environments.Docker:
		return l.dockerManager.StopNode(nodeInstance)
	default:
		return fmt.Errorf("unsupported environment: %s", l.networkCfg.Environment)
	}
}
