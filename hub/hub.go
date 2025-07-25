package hub

import (
	"fmt"

	"github.com/vechain/networkhub/environments"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
)

type NetworkHub struct {
	envs               map[string]environments.Actions
	configuredNetworks map[string]environments.Actions
	networks           map[string]*network.Network
}

func NewNetworkHub() *NetworkHub {
	return &NetworkHub{
		envs:               map[string]environments.Actions{},
		configuredNetworks: map[string]environments.Actions{},
		networks:           map[string]*network.Network{},
	}
}

func (e *NetworkHub) LoadNetworkConfig(cfg *network.Network) (string, error) {
	env, ok := e.envs[cfg.Environment]
	if !ok {
		return "", fmt.Errorf("unable to load env %s", cfg.Environment)
	}

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
	netwk, err := e.GetNetwork(networkID)
	if err != nil {
		return err
	}

	return netwk.StopNetwork()
}

func (e *NetworkHub) Nodes(networkID string) map[string]node.Lifecycle {
	netwk, ok := e.configuredNetworks[networkID]
	if !ok {
		return nil
	}
	return netwk.Nodes()
}

func (e *NetworkHub) RegisterEnvironment(id string, env environments.Actions) {
	e.envs[id] = env
}

func (e *NetworkHub) GetNetworkConfig(id string) (*network.Network, error) {
	loadedNetwork, ok := e.networks[id]
	if !ok {
		return nil, fmt.Errorf("network not found")
	}
	return loadedNetwork, nil
}

func (e *NetworkHub) GetNetwork(id string) (environments.Actions, error) {
	loadedNetwork, ok := e.configuredNetworks[id]
	if !ok {
		return nil, fmt.Errorf("network not found")
	}
	return loadedNetwork, nil
}
