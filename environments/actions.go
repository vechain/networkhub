package environments

import (
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
)

type Factory interface {
	New() Actions
}

type Actions interface {
	LoadConfig(cfg *network.Network) (string, error)
	StartNetwork() error
	StopNetwork() error
	Nodes() map[string]node.Lifecycle
	Config() *network.Network
}

const (
	Local = "local"
)
