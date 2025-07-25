package local

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/vechain/networkhub/environments"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
)

type Local struct {
	localNodes map[string]*Node
	networkCfg *network.Network
	id         string
	started    bool

	mu sync.Mutex
}

func NewLocalEnv() environments.Actions {
	return &Local{
		localNodes: map[string]*Node{},
	}
}

var _ environments.Actions = (*Local)(nil)

func (l *Local) LoadConfig(cfg *network.Network) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.networkCfg = cfg
	l.id = l.networkCfg.ID()

	for _, n := range l.networkCfg.Nodes {
		if err := l.checkNode(n); err != nil {
			return "", fmt.Errorf("failed to check node %s - %w", n.GetID(), err)
		}
	}

	return l.id, nil
}

func (l *Local) StartNetwork() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// speed up p2p bootstrap
	enodes, err := l.enodes()
	if err != nil {
		return fmt.Errorf("failed to get enodes - %w", err)
	}

	for _, nodeCfg := range l.networkCfg.Nodes {
		localNode := NewLocalNode(nodeCfg, enodes)
		if err := localNode.Start(); err != nil {
			return fmt.Errorf("unable to start node - %w", err)
		}

		l.localNodes[nodeCfg.GetID()] = localNode
	}
	l.started = true

	return nil
}

func (l *Local) StopNetwork() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for s, localNode := range l.localNodes {
		err := localNode.Stop()
		if err != nil {
			return fmt.Errorf("unable to stop node %s - %w", s, err)
		}
	}

	l.localNodes = make(map[string]*Node)
	l.started = false

	return nil
}

// AttachNode adds a node to the existing network.
// If the network has started, it will start the node.
func (l *Local) AttachNode(n *node.Config) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, exists := l.localNodes[n.GetID()]; exists {
		return fmt.Errorf("node with ID %s already exists", n.GetID())
	}

	endodes, err := l.enodes()
	if err != nil {
		return fmt.Errorf("failed to get enodes - %w", err)
	}
	localNode := NewLocalNode(n, endodes)
	l.localNodes[n.GetID()] = localNode
	if err := localNode.Start(); err != nil {
		return fmt.Errorf("unable to start node %s - %w", n.GetID(), err)
	}

	if l.started {
		// If the network has already started, we need to start the node immediately.
		if err := localNode.Start(); err != nil {
			return fmt.Errorf("unable to start node %s after attaching - %w", n.GetID(), err)
		}
	}

	return nil
}

func (l *Local) Nodes() map[string]node.Lifecycle {
	l.mu.Lock()
	defer l.mu.Unlock()

	nodes := make(map[string]node.Lifecycle, len(l.localNodes))
	for k, v := range l.localNodes {
		nodes[k] = v
	}
	return nodes
}

func (l *Local) Config() *network.Network {
	return l.networkCfg
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func (l *Local) enodes() ([]string, error) {
	var enodes []string
	for _, node := range l.networkCfg.Nodes {
		enode, err := node.Enode("127.0.0.1")
		if err != nil {
			return nil, fmt.Errorf("failed to get enode for node %s: %w", node.GetID(), err)
		}
		enodes = append(enodes, enode)
	}
	return enodes, nil
}

func (l *Local) checkNode(n *node.Config) error {
	// check if the exec artifact path exists
	if !fileExists(n.GetExecArtifact()) {
		return fmt.Errorf("exec does not exist at path: %s", n.GetExecArtifact())
	}

	if n.GetConfigDir() == "" {
		n.SetConfigDir(filepath.Join(filepath.Dir(n.GetExecArtifact()), n.GetID(), "config"))
	}

	if n.GetDataDir() == "" {
		n.SetDataDir(filepath.Join(filepath.Dir(n.GetExecArtifact()), n.GetID(), "data"))
	}
	return nil
}
