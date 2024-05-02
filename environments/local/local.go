package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vechain/networkhub/environments"
	"github.com/vechain/networkhub/network"
)

type Local struct {
	localNodes map[string]*Node
	networkCfg *network.Network
	id         string
}

func NewLocalEnv() environments.Actions {
	return &Local{
		localNodes: map[string]*Node{},
	}
}

func (l *Local) LoadConfig(cfg *network.Network) (string, error) {
	l.networkCfg = cfg
	l.id = l.networkCfg.Environment + l.networkCfg.ID
	baseTmpDir := filepath.Join(os.TempDir(), l.id)

	// ensure paths exist, use temp dirs if not defined
	for _, n := range l.networkCfg.Nodes {
		if n.ConfigDir == "" {
			n.ConfigDir = filepath.Join(baseTmpDir, n.ID, "config")
		}

		if n.DataDir == "" {
			n.DataDir = filepath.Join(baseTmpDir, n.ID, "data")
		}
	}

	return l.id, nil
}

func (l *Local) StartNetwork() error {
	// speed up p2p bootstrap
	var enodes []string
	for _, node := range l.networkCfg.Nodes {
		enodes = append(enodes, node.Enode)
	}
	enodeString := strings.Join(enodes, ",")

	for _, nodeCfg := range l.networkCfg.Nodes {
		localNode := NewLocalNode(nodeCfg, enodeString)
		if err := localNode.Start(); err != nil {
			return fmt.Errorf("unable to start node - %w", err)
		}

		l.localNodes[nodeCfg.ID] = localNode
	}

	return nil
}

func (l *Local) StopNetwork() error {
	for s, localNode := range l.localNodes {
		err := localNode.Stop()
		if err != nil {
			return fmt.Errorf("unable to stop node %s - %w", s, err)
		}
	}
	return nil
}

func (l *Local) Info() error {
	//TODO implement me
	panic("implement me")
}
