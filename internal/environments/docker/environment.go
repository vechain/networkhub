package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	dockertypes "github.com/docker/docker/api/types"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/thorbuilder"
)

// Environment implements overseer.Environment for docker environments
type Environment struct {
	nodes        map[string]node.Lifecycle
	networkCfg   *network.Network
	started      bool
	networkID    string
	exposedPorts map[string]*ExposedPort
	ipManager    *IpManager
	dockerImage  string
	mu           sync.Mutex
}

// NewEnvironment creates a new docker environment
func NewEnvironment(cfg *network.Network) *Environment {
	// Set up Docker-specific configuration
	networkID := cfg.ID() + "-network"
	
	// Configure exposed ports for each node (this is just parsing, not heavy work)
	exposedPorts := make(map[string]*ExposedPort)
	for _, node := range cfg.Nodes {
		// ensure API ports are exposed to the localhost
		split := strings.Split(node.GetAPIAddr(), ":")
		if len(split) == 2 {
			if _, err := strconv.Atoi(split[1]); err == nil {
				exposedPorts[node.GetID()] = &ExposedPort{
					HostPort:      split[1], // Use the same port as container port
					ContainerPort: split[1],
				}
			}
		}
	}

	return &Environment{
		nodes:        make(map[string]node.Lifecycle),
		networkCfg:   cfg,
		networkID:    networkID,
		exposedPorts: exposedPorts,
		ipManager:    NewIPManagerRandom(),
	}
}

// StartNetwork starts all nodes in the network
func (e *Environment) StartNetwork() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.started {
		return fmt.Errorf("network already started")
	}

	// Build docker image if needed
	if err := e.buildDockerImageIfNeeded(); err != nil {
		return fmt.Errorf("failed to build docker image: %w", err)
	}

	// Create a network for fixed ip addresses (enodes cannot have dns names)
	if err := e.checkOrCreateNetwork(e.networkID, e.ipManager.Subnet()); err != nil {
		return fmt.Errorf("unable to create network: %w", err)
	}

	// Generate enodes for faster p2p bootstrap
	enodes, err := e.generateEnodes()
	if err != nil {
		return fmt.Errorf("failed to generate enodes: %w", err)
	}

	for _, nodeCfg := range e.networkCfg.Nodes {
		// Validate node before starting
		if err := e.checkNode(nodeCfg); err != nil {
			return fmt.Errorf("failed to validate node %s: %w", nodeCfg.GetID(), err)
		}

		nodeInstance, err := e.createNode(nodeCfg, enodes)
		if err != nil {
			return fmt.Errorf("failed to create node %s: %w", nodeCfg.GetID(), err)
		}

		if err := nodeInstance.Start(); err != nil {
			return fmt.Errorf("unable to start node %s: %w", nodeCfg.GetID(), err)
		}

		e.nodes[nodeCfg.GetID()] = nodeInstance
	}

	e.started = true
	return nil
}

// StopNetwork stops all nodes in the network
func (e *Environment) StopNetwork() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.started {
		return nil
	}

	for nodeID, nodeInstance := range e.nodes {
		if err := nodeInstance.Stop(); err != nil {
			return fmt.Errorf("unable to stop node %s: %w", nodeID, err)
		}
	}

	e.nodes = make(map[string]node.Lifecycle)
	e.started = false
	return nil
}

// AddNode adds a node to the existing network
func (e *Environment) AddNode(nodeConfig node.Config) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.networkCfg == nil {
		return fmt.Errorf("network configuration is not loaded")
	}

	if _, exists := e.nodes[nodeConfig.GetID()]; exists {
		return fmt.Errorf("node with ID %s already exists", nodeConfig.GetID())
	}

	e.networkCfg.Nodes = append(e.networkCfg.Nodes, nodeConfig)

	// If network is running, start the new node immediately
	if e.started {
		// Build docker image if needed for this node
		if err := e.buildDockerImageIfNeeded(); err != nil {
			return fmt.Errorf("failed to build docker image for new node: %w", err)
		}

		// Validate the node before starting
		if err := e.checkNode(nodeConfig); err != nil {
			return fmt.Errorf("failed to validate node %s: %w", nodeConfig.GetID(), err)
		}

		enodes, err := e.generateEnodes()
		if err != nil {
			return fmt.Errorf("failed to generate enodes: %w", err)
		}

		nodeInstance, err := e.createNode(nodeConfig, enodes)
		if err != nil {
			return fmt.Errorf("failed to create node %s: %w", nodeConfig.GetID(), err)
		}

		if err := nodeInstance.Start(); err != nil {
			return fmt.Errorf("unable to start node %s after adding: %w", nodeConfig.GetID(), err)
		}

		e.nodes[nodeConfig.GetID()] = nodeInstance
	}

	return nil
}

// RemoveNode removes a node from the network
func (e *Environment) RemoveNode(nodeID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	nodeInstance, exists := e.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node with ID %s does not exist", nodeID)
	}

	// Find and remove from network configuration
	index := -1
	for i, n := range e.networkCfg.Nodes {
		if n.GetID() == nodeID {
			index = i
			break
		}
	}
	if index == -1 {
		return fmt.Errorf("node with ID %s not found in network configuration", nodeID)
	}

	if err := nodeInstance.Stop(); err != nil {
		return fmt.Errorf("unable to stop node %s: %w", nodeID, err)
	}

	e.networkCfg.Nodes = append(e.networkCfg.Nodes[:index], e.networkCfg.Nodes[index+1:]...)
	delete(e.nodes, nodeID)

	return nil
}

