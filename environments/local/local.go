package local

import (
	"fmt"
	"os"
	"path/filepath"

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

	// ensure paths exist, use ExecArtifact base dirs if not defined
	for _, n := range l.networkCfg.Nodes {
		// check if the exec artifact path exists
		if !fileExists(n.GetExecArtifact()) {
			return "", fmt.Errorf("exec does not exist at path: %s", n.GetExecArtifact())
		}

		if n.GetConfigDir() == "" {
			n.SetConfigDir(filepath.Join(filepath.Dir(n.GetExecArtifact()), n.GetID(), "config"))
		}

		if n.GetDataDir() == "" {
			n.SetDataDir(filepath.Join(filepath.Dir(n.GetExecArtifact()), n.GetID(), "data"))
		}
	}

	return l.id, nil
}

func (l *Local) StartNetwork() error {
	// speed up p2p bootstrap
	var enodes []string
	for _, node := range l.networkCfg.Nodes {
		enode, err := node.Enode("127.0.0.1")
		if err != nil {
			return err
		}
		enodes = append(enodes, enode)
	}
	
	for _, nodeCfg := range l.networkCfg.Nodes {
		localNode := NewLocalNode(nodeCfg, enodes)
		if err := localNode.Start(); err != nil {
			return fmt.Errorf("unable to start node - %w", err)
		}

		l.localNodes[nodeCfg.GetID()] = localNode
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}
