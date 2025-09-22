package preset

import (
	"fmt"

	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/thorbuilder"
)

// NewPublicNetwork creates a network configuration for connecting to VeChain public networks.
// networkType must be "test" for testnet or "main" for mainnet.
// branch specifies the thor repository branch to use (defaults to "master" if empty).
func NewPublicNetwork(networkType, branch string) (*network.Network, error) {
	if networkType != "test" && networkType != "main" {
		return nil, fmt.Errorf("invalid network type: %s. Must be 'test' or 'main'", networkType)
	}

	baseID := "testnet"
	if networkType == "main" {
		baseID = "mainnet"
	}

	thorBranch := "master"
	if branch != "" {
		thorBranch = branch
	}

	return &network.Network{
		BaseID:      baseID,
		Environment: "local", // Use local environment for public networks
		Nodes:       []node.Config{},
		ThorBuilder: &thorbuilder.Config{
			DownloadConfig: &thorbuilder.DownloadConfig{
				RepoUrl:    "https://github.com/vechain/thor",
				Branch:     thorBranch,
				IsReusable: true,
			},
		},
	}, nil
}

// NewTestnetNetwork creates a network configuration for connecting to VeChain testnet.
func NewTestnetNetwork() (*network.Network, error) {
	return NewPublicNetwork("test", "")
}

// NewMainnetNetwork creates a network configuration for connecting to VeChain mainnet.
func NewMainnetNetwork() (*network.Network, error) {
	return NewPublicNetwork("main", "")
}