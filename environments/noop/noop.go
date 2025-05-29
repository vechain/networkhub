package noop

import (
	"github.com/vechain/networkhub/environments"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
)

type Noop struct{}

func NewNoopEnv() environments.Actions {
	return &Noop{}
}

func (n Noop) LoadConfig(*network.Network) (string, error) {
	return "noop", nil
}

func (n Noop) StartNetwork() error {
	return nil
}

func (n Noop) StopNetwork() error {
	return nil
}

func (n Noop) Nodes() map[string]node.Lifecycle {
	return map[string]node.Lifecycle{}
}

func (n Noop) Info() error {
	return nil
}

func (n Noop) Config() *network.Network { return nil }
