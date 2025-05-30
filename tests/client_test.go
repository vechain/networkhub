package tests

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vechain/networkhub/entrypoint/client"
	"github.com/vechain/networkhub/preset"
	"github.com/vechain/networkhub/thorbuilder"
	"github.com/vechain/networkhub/utils/common"
	"github.com/vechain/networkhub/utils/datagen"
	thorgenesis "github.com/vechain/thor/v2/genesis"
	"github.com/vechain/thor/v2/thorclient"
)

func TestLocalClient(t *testing.T) {
	// Create client
	c := client.New()

	// Create preset networks
	networkCfg := preset.LocalThreeMasterNodesNetwork()
	basePort := 9100 // avoid port collision with other tests

	// configure local artifacts
	cfg := thorbuilder.DefaultConfig()
	cfg.DownloadConfig.IsReusable = false
	networkCfg.ThorBuilder = cfg

	// modify genesis
	prefundedAcc := datagen.RandAccount().Address
	for _, node := range networkCfg.Nodes {
		nodeGenesis := node.GetGenesis()
		nodeGenesis.Accounts = append(
			nodeGenesis.Accounts,
			thorgenesis.Account{
				Address: *prefundedAcc,
				Balance: (*thorgenesis.HexOrDecimal256)(preset.LargeBigValue),
			})
		node.SetGenesis(nodeGenesis)
		basePort++
		node.SetAPIAddr(fmt.Sprintf("0.0.0.0:%d", basePort))
		basePort++
		node.SetP2PListenPort(basePort)
	}

	// Configure and start network
	net, err := c.Config(networkCfg)
	if err != nil {
		t.Fatalf("Failed to configure network: %v", err)
	}

	// Start network
	if err := c.Start(net.ID()); err != nil {
		t.Fatalf("Failed to start network: %v", err)
	}

	require.NoError(t,
		common.Retry(
			func() error {
				_, err := thorclient.New(networkCfg.Nodes[0].GetHTTPAddr()).Block("best")
				return err
			}, time.Second, 60),
	)

	account, err := thorclient.New(networkCfg.Nodes[0].GetHTTPAddr()).Account(prefundedAcc)
	require.NoError(t, err)
	bal := big.Int(account.Balance)
	require.Equal(t, bal.Cmp(big.NewInt(0)), 1)

	// Stop network
	time.Sleep(5 * time.Second)
	if err := c.Stop(net.ID()); err != nil {
		t.Fatalf("Failed to stop network: %v", err)
	}
}

func TestDockerClient(t *testing.T) {
	// Create client
	c := client.New()

	// Create preset networks
	networkCfg := preset.LocalThreeMasterNodesNetwork()

	// Modify for docker usage
	networkCfg.Environment = "docker"
	dockerImage := "vechain/thor"
	basePort := 9000 // avoid port collision with other tests

	prefundedAcc := datagen.RandAccount().Address
	for i, node := range networkCfg.Nodes {
		// modify genesis
		nodeGenesis := node.GetGenesis()
		nodeGenesis.Accounts = append(
			nodeGenesis.Accounts,
			thorgenesis.Account{
				Address: *prefundedAcc,
				Balance: (*thorgenesis.HexOrDecimal256)(preset.LargeBigValue),
			})
		node.SetGenesis(nodeGenesis)

		// modify node start
		node.SetExecArtifact(dockerImage)
		basePort++
		node.SetAPIAddr(fmt.Sprintf("0.0.0.0:%d", basePort))
		basePort++
		node.SetP2PListenPort(basePort)
		node.SetID(fmt.Sprintf("%s-%d", node.GetID(), i))
	}

	// Configure and start network
	net, err := c.Config(networkCfg)
	require.NoError(t, err)

	// Start network
	if err := c.Start(net.ID()); err != nil {
		t.Fatalf("Failed to start network: %v", err)
	}

	require.NoError(t,
		common.Retry(
			func() error {
				_, err := thorclient.New(networkCfg.Nodes[0].GetHTTPAddr()).Block("best")
				return err
			}, time.Second, 60),
	)

	account, err := thorclient.New(networkCfg.Nodes[0].GetHTTPAddr()).Account(prefundedAcc)
	require.NoError(t, err)
	bal := big.Int(account.Balance)
	require.Equal(t, bal.Cmp(big.NewInt(0)), 1)

	// Stop network
	if err := c.Stop(net.ID()); err != nil {
		t.Fatalf("Failed to stop network: %v", err)
	}
}
