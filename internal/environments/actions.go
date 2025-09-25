package environments

import (
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
)

type Actions interface {
	StartNetwork() error
	StopNetwork() error
	Nodes() map[string]node.Lifecycle
	Config() *network.Network
	AddNode(nodeConfig node.Config) error
	RemoveNode(nodeID string) error
}

const (
	Local = "local"
)
