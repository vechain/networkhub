package overseer

import (
	"fmt"
	"sync"

	"github.com/vechain/networkhub/internal/environments/docker"
	"github.com/vechain/networkhub/internal/environments/local"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
)

// Environment defines the interface that environments must implement
type Environment interface {
	// StartNetwork starts all nodes in the network
	StartNetwork() error
	// StopNetwork stops all nodes in the network
	StopNetwork() error
	// AddNode adds a node to the existing network
	AddNode(nodeConfig node.Config) error
	// RemoveNode removes a node from the network
	RemoveNode(nodeID string) error
	// Nodes returns all nodes in the environment
	Nodes() map[string]node.Lifecycle
	// Config returns the network configuration
	Config() *network.Network
}

// Overseer manages environments and acts as the single point of contact
type Overseer struct {
	environment Environment
	mu          sync.Mutex
}

// New creates a new overseer instance with the given network configuration
func New(cfg *network.Network) (*Overseer, error) {
	// Create the appropriate environment based on environment type
	var env Environment
	var err error
	switch cfg.Environment {
	case "local":
		env, err = createLocalEnvironment(cfg)
	case "docker":
		env, err = createDockerEnvironment(cfg)
	default:
		return nil, fmt.Errorf("unsupported environment: %s", cfg.Environment)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create %s environment: %w", cfg.Environment, err)
	}

	return &Overseer{
		environment: env,
	}, nil
}

// StartNetwork starts all nodes in the network
func (o *Overseer) StartNetwork() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.environment == nil {
		return fmt.Errorf("no environment loaded")
	}
	return o.environment.StartNetwork()
}

// StopNetwork stops all nodes in the network
func (o *Overseer) StopNetwork() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.environment == nil {
		return fmt.Errorf("no environment loaded")
	}
	return o.environment.StopNetwork()
}

// AddNode adds a node to the existing network
func (o *Overseer) AddNode(nodeConfig node.Config) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.environment == nil {
		return fmt.Errorf("no environment loaded")
	}
	return o.environment.AddNode(nodeConfig)
}

// RemoveNode removes a node from the network
func (o *Overseer) RemoveNode(nodeID string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.environment == nil {
		return fmt.Errorf("no environment loaded")
	}
	return o.environment.RemoveNode(nodeID)
}

// Nodes returns all nodes in the environment
func (o *Overseer) Nodes() map[string]node.Lifecycle {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.environment == nil {
		return make(map[string]node.Lifecycle)
	}
	return o.environment.Nodes()
}

// Config returns the network configuration
func (o *Overseer) Config() *network.Network {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.environment == nil {
		return nil
	}
	return o.environment.Config()
}

// Factory functions to create environments
func createLocalEnvironment(cfg *network.Network) (Environment, error) {
	return local.NewEnvironment(cfg), nil
}

func createDockerEnvironment(cfg *network.Network) (Environment, error) {
	return docker.NewEnvironment(cfg), nil
}