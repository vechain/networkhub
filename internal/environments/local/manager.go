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

// Manager handles local process management utilities
type Manager struct {
	mu sync.Mutex
}

// NewManager creates a new local process manager
func NewManager() *Manager {
	return &Manager{}
}

// StartNode starts a local node process
func (m *Manager) StartNode(nodeCfg node.Config, networkCfg *network.Network, enodes []string) (node.Lifecycle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	nodeInstance := NewLocalNode(nodeCfg, networkCfg, enodes)
	if err := nodeInstance.Start(); err != nil {
		return nil, fmt.Errorf("failed to start local node %s: %w", nodeCfg.GetID(), err)
	}

	return nodeInstance, nil
}

// StopNode stops a local node process
func (m *Manager) StopNode(nodeInstance node.Lifecycle) error {
	return nodeInstance.Stop()
}

// BuildThorBinary builds the thor binary if needed and returns the path
func (m *Manager) BuildThorBinary(thorBuilder *thorbuilder.Config) (string, error) {
	if thorBuilder == nil {
		return "", nil // No thor builder configuration
	}

	execPath, err := thorbuilder.NewAndBuild(thorBuilder)
	if err != nil {
		return "", fmt.Errorf("failed to build thor: %w", err)
	}

	return execPath, nil
}

// ValidateNode validates and sets defaults for a node configuration
func (m *Manager) ValidateNode(nodeCfg node.Config) error {
	if nodeCfg.GetExecArtifact() == "" {
		return fmt.Errorf("exec artifact cannot be empty")
	}

	// Check if the exec artifact path exists
	if !fileExists(nodeCfg.GetExecArtifact()) {
		return fmt.Errorf("exec artifact path %s does not exist for node %s", nodeCfg.GetExecArtifact(), nodeCfg.GetID())
	}

	// Set default directories if not configured
	if nodeCfg.GetConfigDir() == "" {
		nodeCfg.SetConfigDir(filepath.Join(filepath.Dir(nodeCfg.GetExecArtifact()), nodeCfg.GetID(), "config"))
	}

	if nodeCfg.GetDataDir() == "" {
		nodeCfg.SetDataDir(filepath.Join(filepath.Dir(nodeCfg.GetExecArtifact()), nodeCfg.GetID(), "data"))
	}

	return nil
}

// GenerateEnodes creates enode strings for all nodes (excluding public network nodes)
func (m *Manager) GenerateEnodes(networkCfg *network.Network) ([]string, error) {
	var enodes []string

	// Skip enode generation entirely for public networks (testnet/mainnet)
	if networkCfg.IsPublicNetwork() {
		return enodes, nil
	}

	for _, node := range networkCfg.Nodes {
		// Use localhost for local environment
		enode, err := node.Enode("127.0.0.1")
		if err != nil {
			return nil, fmt.Errorf("failed to generate enode for node %s: %w", node.GetID(), err)
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
