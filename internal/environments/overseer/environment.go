package overseer

import (
	"fmt"
	"sync"

	"github.com/vechain/networkhub/internal/environments"
	"github.com/vechain/networkhub/internal/environments/docker"
	"github.com/vechain/networkhub/internal/environments/local"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
)

// Overseer centrally orchestrates all node management across environments
type Overseer struct {
	// Central state management
	networkCfg *network.Network
	nodes      map[string]node.Lifecycle
	started    bool

	// Infrastructure utilities
	dockerManager *docker.Manager
	localManager  *local.Manager

	mu sync.Mutex
}

// New creates a new overseer instance with the given network configuration
func New(cfg *network.Network) (*Overseer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("network configuration cannot be nil")
	}

	overseer := &Overseer{
		networkCfg: cfg,
		nodes:      make(map[string]node.Lifecycle),
		started:    false,
	}

	// Initialize the appropriate managers based on environment type
	switch cfg.Environment {
	case environments.Local:
		overseer.localManager = local.NewManager()
	case environments.Docker:
		overseer.dockerManager = docker.NewManager()
		if err := overseer.dockerManager.Initialize(cfg); err != nil {
			return nil, fmt.Errorf("failed to initialize Docker manager: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported environment: %s", cfg.Environment)
	}

	return overseer, nil
}

// StartNetwork starts all nodes in the network
func (o *Overseer) StartNetwork() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.started {
		return fmt.Errorf("network is already running")
	}

	if len(o.networkCfg.Nodes) == 0 {
		// For public networks, it's valid to start with 0 nodes and add them later
		if !o.networkCfg.IsPublicNetwork() {
			return fmt.Errorf("no nodes defined in the network")
		}
		// Mark as started but don't actually start any nodes yet
		o.started = true
		return nil
	}

	// Build thor binary if needed
	if err := o.buildThorBinaryIfNeeded(); err != nil {
		return fmt.Errorf("failed to build thor binary: %w", err)
	}

	// Generate enodes for faster p2p bootstrap
	enodes, err := o.generateEnodes()
	if err != nil {
		return fmt.Errorf("failed to generate enodes: %w", err)
	}

	// Start all nodes
	for _, nodeCfg := range o.networkCfg.Nodes {
		// Validate node before starting
		if err := o.validateNode(nodeCfg); err != nil {
			return fmt.Errorf("failed to validate node %s: %w", nodeCfg.GetID(), err)
		}

		// Start the node using the appropriate manager
		var nodeInstance node.Lifecycle
		switch o.networkCfg.Environment {
		case environments.Local:
			nodeInstance, err = o.localManager.StartNode(nodeCfg, o.networkCfg, enodes)
		case environments.Docker:
			nodeInstance, err = o.dockerManager.StartNode(nodeCfg, o.networkCfg, enodes)
		default:
			return fmt.Errorf("unsupported environment: %s", o.networkCfg.Environment)
		}

		if err != nil {
			return fmt.Errorf("unable to start node %s: %w", nodeCfg.GetID(), err)
		}

		o.nodes[nodeCfg.GetID()] = nodeInstance
	}

	o.started = true
	return nil
}

// StopNetwork stops all nodes in the network
func (o *Overseer) StopNetwork() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if !o.started {
		return nil // Already stopped
	}

	var lastErr error
	for nodeID, nodeInstance := range o.nodes {
		if err := o.stopNode(nodeInstance); err != nil {
			lastErr = fmt.Errorf("failed to stop node %s: %w", nodeID, err)
		}
	}

	o.nodes = make(map[string]node.Lifecycle)
	o.started = false

	// Clean up Docker resources if using Docker environment
	if o.networkCfg.Environment == environments.Docker && o.dockerManager != nil {
		if err := o.dockerManager.Cleanup(); err != nil {
			if lastErr == nil {
				lastErr = fmt.Errorf("failed to cleanup Docker resources: %w", err)
			}
		}
	}

	return lastErr
}

