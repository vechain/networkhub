package local

import (
	"fmt"
	"strings"

	"github.com/vechain/networkhub/environments"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
)

type Local struct {
	localNodes map[string]*LocalNode
}

func NewLocalEnv() environments.Actions {
	return &Local{
		localNodes: map[string]*LocalNode{},
	}
}

func (l *Local) LoadConfig() error {
	//TODO implement me
	panic("implement me")
}

func (l *Local) StartNetwork(cfg *network.Network) error {
	var enodes []string
	for _, node := range cfg.Nodes {
		enodes = append(enodes, node.Enode)
	}

	enodeString := strings.Join(enodes, ",")
	for _, node := range cfg.Nodes {
		localNode, err := l.startNode(node, enodeString)
		if err != nil {
			return fmt.Errorf("unable to start node - %w", err)
		}
		l.localNodes[node.ID] = localNode
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

func (l *Local) startNode(nodeCfg *node.Node, enodeString string) (*LocalNode, error) {
	localNode := NewLocalNode(nodeCfg, enodeString)
	return localNode, localNode.Start()
}
