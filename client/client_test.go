package client

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vechain/networkhub/internal/environments"
	"github.com/vechain/networkhub/preset"
	"github.com/vechain/networkhub/thorbuilder"
	"github.com/vechain/networkhub/utils/common"
	"github.com/vechain/networkhub/utils/datagen"
	thorgenesis "github.com/vechain/thor/v2/genesis"
	"github.com/vechain/thor/v2/thorclient"
)

func TestLocalClient(t *testing.T) {
	// Create preset networks
	networkCfg := preset.LocalThreeNodesNetwork()
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

	// Create client with network configuration
	c, err := New(networkCfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
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
	bal := (*big.Int)(account.Balance)
	require.Equal(t, bal.Cmp(big.NewInt(0)), 1)

	// Stop network
	time.Sleep(5 * time.Second)
	if err := c.Stop(); err != nil {
		t.Fatalf("Failed to stop network: %v", err)
	}
}

func TestDockerClient(t *testing.T) {
	// Create preset networks
	networkCfg := preset.LocalThreeNodesNetwork()

	// Modify for docker usage
	networkCfg.Environment = environments.Docker
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

	// Create client with network configuration
	c, err := New(networkCfg)
	require.NoError(t, err)

	// Start network
	if err := c.Start(); err != nil {
		t.Fatalf("Failed to start network: %v", err)
	}

	network, err := c.GetNetwork()
	require.NoError(t, err)

	require.NoError(t, network.HealthCheck(3, time.Minute))

	account, err := thorclient.New(networkCfg.Nodes[0].GetHTTPAddr()).Account(prefundedAcc)
	require.NoError(t, err)
	bal := (*big.Int)(account.Balance)
	require.Equal(t, bal.Cmp(big.NewInt(0)), 1)

	// Stop network
	if err := c.Stop(); err != nil {
		t.Fatalf("Failed to stop network: %v", err)
	}
}

func TestAddRemoveNodes(t *testing.T) {
	// Create initial network with 2 nodes
	networkCfg := preset.LocalThreeNodesNetwork()
	basePort := 9400 // avoid port collision with other tests

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

	// Create client with network configuration
	c, err := New(networkCfg)
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

	// Test adding node with invalid network configuration
	// Create a minimal invalid network config for testing error conditions
	invalidNetworkCfg := preset.LocalThreeNodesNetwork()
	invalidNetworkCfg.Nodes = nil // empty nodes
	c2, err2 := New(invalidNetworkCfg)
	require.NoError(t, err2) // Constructor should succeed
	err = c2.AddNode(thirdNode)
	require.NoError(t, err) // Should be able to add node even to initially empty network

	// Stop the original network
	err = c.Stop()
	require.NoError(t, err)
}

func TestClientAdditionalArgs(t *testing.T) {
	// Create network with LocalThreeMasterNodes preset
	networkCfg := preset.LocalThreeNodesNetwork()

	// Configure thor builder
	cfg := thorbuilder.DefaultConfig()
	cfg.DownloadConfig.IsReusable = false
	networkCfg.ThorBuilder = cfg

	// Update ports to avoid conflicts with other tests
	basePort := 8600 // Different range from other tests
	for _, node := range networkCfg.Nodes {
		basePort++
		node.SetAPIAddr(fmt.Sprintf("127.0.0.1:%d", basePort))
		basePort++
		node.SetP2PListenPort(basePort)

		// Set additional args for all nodes to enable call tracer
		node.SetAdditionalArgs(map[string]string{
			"api-allowed-tracers": "call",
		})
	}

	// Create client with the network
	c, err := New(networkCfg)
	require.NoError(t, err)
	require.NoError(t, c.Start())

	// Cleanup
	defer func() {
		if err := c.Stop(); err != nil {
			t.Logf("Warning: failed to stop client: %v", err)
		}
	}()

	// Wait for first node to be accessible
	client := thorclient.New(networkCfg.Nodes[0].GetHTTPAddr())
	t.Logf("Waiting for node at %s to be ready...", networkCfg.Nodes[0].GetHTTPAddr())

	require.NoError(t, common.Retry(func() error {
		_, err := client.Block("best")
		if err != nil {
			t.Logf("Still waiting for node: %v", err)
		}
		return err
	}, time.Second, 60))

	t.Log("Node is ready, testing debug tracer API...")

	// Test the additional args by making a debug tracer API call
	res, statusCode, err := client.RawHTTPClient().RawHTTPPost("/debug/tracers/call", []byte(`{
  "value": "0x0",
  "to": "0x0000000000000000000000000000456E65726779",
  "data": "0xa9059cbb0000000000000000000000000f872421dc479f3c11edd89512731814d0598db50000000000",
  "caller": "0x7567d83b7b8d80addcb281a71d54fc7b3364ffed",
  "gasPayer": "0xd3ae78222beadb038203be21ed5ce7c9b1bff602",
  "name": "call"
}`))
	require.NoError(t, err)
	require.Equal(t, 200, statusCode)

	// Verify we got a valid response
	body := string(res)
	require.NotEmpty(t, body)
	t.Logf("Successfully called debug tracer API with response: %s", body)
}
