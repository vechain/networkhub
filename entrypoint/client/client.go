package client

import (
	"encoding/json"
	"fmt"
	"github.com/vechain/networkhub/hub"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/preset"
)

type Client struct {
	networkHub *hub.NetworkHub
	presets    *preset.Networks
	storage    *Storage
}

func New(networkHub *hub.NetworkHub, presets *preset.Networks) *Client {
	return &Client{
		networkHub: networkHub,
		presets:    presets,
		storage:    NewInMemStorage(),
	}
}

func (c *Client) Stop(id string) error {
	return c.networkHub.StopNetwork(id)
}

func (c *Client) Start(id string) error {
	return c.networkHub.StartNetwork(id)
}

func (c *Client) Config(config string) (string, error) {
	var netCfg network.Network

	if err := json.Unmarshal([]byte(config), &netCfg); err != nil {
		return "", err
	}
	return c.config(&netCfg)
}

func (c *Client) LoadExistingNetworks() error {
	nets, err := c.storage.LoadExistingNetworks()
	if err != nil {
		return fmt.Errorf("unable to load existing networks: %w", err)
	}

	for networkID, net := range nets {
		loadedID, err := c.networkHub.LoadNetworkConfig(net)
		if err != nil {
			return err
		}

		if networkID != loadedID {
			return fmt.Errorf("unexpected networkID loaded: storedID:%s configuredID:%s", networkID, loadedID)
		}
	}

	return nil
}

func (c *Client) Preset(presetNetwork string, environment, artifactPath string) (string, error) {
	netCfg, err := c.presets.Load(presetNetwork, &preset.APIConfigPayload{Environment: environment, ArtifactPath: artifactPath})
	if err != nil {
		return "", fmt.Errorf("unable to load network preset: %w", err)
	}
	return c.config(netCfg)
}

func (c *Client) config(netCfg *network.Network) (string, error) {
	networkID, err := c.networkHub.LoadNetworkConfig(netCfg)
	if err != nil {
		return "", fmt.Errorf("unable to load config: %w", err)
	}

	networkInst, err := c.networkHub.GetNetwork(networkID)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve network: %w", err)
	}

	err = c.storage.Store(networkID, networkInst)
	if err != nil {
		return "", fmt.Errorf("unable to store network: %w", err)
	}

	return networkID, nil
}
