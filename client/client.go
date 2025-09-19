package client

import (
	"fmt"

	"github.com/vechain/networkhub/environments"
	"github.com/vechain/networkhub/environments/docker"
	"github.com/vechain/networkhub/environments/local"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
)

type Client struct {
	network     *network.Network
	environment environments.Actions
	factories   map[string]environments.Factory
}

func New() *Client {
	factories := map[string]environments.Factory{
		"local":  local.NewFactory(),
		"docker": docker.NewFactory(),
	}

	return &Client{
		factories: factories,
	}
}

func NewWithNetwork(net *network.Network) (*Client, error) {
	c := New()
	if err := c.LoadNetwork(net); err != nil {
		return nil, err
	}
	return c, nil
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

func (c *Client) LoadNetwork(net *network.Network) error {
	factory, ok := c.factories[net.Environment]
	if !ok {
		return fmt.Errorf("unsupported environment: %s", net.Environment)
	}

	env := factory.New()
	_, err := env.LoadConfig(net)
	if err != nil {
		return fmt.Errorf("failed to load network config: %w", err)
	}

	c.network = net
	c.environment = env
	return nil
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
