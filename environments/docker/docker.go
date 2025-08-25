package docker

import (
	"context"
	"fmt"
	"os"

	"log/slog"
	"strconv"
	"strings"

	dockertypes "github.com/docker/docker/api/types"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/vechain/networkhub/environments"
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
	if err := CheckOrCreateNetwork(d.networkID, d.ipManager.Subnet()); err != nil {
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

		ports := d.exposedPorts[nodeCfg.GetID()]
		dockerNode := NewDockerNode(nodeCfg, enodes, d.networkID, ports.hostPort, ports.containerPort, nextIpAddr)
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

func (d *Docker) Info() error {
	//TODO implement me
	panic("implement me")
}

func (d *Docker) Config() *network.Network {
	return d.networkCfg
}

func CheckOrCreateNetwork(networkName, subnet string) error {
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

func GetDockerImageTag() string {
	env := os.Getenv("THOR_BRANCH")
	if env != "" {
		if env == "release/hayabusa" {
			return "ghcr.io/vechain/thor:release-hayabusa-latest"
		}
	}

	// Default to the latest tag if no specific branch is set
	return "vechain/thor:latest"
}
