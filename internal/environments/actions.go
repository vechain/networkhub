package environments

import (
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
)

type Actions interface {
	StartNetwork() error
	StopNetwork() error
	Nodes() map[string]node.Lifecycle
	Config() *network.Network
	AddNode(nodeConfig node.Config) error
	RemoveNode(nodeID string) error
}

// Environment types define how nodes are executed and managed
const (
	// Local environment runs nodes as local processes
	Local = "local"

	// Docker environment runs nodes in Docker containers
	Docker = "docker"
)

// Thor binary network arguments used when executing thor nodes
const (
	// ThorNetworkMain is the thor --network argument for mainnet
	ThorNetworkMain = "main"

	// ThorNetworkTest is the thor --network argument for testnet
	ThorNetworkTest = "test"
)
