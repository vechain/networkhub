package environments

import (
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
)

type Actions interface {
	LoadConfig(cfg *network.Network) (string, error)
	StartNetwork() error
	StopNetwork() error
	Nodes() map[string]node.Lifecycle
	Info() error
}

const (
	Local = "local"
)
