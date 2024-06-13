package hub

import (
	"fmt"
	"github.com/vechain/networkhub/environments"
	"github.com/vechain/networkhub/network"
)

type NetworkHub struct {
	envFuncs           map[string]func() environments.Actions
	configuredNetworks map[string]environments.Actions
	networks           map[string]*network.Network
}

func NewNetworkHub() *NetworkHub {
	return &NetworkHub{
		envFuncs:           map[string]func() environments.Actions{},
		configuredNetworks: map[string]environments.Actions{},
		networks:           map[string]*network.Network{},
	}
}

func (e *NetworkHub) LoadNetworkConfig(cfg *network.Network) (string, error) {
	envFunc, ok := e.envFuncs[cfg.Environment]
	if !ok {
		return "", fmt.Errorf("unable to load env %s", cfg.Environment)
	}

	env := envFunc()
	networkID, err := env.LoadConfig(cfg)
	if err != nil {
		return "", fmt.Errorf("unable to load config - %w", err)
	}
	e.configuredNetworks[networkID] = env
	e.networks[networkID] = cfg

	return networkID, nil
}

func (e *NetworkHub) StartNetwork(networkID string) error {
	netwk, ok := e.configuredNetworks[networkID]
	if !ok {
		return fmt.Errorf("network %s is not configured", networkID)
	}

	return netwk.StartNetwork()
}

func (e *NetworkHub) StopNetwork(networkID string) error {
	netwk, ok := e.configuredNetworks[networkID]
	if !ok {
		return fmt.Errorf("network %s is not configured", networkID)
	}
	return netwk.StopNetwork()
}

func (e *NetworkHub) InfoNetwork(networkID string) error {
	netwk, ok := e.configuredNetworks[networkID]
	if !ok {
		return fmt.Errorf("network %s is not configured", networkID)
	}
	return netwk.StopNetwork()
}

func (e *NetworkHub) RegisterEnvironment(id string, env func() environments.Actions) {
	e.envFuncs[id] = env
}

func (e *NetworkHub) GetNetwork(id string) (*network.Network, error) {
	loadedNetwork, ok := e.networks[id]
	if !ok {
		return nil, fmt.Errorf("network not found")
	}
	return loadedNetwork, nil
}
