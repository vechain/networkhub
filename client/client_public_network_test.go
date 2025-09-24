package client

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/preset"
	"github.com/vechain/thor/v2/thorclient"
)

func TestMainnetConnection(t *testing.T) {
	// Create mainnet network and client
	network, err := preset.NewMainnetNetwork()
	require.NoError(t, err)
	require.Equal(t, "mainnet", network.BaseID)

	c, err := NewWithNetwork(network)
	require.NoError(t, err)

	// Create a node to connect to mainnet
	apiPort, p2pPort := randomPorts()
	mainnetNode := &node.BaseNode{
		ID:             "mainnet-test-node",
		P2PListenPort:  p2pPort,
		APIAddr:        fmt.Sprintf("127.0.0.1:%d", apiPort),
		AdditionalArgs: map[string]string{"network": "main"},
	}

	// Add the node to the network
	err = c.AddNode(mainnetNode)
	require.NoError(t, err)

	// Cleanup
	defer func() {
		if err := c.Stop(); err != nil {
			t.Logf("Warning: failed to stop client: %v", err)
		}
	}()

	// Wait for node to start and begin syncing
	t.Log("Waiting for mainnet node to start...")
	time.Sleep(15 * time.Second)

	// Get the best block from local node
	localClient := thorclient.New(fmt.Sprintf("http://127.0.0.1:%d", apiPort))
	localBlock, err := localClient.Block("best")
	require.NoError(t, err)
	require.NotNil(t, localBlock)

	t.Logf("Local mainnet node best block: #%d %s", localBlock.Number, localBlock.ID)

	// Query the public mainnet API for the same block
	publicClient := thorclient.New("https://mainnet.vechain.org")
	publicBlock, err := publicClient.Block(localBlock.ID.String())
	require.NoError(t, err)
	require.NotNil(t, publicBlock)

	// Compare key fields
	require.Equal(t, localBlock.ID, publicBlock.ID, "Block IDs should match")
	require.Equal(t, localBlock.Number, publicBlock.Number, "Block numbers should match")
	require.Equal(t, localBlock.ParentID, publicBlock.ParentID, "Parent IDs should match")
	require.Equal(t, localBlock.Timestamp, publicBlock.Timestamp, "Timestamps should match")

	t.Logf("Successfully verified mainnet block #%d matches public API", localBlock.Number)
}

func TestTestnetConnection(t *testing.T) {
	// Create testnet network and client
	network, err := preset.NewTestnetNetwork()
	require.NoError(t, err)
	require.Equal(t, "testnet", network.BaseID)

	c, err := NewWithNetwork(network)
	require.NoError(t, err)

	// Create a node to connect to testnet
	apiPort, p2pPort := randomPorts()
	testnetNode := &node.BaseNode{
		ID:             "testnet-test-node",
		P2PListenPort:  p2pPort,
		APIAddr:        fmt.Sprintf("127.0.0.1:%d", apiPort),
		AdditionalArgs: map[string]string{"network": "test"},
	}

	// Add the node to the network
	err = c.AddNode(testnetNode)
	require.NoError(t, err)

	// Cleanup
	defer func() {
		if err := c.Stop(); err != nil {
			t.Logf("Warning: failed to stop client: %v", err)
		}
	}()

	// Wait for node to start and begin syncing
	t.Log("Waiting for testnet node to start...")
	time.Sleep(15 * time.Second)

	// Get the best block from local node
	localClient := thorclient.New(fmt.Sprintf("http://127.0.0.1:%d", apiPort))
	localBlock, err := localClient.Block("best")
	require.NoError(t, err)
	require.NotNil(t, localBlock)

	t.Logf("Local testnet node best block: #%d %s", localBlock.Number, localBlock.ID)

	// Query the public testnet API for the same block
	publicClient := thorclient.New("https://testnet.vechain.org")
	publicBlock, err := publicClient.Block(localBlock.ID.String())
	require.NoError(t, err)
	require.NotNil(t, publicBlock)

	// Compare key fields
	require.Equal(t, localBlock.ID, publicBlock.ID, "Block IDs should match")
	require.Equal(t, localBlock.Number, publicBlock.Number, "Block numbers should match")
	require.Equal(t, localBlock.ParentID, publicBlock.ParentID, "Parent IDs should match")
	require.Equal(t, localBlock.Timestamp, publicBlock.Timestamp, "Timestamps should match")

	t.Logf("Successfully verified testnet block #%d matches public API", localBlock.Number)
}

// randomPorts generates two random port numbers for API and P2P
func randomPorts() (int, int) {
	// Generate random ports in range 9400-9500 to avoid conflicts
	apiPort := randomPortInRange(9400, 9450)
	p2pPort := randomPortInRange(9451, 9500)
	return apiPort, p2pPort
}

// randomPortInRange generates a random port number in the given range
func randomPortInRange(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	return min + int(n.Int64())
}

func TestNewPublicNetworkErrors(t *testing.T) {
	// Test invalid network type
	_, err := preset.NewPublicNetwork("invalid", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid network type: invalid")

	// Test empty network type
	_, err = preset.NewPublicNetwork("", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid network type")

	// Test case sensitivity
	_, err = preset.NewPublicNetwork("Test", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid network type: Test")

	_, err = preset.NewPublicNetwork("Main", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid network type: Main")
}

func TestPublicNetworkConfiguration(t *testing.T) {
	// Test testnet configuration
	testnet, err := preset.NewTestnetNetwork()
	require.NoError(t, err)
	require.Equal(t, "testnet", testnet.BaseID)
	require.Equal(t, "local", testnet.Environment) // Uses local environment
	require.NotNil(t, testnet.ThorBuilder)
	require.Equal(t, "https://github.com/vechain/thor", testnet.ThorBuilder.DownloadConfig.RepoUrl)
	require.Equal(t, "master", testnet.ThorBuilder.DownloadConfig.Branch)
	require.True(t, testnet.ThorBuilder.DownloadConfig.IsReusable)
	require.Len(t, testnet.Nodes, 0) // Should start with no nodes

	// Test mainnet configuration
	mainnet, err := preset.NewMainnetNetwork()
	require.NoError(t, err)
	require.Equal(t, "mainnet", mainnet.BaseID)
	require.Equal(t, "local", mainnet.Environment) // Uses local environment
	require.NotNil(t, mainnet.ThorBuilder)
	require.Equal(t, "https://github.com/vechain/thor", mainnet.ThorBuilder.DownloadConfig.RepoUrl)
	require.Equal(t, "master", mainnet.ThorBuilder.DownloadConfig.Branch)
	require.True(t, mainnet.ThorBuilder.DownloadConfig.IsReusable)
	require.Len(t, mainnet.Nodes, 0) // Should start with no nodes

	// Test custom branch
	customNet, err := preset.NewPublicNetwork("test", "custom-branch")
	require.NoError(t, err)
	require.Equal(t, "custom-branch", customNet.ThorBuilder.DownloadConfig.Branch)
}
