package local

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/thorbuilder"
)

// Environment implements overseer.Environment for local environments
type Environment struct {
	nodes      map[string]node.Lifecycle
	networkCfg *network.Network
	started    bool
	mu         sync.Mutex
}

// NewEnvironment creates a new local environment
func NewEnvironment(cfg *network.Network) *Environment {
	return &Environment{
		nodes:      make(map[string]node.Lifecycle),
		networkCfg: cfg,
	}
}

// StartNetwork starts all nodes in the network
func (e *Environment) StartNetwork() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.started {
		return fmt.Errorf("network already started")
	}

	// Handle thor binary building once for all nodes
	if err := e.buildThorBinaryIfNeeded(); err != nil {
		return fmt.Errorf("failed to build thor binary: %w", err)
	}

	// Generate enodes for faster p2p bootstrap
	enodes, err := e.generateEnodes()
	if err != nil {
		return fmt.Errorf("failed to generate enodes: %w", err)
	}

	for _, nodeCfg := range e.networkCfg.Nodes {
		// Validate node before starting
		if err := e.checkNode(nodeCfg); err != nil {
			return fmt.Errorf("failed to validate node %s: %w", nodeCfg.GetID(), err)
		}

		nodeInstance := NewLocalNode(nodeCfg, e.networkCfg, enodes)

		if err := nodeInstance.Start(); err != nil {
			return fmt.Errorf("unable to start node %s: %w", nodeCfg.GetID(), err)
		}

		e.nodes[nodeCfg.GetID()] = nodeInstance
	}

	e.started = true
	return nil
}

// StopNetwork stops all nodes in the network
func (e *Environment) StopNetwork() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.started {
		return nil
	}

	for nodeID, nodeInstance := range e.nodes {
		if err := nodeInstance.Stop(); err != nil {
			return fmt.Errorf("unable to stop node %s: %w", nodeID, err)
		}
	}

	e.nodes = make(map[string]node.Lifecycle)
	e.started = false
	return nil
}

// AddNode adds a node to the existing network
func (e *Environment) AddNode(nodeConfig node.Config) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.networkCfg == nil {
		return fmt.Errorf("network configuration is not loaded")
	}

	if _, exists := e.nodes[nodeConfig.GetID()]; exists {
		return fmt.Errorf("node with ID %s already exists", nodeConfig.GetID())
	}

	e.networkCfg.Nodes = append(e.networkCfg.Nodes, nodeConfig)

	// If network is running, start the new node immediately
	if e.started {
		// Build Thor binary if needed for this node
		if err := e.buildThorBinaryIfNeeded(); err != nil {
			return fmt.Errorf("failed to build thor binary for new node: %w", err)
		}

		// Validate the node before starting
		if err := e.checkNode(nodeConfig); err != nil {
			return fmt.Errorf("failed to validate node %s: %w", nodeConfig.GetID(), err)
		}

		enodes, err := e.generateEnodes()
		if err != nil {
			return fmt.Errorf("failed to generate enodes: %w", err)
		}

		nodeInstance := NewLocalNode(nodeConfig, e.networkCfg, enodes)
		if err := nodeInstance.Start(); err != nil {
			return fmt.Errorf("unable to start node %s after adding: %w", nodeConfig.GetID(), err)
		}

		e.nodes[nodeConfig.GetID()] = nodeInstance
	}

	return nil
}

// RemoveNode removes a node from the network
func (e *Environment) RemoveNode(nodeID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	nodeInstance, exists := e.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node with ID %s does not exist", nodeID)
	}

	// Find and remove from network configuration
	index := -1
	for i, n := range e.networkCfg.Nodes {
		if n.GetID() == nodeID {
			index = i
			break
		}
	}
	if index == -1 {
		return fmt.Errorf("node with ID %s not found in network configuration", nodeID)
	}

	if err := nodeInstance.Stop(); err != nil {
		return fmt.Errorf("unable to stop node %s: %w", nodeID, err)
	}

	e.networkCfg.Nodes = append(e.networkCfg.Nodes[:index], e.networkCfg.Nodes[index+1:]...)
	delete(e.nodes, nodeID)

	return nil
}

// Nodes returns all nodes in the environment
func (e *Environment) Nodes() map[string]node.Lifecycle {
	e.mu.Lock()
	defer e.mu.Unlock()

	nodes := make(map[string]node.Lifecycle, len(e.nodes))
	for k, v := range e.nodes {
		nodes[k] = v
	}
	return nodes
}

// Config returns the network configuration
func (e *Environment) Config() *network.Network {
	return e.networkCfg
}

// checkNode validates and sets defaults for a node configuration
func (e *Environment) checkNode(n node.Config) error {
	// check if the exec artifact path exists
	if !fileExists(n.GetExecArtifact()) {
		return fmt.Errorf("exec artifact path %s does not exist for node %s", n.GetExecArtifact(), n.GetID())
	}

	if n.GetConfigDir() == "" {
		n.SetConfigDir(filepath.Join(filepath.Dir(n.GetExecArtifact()), n.GetID(), "config"))
	}

	if n.GetDataDir() == "" {
		n.SetDataDir(filepath.Join(filepath.Dir(n.GetExecArtifact()), n.GetID(), "data"))
	}
	return nil
}

// buildThorBinaryIfNeeded builds the thor binary once and sets exec artifact for all nodes
func (e *Environment) buildThorBinaryIfNeeded() error {
	if e.networkCfg.ThorBuilder == nil {
		return nil // No thor builder configured
	}

	execPath, err := thorbuilder.NewAndBuild(e.networkCfg.ThorBuilder)
	if err != nil {
		return fmt.Errorf("failed to build thor binary: %w", err)
	}

	// Set exec artifact for all nodes that don't have one
	for _, nodeConfig := range e.networkCfg.Nodes {
		if nodeConfig.GetExecArtifact() == "" {
			nodeConfig.SetExecArtifact(execPath)
		}
	}

	return nil
}

// generateEnodes creates enode strings for all nodes (excluding public network nodes)
func (e *Environment) generateEnodes() ([]string, error) {
	var enodes []string
	// Skip enode generation entirely for public networks (testnet/mainnet)
	if e.networkCfg.IsPublicNetwork() {
		return enodes, nil
	}
	
	for _, node := range e.networkCfg.Nodes {

		// Use localhost for local environment
		enode, err := node.Enode("127.0.0.1")
		if err != nil {
			return nil, fmt.Errorf("failed to get enode for node %s: %w", node.GetID(), err)
		}
		enodes = append(enodes, enode)
	}
	return enodes, nil
}

// Helper functions

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

