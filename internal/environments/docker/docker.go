package docker

import (
	"context"
	"fmt"
	"github.com/vechain/networkhub/internal/environments"

	"log/slog"
	"strconv"
	"strings"

	dockertypes "github.com/docker/docker/api/types"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/thorbuilder"
)

type Docker struct {
	dockerNodes  map[string]*Node
	networkCfg   *network.Network
	id           string
	networkID    string
	exposedPorts map[string]*exposedPort
	ipManager    *IpManager
	dockerImage  string
}

type Factory struct{}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) New() environments.Actions {
	return NewEnv()
}

func NewEnv() *Docker {
	return &Docker{
		dockerNodes:  map[string]*Node{},
		exposedPorts: map[string]*exposedPort{},
		ipManager:    NewIPManagerRandom(),
	}
}

func (d *Docker) LoadConfig(cfg *network.Network) (string, error) {
	d.networkCfg = cfg
	d.id = d.networkCfg.ID()
	d.networkID = d.id + "-network"

	if cfg.ThorBuilder != nil {
		builder := thorbuilder.New(cfg.ThorBuilder)
		if err := builder.Download(); err != nil {
			return "", err
		}

		dockerImage, err := builder.BuildDockerImage()
		if err != nil {
			return "", fmt.Errorf("failed to build thor binary - %w", err)
		}

		d.dockerImage = dockerImage
	}

	for i, node := range cfg.Nodes {
		// use preset dirs if not defined
		if node.GetConfigDir() == "" {
			node.SetConfigDir("/home/thor")
		}
		if node.GetDataDir() == "" {
			node.SetDataDir("/home/thor")
		}
		if node.GetExecArtifact() == "" {
			if d.dockerImage == "" {
				return "", fmt.Errorf("docker image is not set, please provide a valid docker image")
			}
			node.SetExecArtifact(d.dockerImage)
		}

		// ensure API ports are exposed to the localhost
		split := strings.Split(node.GetAPIAddr(), ":")
		if len(split) != 2 {
			return "", fmt.Errorf("unable to parse API Addr")
		}

		exposedAPIPort, err := strconv.Atoi(split[1])
		if err != nil {
			return "", err
		}

		d.exposedPorts[node.GetID()] = &exposedPort{
			hostPort:      fmt.Sprintf("%d", exposedAPIPort+i),
			containerPort: split[1],
		}
	}

	return d.id, nil
}

func (d *Docker) StartNetwork() error {
	// create a network for fixed ip addresses (enodes cannot have dns names)
	if err := d.checkOrCreateNetwork(d.networkID, d.ipManager.Subnet()); err != nil {
		return fmt.Errorf("unable to create network: %w", err)
	}

	for _, nodeCfg := range d.networkCfg.Nodes {
		// calculate the node ip address
		nextIpAddr, err := d.ipManager.NextIP(nodeCfg.GetID())
		if err != nil {
			return err
		}

		// speed up p2p bootstrap
		var enodes []string
		for _, node := range d.networkCfg.Nodes {
			if node.GetID() == nodeCfg.GetID() {
				break
			}
			enode, err := node.Enode(d.ipManager.GetNodeIP(node.GetID()))
			if err != nil {
				return err
			}
			enodes = append(enodes, enode)
		}

		dockerNode := NewDockerNode(nodeCfg, enodes, d.networkID, d.exposedPorts[nodeCfg.GetID()], nextIpAddr)
		if err := dockerNode.Start(); err != nil {
			return fmt.Errorf("unable to start node - %w", err)
		}

		d.dockerNodes[nodeCfg.GetID()] = dockerNode
	}

	return nil
}

func (d *Docker) Nodes() map[string]node.Lifecycle {
	nodes := make(map[string]node.Lifecycle, len(d.dockerNodes))
	for k, v := range d.dockerNodes {
		nodes[k] = v
	}
	return nodes
}

func (d *Docker) StopNetwork() error {
	for s, dockerNode := range d.dockerNodes {
		err := dockerNode.Stop()
		if err != nil {
			return fmt.Errorf("unable to stop node %s - %w", s, err)
		}
	}
	return nil
}

