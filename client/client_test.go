package client

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vechain/networkhub/preset"
	"github.com/vechain/networkhub/thorbuilder"
	"github.com/vechain/networkhub/utils/common"
	"github.com/vechain/networkhub/utils/datagen"
	thorgenesis "github.com/vechain/thor/v2/genesis"
	"github.com/vechain/thor/v2/thorclient"
)

func TestLocalClient(t *testing.T) {
	// Create client
	c := New()

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
	err := c.LoadNetwork(networkCfg)
	if err != nil {
		t.Fatalf("Failed to load network: %v", err)
	}

	// Start network
	if err := c.Start(); err != nil {
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
	if err := c.Stop(); err != nil {
		t.Fatalf("Failed to stop network: %v", err)
	}
}

func TestDockerClient(t *testing.T) {
	// Create client
	c := New()

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
	err := c.LoadNetwork(networkCfg)
	require.NoError(t, err)

	// Start network
	if err := c.Start(); err != nil {
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
	if err := c.Stop(); err != nil {
		t.Fatalf("Failed to stop network: %v", err)
	}
}

func TestAddRemoveNodes(t *testing.T) {
	// Create client
	c := New()

	// Create initial network with 2 nodes
	networkCfg := preset.LocalThreeMasterNodesNetwork()
	basePort := 9300 // avoid port collision with other tests

	// Configure thor builder
	cfg := thorbuilder.DefaultConfig()
	cfg.DownloadConfig.IsReusable = false
	networkCfg.ThorBuilder = cfg

	// Keep only 2 nodes initially
	originalNodes := networkCfg.Nodes[:2]
	thirdNode := networkCfg.Nodes[2] // save for later addition

	for i, node := range originalNodes {
		basePort++
		node.SetAPIAddr(fmt.Sprintf("0.0.0.0:%d", basePort))
		basePort++
		node.SetP2PListenPort(basePort)
		node.SetID(fmt.Sprintf("node-%d", i))
	}
	networkCfg.Nodes = originalNodes

	// Load and start network with 2 nodes
	err := c.LoadNetwork(networkCfg)
	require.NoError(t, err)

	err = c.Start()
	require.NoError(t, err)

	// Wait for network to be ready
	time.Sleep(10 * time.Second)

	// Verify we have 2 nodes initially
	network, err := c.GetNetwork()
	require.NoError(t, err)
	require.Len(t, network.Nodes, 2)

	// Verify nodes are running
	nodes, err := c.Nodes()
	require.NoError(t, err)
	require.Len(t, nodes, 2)

	// verify the network health
	require.NoError(t, network.HealthCheck(3, 60*time.Second))

	// Add third node to running network
	basePort++
	thirdNode.SetAPIAddr(fmt.Sprintf("0.0.0.0:%d", basePort))
	basePort++
	thirdNode.SetP2PListenPort(basePort)
	thirdNode.SetID("node-3")

	err = c.AddNode(thirdNode)
	require.NoError(t, err)

	// Verify we have 3 nodes in configuration
	network, err = c.GetNetwork()
	require.NoError(t, err)
	require.Len(t, network.Nodes, 3)

	// Verify we have 3 nodes in config
	nodes, err = c.Nodes()
	require.NoError(t, err)

	// Wait for node to start
	time.Sleep(5 * time.Second)

	// Verify all 3 nodes are now running
	nodes, err = c.Nodes()
	require.NoError(t, err)
	require.Len(t, nodes, 3)

	// Check the third node exists and is running
	_, exists := nodes["node-3"]
	require.True(t, exists, "Third node should be running")

	// verify the network health
	require.NoError(t, network.HealthCheck(5, 60*time.Second))

	// Remove the third node from running network
	err = c.RemoveNode("node-3")
	require.NoError(t, err)

	// Verify we're back to 2 nodes in configuration
	network, err = c.GetNetwork()
	require.NoError(t, err)
	require.Len(t, network.Nodes, 2)

	// Verify only 2 nodes are running
	nodes, err = c.Nodes()
	require.NoError(t, err)
	require.Len(t, nodes, 2)

	// verify the network health
	require.NoError(t, network.HealthCheck(7, 60*time.Second))

	// Verify the third node is no longer running
	_, exists = nodes["node-3"]
	require.False(t, exists, "Third node should not be running after removal")

	// Test removing non-existent node
	err = c.RemoveNode("non-existent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "node with ID non-existent does not exist")

	// Test adding node without network loaded
	c2 := New()
	err = c2.AddNode(thirdNode)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no network loaded")

	// Stop the original network
	err = c.Stop()
	require.NoError(t, err)
}
