package client

import (
	"fmt"
	"strings"

	"github.com/vechain/networkhub/internal/environments"
	"github.com/vechain/networkhub/internal/environments/launcher"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
)

type Client struct {
	network *network.Network
	actions environments.Actions
}

func New(net *network.Network) (*Client, error) {
	env, err := launcher.New(net)
	if err != nil {
		return nil, err
	}

	c := &Client{
		network: net,
		actions: env,
	}

	// Auto-start for public networks (testnet/mainnet)
	if strings.Contains(net.ID(), network.Mainnet) || strings.Contains(net.ID(), network.Testnet) {
		if err := c.Start(); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Client) Stop() error {
	if c.actions == nil {
		return fmt.Errorf("no network loaded")
	}
	return c.actions.StopNetwork()
}

func (c *Client) Start() error {
	if c.actions == nil {
		return fmt.Errorf("no network loaded")
	}
	return c.actions.StartNetwork()
}

func (c *Client) GetNetwork() (*network.Network, error) {
	if c.network == nil {
		return nil, fmt.Errorf("no network loaded")
	}
	return c.network, nil
}

func (c *Client) Nodes() (map[string]node.Lifecycle, error) {
	if c.actions == nil {
		return nil, fmt.Errorf("no network loaded")
	}
	return c.actions.Nodes(), nil
}

func (c *Client) AddNode(nodeConfig node.Config) error {
	if c.network == nil {
		return fmt.Errorf("no network loaded")
	}

	if c.actions == nil {
		return fmt.Errorf("environment not initialized")
	}

	// Use environment's AddNode method
	if err := c.actions.AddNode(nodeConfig); err != nil {
		return fmt.Errorf("failed to add node to environment: %w", err)
	}

	// Update client's network reference
	c.network = c.actions.Config()

	return nil
}

func (c *Client) RemoveNode(nodeID string) error {
	if c.network == nil {
		return fmt.Errorf("no network loaded")
	}

	if c.actions == nil {
		return fmt.Errorf("environment not initialized")
	}

	// Use environment's RemoveNode method
	if err := c.actions.RemoveNode(nodeID); err != nil {
		return fmt.Errorf("failed to remove node from environment: %w", err)
	}

	// Update client's network reference
	c.network = c.actions.Config()

	return nil
}
