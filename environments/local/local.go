package local

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/vechain/networkhub/environments"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/thorbuilder"
	"github.com/vechain/networkhub/utils/ports"
)

type Local struct {
	localNodes map[string]*Node
	networkCfg *network.Network
	id         string
	started    bool
	execPath   string // path to the thor binary

	mu sync.Mutex
}

var buildMutex sync.Mutex

type Factory struct{}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) New() environments.Actions {
	return NewEnv()
}

func NewEnv() *Local {
	return &Local{
		localNodes: make(map[string]*Node),
	}
}

func (l *Local) LoadConfig(cfg *network.Network) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.networkCfg = cfg
	l.id = l.networkCfg.ID()

	if cfg.ThorBuilder != nil {
		// serialize repository download/build across concurrent test runs
		buildMutex.Lock()
		defer buildMutex.Unlock()

		builder := thorbuilder.New(cfg.ThorBuilder)
		if err := builder.Download(); err != nil {
			return "", err
		}

		path, err := builder.Build()
		if err != nil {
			return "", fmt.Errorf("failed to build thor binary - %w", err)
		}

		l.execPath = path
	}

	for _, n := range l.networkCfg.Nodes {
		if err := l.ensureNodePorts(n); err != nil {
			return "", err
		}
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

	for _, nodeConfig := range l.networkCfg.Nodes {
		if err := nodeConfig.HealthCheck(0, 30*time.Second); err != nil {
			return err
		}
	}

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

	// release allocated ports for this network id
	if err := ports.Default().ReleaseAll(l.id); err != nil {
		return fmt.Errorf("failed to release ports: %w", err)
	}

	return nil
}

// AttachNode adds a node to the existing network.
// If the network has started, it will start the node.
func (l *Local) AttachNode(
	n node.Config,
	buildConfig *thorbuilder.Config,
	additionalArgs map[string]string,
) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.validateAttachable(n); err != nil {
		return err
	}
	if err := l.buildExecIfNeeded(buildConfig, n); err != nil {
		return err
	}

	for k, v := range additionalArgs {
		n.AddAdditionalArg(k, v)
	}

	if err := l.ensureNodeReady(n); err != nil {
		return err
	}

	enodesList, err := l.enodes()
	if err != nil {
		return fmt.Errorf("failed to get enodes - %w", err)
	}
	localNode := NewLocalNode(n, enodesList)
	l.localNodes[n.GetID()] = localNode
	l.networkCfg.Nodes = append(l.networkCfg.Nodes, n)

	if l.started {
		if err := localNode.Start(); err != nil {
			return fmt.Errorf("unable to start node %s after attaching - %w", n.GetID(), err)
		}
	}

	if err := n.HealthCheck(0, 30*time.Second); err != nil {
		return fmt.Errorf("failed to health check attached node: %w", err)
	}
	return nil
}

// validateAttachable ensures the network is loaded and the node id is unique.
func (l *Local) validateAttachable(n node.Config) error {
	if l.networkCfg == nil {
		return fmt.Errorf("network configuration is not loaded")
	}
	if _, exists := l.localNodes[n.GetID()]; exists {
		return fmt.Errorf("node with ID %s already exists", n.GetID())
	}
	return nil
}

// buildExecIfNeeded downloads/builds thor and sets the exec artifact if a config is provided.
func (l *Local) buildExecIfNeeded(buildConfig *thorbuilder.Config, n node.Config) error {
	if buildConfig == nil {
		return nil
	}
	buildMutex.Lock()
	defer buildMutex.Unlock()
	builder := thorbuilder.New(buildConfig)
	if err := builder.Download(); err != nil {
		return err
	}
	path, err := builder.Build()
	if err != nil {
		return fmt.Errorf("failed to build node: %w", err)
	}
	n.SetExecArtifact(path)
	return nil
}

// ensureNodeReady makes sure ports are assigned and file paths are valid.
func (l *Local) ensureNodeReady(n node.Config) error {
	if err := l.ensureNodePorts(n); err != nil {
		return err
	}
	if err := l.checkNode(n); err != nil {
		return err
	}
	return nil
}

func (l *Local) RemoveNode(nodeID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, exists := l.localNodes[nodeID]; !exists {
		return fmt.Errorf("node with ID %s does not exist", nodeID)
	}

	index, found := -1, false
	for i, n := range l.networkCfg.Nodes {
		if n.GetID() == nodeID {
			index = i
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("node with ID %s not found in network configuration", nodeID)
	}

	if err := l.localNodes[nodeID].Stop(); err != nil {
		return fmt.Errorf("unable to stop node %s - %w", nodeID, err)
	}

	l.networkCfg.Nodes = append(l.networkCfg.Nodes[:index], l.networkCfg.Nodes[index+1:]...)
	delete(l.localNodes, nodeID)
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

func (l *Local) checkNode(n node.Config) error {
	// check if the exec artifact path exists
	if !fileExists(n.GetExecArtifact()) {
		if l.execPath == "" {
			return fmt.Errorf("exec artifact path %s does not exist for node %s", n.GetExecArtifact(), n.GetID())
		}
		n.SetExecArtifact(l.execPath)
	}

	if n.GetConfigDir() == "" {
		n.SetConfigDir(filepath.Join(filepath.Dir(n.GetExecArtifact()), n.GetID(), "config"))
	}

	if n.GetDataDir() == "" {
		n.SetDataDir(filepath.Join(filepath.Dir(n.GetExecArtifact()), n.GetID(), "data"))
	}
	return nil
}

// ensureNodePorts assigns API and P2P ports if missing, using the shared port manager.
func (l *Local) ensureNodePorts(n node.Config) error {
	if n.GetAPIAddr() == "" {
		allocated, err := ports.Default().Allocate(l.id)
		if err != nil {
			return fmt.Errorf("failed to allocate API port: %w", err)
		}
		n.SetAPIAddr(fmt.Sprintf("%s:%d", "0.0.0.0", allocated))
	}

	if n.GetP2PListenPort() <= 0 {
		p, err := ports.Default().Allocate(l.id)
		if err != nil {
			return fmt.Errorf("failed to allocate P2P port: %w", err)
		}
		n.SetP2PListenPort(p)
	}
	return nil
}