// AddNode adds a node to the existing docker network.
// If the network has started, it will start the node.
func (d *Docker) AddNode(nodeConfig node.Config) error {
	if d.networkCfg == nil {
		return fmt.Errorf("network configuration is not loaded")
	}

	if _, exists := d.dockerNodes[nodeConfig.GetID()]; exists {
		return fmt.Errorf("node with ID %s already exists", nodeConfig.GetID())
	}

	// Check node configuration and set defaults
	if err := d.checkNode(nodeConfig); err != nil {
		return err
	}

	// Add node to network configuration
	d.networkCfg.Nodes = append(d.networkCfg.Nodes, nodeConfig)

	// Setup exposed ports for the new node
	split := strings.Split(nodeConfig.GetAPIAddr(), ":")
	if len(split) != 2 {
		return fmt.Errorf("unable to parse API Addr")
	}

	exposedAPIPort, err := strconv.Atoi(split[1])
	if err != nil {
		return err
	}

	// Use next available port number
	nextPortIndex := len(d.exposedPorts)
	d.exposedPorts[nodeConfig.GetID()] = &exposedPort{
		hostPort:      fmt.Sprintf("%d", exposedAPIPort+nextPortIndex),
		containerPort: split[1],
	}

	// If network is running, start the new node
	if len(d.dockerNodes) > 0 {
		// Calculate the node ip address
		nextIpAddr, err := d.ipManager.NextIP(nodeConfig.GetID())
		if err != nil {
			return err
		}

		// Get enodes from existing running nodes
		var enodes []string
		for _, node := range d.networkCfg.Nodes {
			if node.GetID() == nodeConfig.GetID() {
				continue
			}
			if d.ipManager.GetNodeIP(node.GetID()) != "" {
				enode, err := node.Enode(d.ipManager.GetNodeIP(node.GetID()))
				if err != nil {
					return err
				}
				enodes = append(enodes, enode)
			}
		}

		dockerNode := NewDockerNode(nodeConfig, enodes, d.networkID, d.exposedPorts[nodeConfig.GetID()], nextIpAddr)
		if err := dockerNode.Start(); err != nil {
			return fmt.Errorf("unable to start node %s - %w", nodeConfig.GetID(), err)
		}

		d.dockerNodes[nodeConfig.GetID()] = dockerNode
	}

	return nil
}

// RemoveNode removes a node from the docker network.
func (d *Docker) RemoveNode(nodeID string) error {
	if _, exists := d.dockerNodes[nodeID]; !exists {
		return fmt.Errorf("node with ID %s does not exist", nodeID)
	}

	// Stop and remove the docker container
	if err := d.dockerNodes[nodeID].Stop(); err != nil {
		return fmt.Errorf("unable to stop node %s - %w", nodeID, err)
	}

	// Remove from docker nodes map
	delete(d.dockerNodes, nodeID)

	// Remove from exposed ports
	delete(d.exposedPorts, nodeID)

	// Remove from network configuration
	var newNodes []node.Config
	for _, n := range d.networkCfg.Nodes {
		if n.GetID() != nodeID {
			newNodes = append(newNodes, n)
		}
	}
	d.networkCfg.Nodes = newNodes

	return nil
}

// checkNode validates and sets defaults for a docker node configuration
func (d *Docker) checkNode(nodeConfig node.Config) error {
	// use preset dirs if not defined
	if nodeConfig.GetConfigDir() == "" {
		nodeConfig.SetConfigDir("/home/thor")
	}
	if nodeConfig.GetDataDir() == "" {
		nodeConfig.SetDataDir("/home/thor")
	}
	if nodeConfig.GetExecArtifact() == "" {
		if d.dockerImage == "" {
			return fmt.Errorf("docker image is not set, please provide a valid docker image")
		}
		nodeConfig.SetExecArtifact(d.dockerImage)
	}
	return nil
}

func (d *Docker) Info() error {
	//TODO implement me
	panic("implement me")
}

func (d *Docker) Config() *network.Network {
	return d.networkCfg
}

func (d *Docker) checkOrCreateNetwork(networkName, subnet string) error {
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

type exposedPort struct {
	hostPort      string
	containerPort string
}
