package client

import (
	"fmt"

	"github.com/vechain/networkhub/environments/docker"

	"github.com/vechain/networkhub/environments/local"
	"github.com/vechain/networkhub/hub"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/preset"
)

type Client struct {
	networkHub *hub.NetworkHub
	presets    *preset.Networks
	storage    *Storage
}

func New() *Client {
	envManager := hub.NewNetworkHub()
	envManager.RegisterEnvironment("local", local.NewLocalEnv)
	envManager.RegisterEnvironment("docker", docker.NewDockerEnv)

	presets := preset.NewPresetNetworks()
	//presets.Register("threeMasterNodesNetwork", preset.LocalThreeMasterNodesNetwork())
	//presets.Register("sixNodesNetwork", preset.LocalSixNodesNetwork())

	return &Client{
		networkHub: envManager,
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
	return c.Config(netCfg)
}

func (c *Client) Config(netCfg *network.Network) (string, error) {
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
