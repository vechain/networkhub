package client

import (
	"fmt"
	"strings"

	"github.com/vechain/networkhub/internal/environments"
	"github.com/vechain/networkhub/internal/environments/overseer"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
)

type Client struct {
	network     *network.Network
	environment environments.Actions
}

func New(net *network.Network) (*Client, error) {
	env, err := overseer.New(net)
	if err != nil {
		return nil, err
	}

	c := &Client{
		network:     net,
		environment: env,
	}

	// Auto-start for public networks (testnet/mainnet)
	if strings.Contains(net.ID(), "mainnet") || strings.Contains(net.ID(), "testnet") {
		if err := c.Start(); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// NewWithNetwork is an alias for New for backward compatibility
func NewWithNetwork(net *network.Network) (*Client, error) {
	return New(net)
}

func (c *Client) Stop() error {
	if c.environment == nil {
		return fmt.Errorf("no network loaded")
	}
	return c.environment.StopNetwork()
}

func (c *Client) Start() error {
	if c.environment == nil {
		return fmt.Errorf("no network loaded")
	}
	return c.environment.StartNetwork()
}

func (c *Client) GetNetwork() (*network.Network, error) {
	if c.network == nil {
		return nil, fmt.Errorf("no network loaded")
	}
	return c.network, nil
}

func (c *Client) Nodes() (map[string]node.Lifecycle, error) {
	if c.environment == nil {
		return nil, fmt.Errorf("no network loaded")
	}
	return c.environment.Nodes(), nil
}

func (c *Client) AddNode(nodeConfig node.Config) error {
	if c.network == nil {
		return fmt.Errorf("no network loaded")
	}

	if c.environment == nil {
		return fmt.Errorf("environment not initialized")
	}

	// Use environment's AddNode method
	if err := c.environment.AddNode(nodeConfig); err != nil {
		return fmt.Errorf("failed to add node to environment: %w", err)
	}

	// Update client's network reference
	c.network = c.environment.Config()

	return nil
}

func (c *Client) RemoveNode(nodeID string) error {
	if c.network == nil {
		return fmt.Errorf("no network loaded")
	}

	if c.environment == nil {
		return fmt.Errorf("environment not initialized")
	}

	// Use environment's RemoveNode method
	if err := c.environment.RemoveNode(nodeID); err != nil {
		return fmt.Errorf("failed to remove node from environment: %w", err)
	}

	// Update client's network reference
	c.network = c.environment.Config()

	return nil
}
