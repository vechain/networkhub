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
	"github.com/vechain/networkhub/environments"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
)

type Docker struct {
	dockerNodes  map[string]*Node
	networkCfg   *network.Network
	id           string
	networkID    string
	exposedPorts map[string]*exposedPort
	ipManager    *IpManager
	started      bool
	mu           sync.Mutex
}

func NewDockerEnv() *Docker {
	return &Docker{
		dockerNodes:  map[string]*Node{},
		exposedPorts: map[string]*exposedPort{},
		ipManager:    NewIPManagerRandom(),
	}
}

var _ environments.Actions = (*Docker)(nil)

func (d *Docker) LoadConfig(cfg *network.Network) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.networkCfg = cfg
	d.id = d.networkCfg.ID()
	d.networkID = d.id + "-network"

	for i, node := range cfg.Nodes {
		// use preset dirs if not defined
		if node.GetConfigDir() == "" {
			node.SetConfigDir("/home/thor")
		}
		if node.GetDataDir() == "" {
			node.SetDataDir("/home/thor")
		}
		// initial ip addr config, in case `AttachNode` is called later
		_, err := d.ipManager.NextIP(node.GetID())
		if err != nil {
			return "", fmt.Errorf("unable to get next ip for node %s - %w", node.GetID(), err)
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
	d.mu.Lock()
	defer d.mu.Unlock()

	// create a network for fixed ip addresses (enodes cannot have dns names)
	if err := d.checkOrCreateNetwork(d.networkID, d.ipManager.Subnet()); err != nil {
		return fmt.Errorf("unable to create network: %w", err)
	}

	for _, nodeCfg := range d.networkCfg.Nodes {
		enodes, err := d.enodes(nodeCfg)
		if err != nil {
			return fmt.Errorf("failed to get enodes - %w", err)
		}
		ipAddr, err := d.ipManager.NextIP(nodeCfg.GetID())
		if err != nil {
			return fmt.Errorf("unable to get next ip for node %s - %w", nodeCfg.GetID(), err)
		}
		dockerNode := NewDockerNode(nodeCfg, enodes, d.networkID, d.exposedPorts[nodeCfg.GetID()], ipAddr)
		if err := dockerNode.Start(); err != nil {
			return fmt.Errorf("unable to start node - %w", err)
		}

		d.dockerNodes[nodeCfg.GetID()] = dockerNode
	}

	d.started = true

	return nil
}

func (d *Docker) Nodes() map[string]node.Lifecycle {
	d.mu.Lock()
	defer d.mu.Unlock()

	nodes := make(map[string]node.Lifecycle, len(d.dockerNodes))
	for k, v := range d.dockerNodes {
		nodes[k] = v
	}
	return nodes
}

func (d *Docker) StopNetwork() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for s, dockerNode := range d.dockerNodes {
		err := dockerNode.Stop()
		if err != nil {
			return fmt.Errorf("unable to stop node %s - %w", s, err)
		}
	}

	d.started = false
	return nil
}

func (d *Docker) AttachNode(n *node.Config) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.dockerNodes[n.GetID()]; exists {
		return fmt.Errorf("node with ID %s already exists", n.GetID())
	}

	enodes, err := d.enodes(n)
	if err != nil {
		return fmt.Errorf("failed to get enodes - %w", err)
	}
	ipAddr, err := d.ipManager.NextIP(n.GetID())
	if err != nil {
		return fmt.Errorf("unable to get next ip for node %s - %w", n.GetID(), err)
	}
	dockerNode := NewDockerNode(n, enodes, d.networkID, d.exposedPorts[n.GetID()], ipAddr)

	split := strings.Split(n.GetAPIAddr(), ":")
	if len(split) != 2 {
		return fmt.Errorf("unable to parse API Addr")
	}
	exposedAPIPort, err := strconv.Atoi(split[1])
	if err != nil {
		return fmt.Errorf("unable to parse API port: %w", err)
	}
	d.exposedPorts[n.GetID()] = &exposedPort{
		hostPort:      fmt.Sprintf("%d", exposedAPIPort+len(d.dockerNodes)),
		containerPort: split[1],
	}

	d.dockerNodes[n.GetID()] = dockerNode

	if d.started {
		// If the network has already started, we need to start the node immediately.
		if err := dockerNode.Start(); err != nil {
			return fmt.Errorf("unable to start node %s after attaching - %w", n.GetID(), err)
		}
	} else {
		slog.Info("Network not started yet, node will be started when network starts", "nodeID", n.GetID())
	}

	return nil
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

func (d *Docker) enodes(exclude *node.Config) ([]string, error) {
	var enodes []string
	for _, node := range d.networkCfg.Nodes {
		if node.GetID() == exclude.GetID() {
			continue // skip the node that is being configured
		}
		enode, err := node.Enode(d.ipManager.GetNodeIP(node.GetID()))
		if err != nil {
			return nil, err
		}
		enodes = append(enodes, enode)
	}

	return enodes, nil
}

type exposedPort struct {
	hostPort      string
	containerPort string
}