// AddNode adds a node to the existing network
func (o *Overseer) AddNode(nodeConfig node.Config) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.networkCfg == nil {
		return fmt.Errorf("network configuration is not loaded")
	}

	if _, exists := o.nodes[nodeConfig.GetID()]; exists {
		return fmt.Errorf("node with ID %s already exists", nodeConfig.GetID())
	}

	// Add to network configuration
	o.networkCfg.Nodes = append(o.networkCfg.Nodes, nodeConfig)

	// If network is running, start the new node immediately
	if o.started {
		// Build Thor binary if needed
		if err := o.buildThorBinaryIfNeeded(); err != nil {
			return fmt.Errorf("failed to build thor binary for new node: %w", err)
		}

		// Validate the node before starting
		if err := o.validateNode(nodeConfig); err != nil {
			return fmt.Errorf("failed to validate node %s: %w", nodeConfig.GetID(), err)
		}

		// Generate fresh enodes including the new node
		enodes, err := o.generateEnodes()
		if err != nil {
			return fmt.Errorf("failed to generate enodes: %w", err)
		}

		// Start the node using the appropriate manager
		var nodeInstance node.Lifecycle
		switch o.networkCfg.Environment {
		case environments.Local:
			nodeInstance, err = o.localManager.StartNode(nodeConfig, o.networkCfg, enodes)
		case environments.Docker:
			nodeInstance, err = o.dockerManager.StartNode(nodeConfig, o.networkCfg, enodes)
		default:
			return fmt.Errorf("unsupported environment: %s", o.networkCfg.Environment)
		}

		if err != nil {
			return fmt.Errorf("unable to start node %s after adding: %w", nodeConfig.GetID(), err)
		}

		o.nodes[nodeConfig.GetID()] = nodeInstance
	}

	return nil
}

// RemoveNode removes a node from the network
func (o *Overseer) RemoveNode(nodeID string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	nodeInstance, exists := o.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node with ID %s does not exist", nodeID)
	}

	// Find and remove from network configuration
	index := -1
	for i, n := range o.networkCfg.Nodes {
		if n.GetID() == nodeID {
			index = i
			break
		}
	}
	if index == -1 {
		return fmt.Errorf("node with ID %s not found in network configuration", nodeID)
	}

	// Stop the node
	if err := o.stopNode(nodeInstance); err != nil {
		return fmt.Errorf("unable to stop node %s: %w", nodeID, err)
	}

	// Remove from configuration and tracking
	o.networkCfg.Nodes = append(o.networkCfg.Nodes[:index], o.networkCfg.Nodes[index+1:]...)
	delete(o.nodes, nodeID)

	return nil
}

// Nodes returns all nodes in the environment
func (o *Overseer) Nodes() map[string]node.Lifecycle {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Return a copy to prevent external modification
	nodes := make(map[string]node.Lifecycle, len(o.nodes))
	for id, node := range o.nodes {
		nodes[id] = node
	}
	return nodes
}

// Config returns the network configuration
func (o *Overseer) Config() *network.Network {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.networkCfg
}

// Helper methods for orchestration

// buildThorBinaryIfNeeded builds the thor binary if needed and sets exec artifact for nodes
func (o *Overseer) buildThorBinaryIfNeeded() error {
	var execPath string
	var err error

	switch o.networkCfg.Environment {
	case environments.Local:
		execPath, err = o.localManager.BuildThorBinary(o.networkCfg.ThorBuilder)
	case environments.Docker:
		execPath, err = o.dockerManager.BuildThorBinary(o.networkCfg.ThorBuilder)
	default:
		return fmt.Errorf("unsupported environment: %s", o.networkCfg.Environment)
	}

	if err != nil {
		return err
	}

	// Set exec artifact for nodes that don't have one configured
	if execPath != "" {
		for _, nodeConfig := range o.networkCfg.Nodes {
			if nodeConfig.GetExecArtifact() == "" {
				nodeConfig.SetExecArtifact(execPath)
			}
		}
	}

	return nil
}

// validateNode validates a node configuration
func (o *Overseer) validateNode(nodeCfg node.Config) error {
	switch o.networkCfg.Environment {
	case environments.Local:
		return o.localManager.ValidateNode(nodeCfg)
	case environments.Docker:
		return o.dockerManager.ValidateNode(nodeCfg)
	default:
		return fmt.Errorf("unsupported environment: %s", o.networkCfg.Environment)
	}
}

// generateEnodes generates enode strings for all nodes
func (o *Overseer) generateEnodes() ([]string, error) {
	switch o.networkCfg.Environment {
	case environments.Local:
		return o.localManager.GenerateEnodes(o.networkCfg)
	case environments.Docker:
		return o.dockerManager.GenerateEnodes(o.networkCfg)
	default:
		return nil, fmt.Errorf("unsupported environment: %s", o.networkCfg.Environment)
	}
}

// stopNode stops a node instance
func (o *Overseer) stopNode(nodeInstance node.Lifecycle) error {
	switch o.networkCfg.Environment {
	case environments.Local:
		return o.localManager.StopNode(nodeInstance)
	case environments.Docker:
		return o.dockerManager.StopNode(nodeInstance)
	default:
		return fmt.Errorf("unsupported environment: %s", o.networkCfg.Environment)
	}
}
