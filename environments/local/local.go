package local

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/vechain/networkhub/environments"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
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
	l.id = hashObject(cfg)
	return l.id, nil
}

func (l *Local) StartNetwork() error {
	var enodes []string
	for _, node := range l.networkCfg.Nodes {
		enodes = append(enodes, node.Enode)
	}

	enodeString := strings.Join(enodes, ",")
	for _, node := range l.networkCfg.Nodes {
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

func (l *Local) startNode(nodeCfg *node.Node, enodeString string) (*Node, error) {
	localNode := NewLocalNode(nodeCfg, enodeString)
	return localNode, localNode.Start()
}

func hashObject(obj interface{}) string {
	// Create a buffer to hold the encoded data
	var buf bytes.Buffer

	// New encoder that writes to the buffer
	encoder := gob.NewEncoder(&buf)

	// Encode the object; handle errors
	if err := encoder.Encode(obj); err != nil {
		log.Fatalf("Failed to encode object: %v", err)
	}

	// Compute SHA-256 checksum on the buffer's bytes
	hash := sha256.New()
	hash.Write(buf.Bytes())
	hashBytes := hash.Sum(nil)

	// Convert hash bytes to hex string
	return hex.EncodeToString(hashBytes)
}
