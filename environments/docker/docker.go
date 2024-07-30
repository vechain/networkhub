package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/docker/docker/client"
	"github.com/vechain/networkhub/environments"
	"github.com/vechain/networkhub/network"

	dockernetwork "github.com/docker/docker/api/types/network"
)

type Docker struct {
	dockerNodes  map[string]*Node
	networkCfg   *network.Network
	id           string
	networkID    string
	exposedPorts map[string]*exposedPort
	ipManager    *IpManager
}

func NewDockerEnv() environments.Actions {
	return &Docker{
		dockerNodes:  map[string]*Node{},
		exposedPorts: map[string]*exposedPort{},
		ipManager:    NewIPManagerRandom(),
	}
}

func (d *Docker) LoadConfig(cfg *network.Network) (string, error) {
	d.networkCfg = cfg
	d.id = d.networkCfg.Environment + d.networkCfg.ID
	d.networkID = d.id + "-network"

	for i, node := range cfg.Nodes {
		// use preset dirs if not defined
		if node.GetConfigDir() == "" {
			node.SetConfigDir("/home/thor")
		}
		if node.GetDataDir() == "" {
			node.SetDataDir("/home/thor")
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

func (d *Docker) StopNetwork() error {
	for s, dockerNode := range d.dockerNodes {
		err := dockerNode.Stop()
		if err != nil {
			return fmt.Errorf("unable to stop node %s - %w", s, err)
		}
	}
	return nil
}

func (d *Docker) Info() error {
	//TODO implement me
	panic("implement me")
}

func (d *Docker) checkOrCreateNetwork(networkName, subnet string) error {
	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("could not create Docker client: %v", err)
	}
	defer cli.Close()

	// List existing networks
	networks, err := cli.NetworkList(context.Background(), dockernetwork.ListOptions{})
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
	networkCreate := dockernetwork.CreateOptions{
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