// Nodes returns all nodes in the environment
func (e *Environment) Nodes() map[string]node.Lifecycle {
	e.mu.Lock()
	defer e.mu.Unlock()

	nodes := make(map[string]node.Lifecycle, len(e.nodes))
	for k, v := range e.nodes {
		nodes[k] = v
	}
	return nodes
}

// Config returns the network configuration
func (e *Environment) Config() *network.Network {
	return e.networkCfg
}

// checkNode validates and sets defaults for a node configuration
func (e *Environment) checkNode(nodeConfig node.Config) error {
	// use preset dirs if not defined
	if nodeConfig.GetConfigDir() == "" {
		nodeConfig.SetConfigDir("/home/thor")
	}
	if nodeConfig.GetDataDir() == "" {
		nodeConfig.SetDataDir("/home/thor")
	}
	if nodeConfig.GetExecArtifact() == "" {
		if e.dockerImage == "" {
			return fmt.Errorf("docker image is not set, please provide a valid docker image")
		}
		nodeConfig.SetExecArtifact(e.dockerImage)
	}
	return nil
}

// createNode creates a new docker node
func (e *Environment) createNode(nodeConfig node.Config, enodes []string) (node.Lifecycle, error) {
	// Get the already-assigned IP address (assigned during enode generation)
	ipAddr := e.ipManager.GetNodeIP(nodeConfig.GetID())
	if ipAddr == "" {
		// Fallback: assign IP if not already assigned (shouldn't happen in normal flow)
		var err error
		ipAddr, err = e.ipManager.NextIP(nodeConfig.GetID())
		if err != nil {
			return nil, err
		}
	}

	// Get exposed port for this node
	exposedPort := e.exposedPorts[nodeConfig.GetID()]
	if exposedPort == nil {
		return nil, fmt.Errorf("no exposed port configured for node %s", nodeConfig.GetID())
	}

	return NewDockerNode(nodeConfig, enodes, e.networkID, exposedPort, ipAddr), nil
}

// buildDockerImageIfNeeded builds the docker image once if ThorBuilder is configured
func (e *Environment) buildDockerImageIfNeeded() error {
	if e.networkCfg.ThorBuilder == nil {
		return nil // No thor builder configured
	}

	builder := thorbuilder.New(e.networkCfg.ThorBuilder)
	if err := builder.Download(); err != nil {
		return err
	}

	dockerImage, err := builder.BuildDockerImage()
	if err != nil {
		return fmt.Errorf("failed to build thor binary - %w", err)
	}

	e.dockerImage = dockerImage
	return nil
}

// generateEnodes creates enode strings for all nodes (excluding public network nodes)
func (e *Environment) generateEnodes() ([]string, error) {
	var enodes []string
	for _, node := range e.networkCfg.Nodes {
		// Skip enode generation for public network nodes (testnet/mainnet)
		if isPublicNetworkNode(node) {
			continue
		}

		// Pre-allocate IP address if not already assigned
		ipAddr := e.ipManager.GetNodeIP(node.GetID())
		if ipAddr == "" {
			// Assign IP address now for enode generation
			var err error
			ipAddr, err = e.ipManager.NextIP(node.GetID())
			if err != nil {
				return nil, fmt.Errorf("failed to allocate IP for node %s: %w", node.GetID(), err)
			}
		}

		enode, err := node.Enode(ipAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to get enode for node %s: %w", node.GetID(), err)
		}
		enodes = append(enodes, enode)
	}
	return enodes, nil
}

// checkOrCreateNetwork creates or verifies the Docker network exists
func (e *Environment) checkOrCreateNetwork(networkName, subnet string) error {
	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("could not create Docker client: %v", err)
	}
	defer cli.Close()

	// List existing networks
	networks, err := cli.NetworkList(context.Background(), dockertypes.NetworkListOptions{})
	if err != nil {
		return fmt.Errorf("could not list Docker networks: %v", err)
	}

	// Check if the network already exists
	for _, net := range networks {
		if net.Name == networkName {
			slog.Info("Network already exists", "networkName", networkName)
			err := cli.NetworkRemove(context.Background(), networkName)
			if err != nil {
				return err
			}
			slog.Info("Removed existing network", "networkName", networkName)
		}
	}

	// Network does not exist, create it
	// Define the network configuration
	networkCreate := dockertypes.NetworkCreate{
		Driver: "bridge",
		IPAM: &dockernetwork.IPAM{
			Driver: "default",
			Config: []dockernetwork.IPAMConfig{
				{
					Subnet: subnet,
				},
			},
		},
	}
	_, err = cli.NetworkCreate(context.Background(), networkName, networkCreate)
	if err != nil {
		return fmt.Errorf("could not create Docker network: %v", err)
	}

	slog.Info("Network created", "networkName", networkName)
	return nil
}

// Helper functions and types

// ExposedPort represents a port mapping between host and container
type ExposedPort struct {
	HostPort      string
	ContainerPort string
}

// isPublicNetworkNode checks if a node is configured for a public network (testnet/mainnet)
func isPublicNetworkNode(node node.Config) bool {
	networkArg, exists := node.GetAdditionalArgs()["network"]
	return exists && (networkArg == "test" || networkArg == "main")
}