package local

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/vechain/networkhub/environments"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/thorbuilder"
)

type Local struct {
	localNodes map[string]*Node
	networkCfg *network.Network
	id         string
	started    bool
	execPath   string // path to the thor binary

	mu sync.Mutex
}

type Factory struct{}

type PublicNetworkConfig struct {
	NodeID      string
	NetworkType string // "test" for testnet, "main" for mainnet
	APIAddr     string
	P2PPort     int
	Branch      string
}

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
func (l *Local) AttachNode(n node.Config) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.networkCfg == nil {
		return fmt.Errorf("network configuration is not loaded")
	}

	if _, exists := l.localNodes[n.GetID()]; exists {
		return fmt.Errorf("node with ID %s already exists", n.GetID())
	}

	if err := l.checkNode(n); err != nil {
		return err
	}

	endodes, err := l.enodes()
	if err != nil {
		return fmt.Errorf("failed to get enodes - %w", err)
	}
	localNode := NewLocalNode(n, endodes)
	l.localNodes[n.GetID()] = localNode
	l.networkCfg.Nodes = append(l.networkCfg.Nodes, n)

	if l.started {
		// If the network has already started, we need to start the node immediately.
		if err := localNode.Start(); err != nil {
			return fmt.Errorf("unable to start node %s after attaching - %w", n.GetID(), err)
		}
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
		// Skip enode generation for public network nodes (testnet/mainnet) and solo nodes
		if shouldNotAppendEnodes(node) {
			continue
		}

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

// startNetworkWithConfig is a helper function that creates and starts a network with the given configuration
func (l *Local) startNetworkWithConfig(nodes []node.Config, environment, baseID, branch string) error {
	// Determine Thor branch to use
	thorBranch := "master"
	if branch != "" {
		thorBranch = branch
	}

	// Create network configuration
	networkCfg := &network.Network{
		BaseID:      baseID,
		Environment: environment,
		Nodes:       nodes,
		ThorBuilder: &thorbuilder.Config{
			DownloadConfig: &thorbuilder.DownloadConfig{
				RepoUrl:    "https://github.com/vechain/thor",
				Branch:     thorBranch,
				IsReusable: true,
			},
		},
	}

	// Load the configuration
	_, err := l.LoadConfig(networkCfg)
	if err != nil {
		return fmt.Errorf("failed to load network configuration: %w", err)
	}

	// Start the network
	err = l.StartNetwork()
	if err != nil {
		return fmt.Errorf("failed to start network: %w", err)
	}

	return nil
}

// AttachToPublicNetworkAndStart creates and starts a node connected to a public network (testnet/mainnet)
func (l *Local) AttachToPublicNetworkAndStart(config PublicNetworkConfig) error {
	publicNode := &node.BaseNode{
		ID:             config.NodeID,
		APICORS:        "*",
		Type:           node.RegularNode,
		Verbosity:      3,
		P2PListenPort:  config.P2PPort,
		APIAddr:        config.APIAddr,
		AdditionalArgs: map[string]string{"network": config.NetworkType},
	}

	environment := "testnet"
	if config.NetworkType == "main" {
		environment = "mainnet"
	}

	return l.startNetworkWithConfig([]node.Config{publicNode}, environment, "baseID", config.Branch)
}

// StartSoloNode creates and starts a solo node
func (l *Local) StartSoloNode(config SoloNodeConfig) error {
	soloNodeConfig := CreateSoloNodeConfig(config)
	return l.startNetworkWithConfig([]node.Config{soloNodeConfig}, "solo", "solo-network", config.Branch)
}
