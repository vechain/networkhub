package preset

import (
	"fmt"

	"github.com/vechain/networkhub/internal/environments"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/thorbuilder"
)

// NewPublicNetwork creates a network configuration for connecting to VeChain public networks.
// networkType must be environments.ThorNetworkTest for testnet or environments.ThorNetworkMain for mainnet.
// branch specifies the thor repository branch to use (defaults to "master" if empty).
func NewPublicNetwork(networkType, branch string) (*network.Network, error) {
	if networkType != environments.ThorNetworkTest && networkType != environments.ThorNetworkMain {
		return nil, fmt.Errorf("invalid network type: %s. Must be '%s' or '%s'", networkType, environments.ThorNetworkTest, environments.ThorNetworkMain)
	}

	baseID := network.Testnet
	if networkType == environments.ThorNetworkMain {
		baseID = network.Mainnet
	}

	thorBranch := "master"
	if branch != "" {
		thorBranch = branch
	}

	return &network.Network{
		BaseID:      baseID,
		Environment: environments.Local, // Use local environment for public networks
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
	return NewPublicNetwork(environments.ThorNetworkTest, "")
}

// NewMainnetNetwork creates a network configuration for connecting to VeChain mainnet.
func NewMainnetNetwork() (*network.Network, error) {
	return NewPublicNetwork(environments.ThorNetworkMain, "")
}
