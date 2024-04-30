package noop

import (
	"github.com/vechain/networkhub/environments"
	"github.com/vechain/networkhub/network"
)

type Noop struct{}

func NewNoopEnv() environments.Actions {
	return &Noop{}
}

func (n Noop) LoadConfig() error {
	return nil
}

func (n Noop) StartNetwork(cfg *network.Network) error {
	return nil
}

func (n Noop) StopNetwork() error {
	return nil
}

func (n Noop) Info() error {
	return nil
}
