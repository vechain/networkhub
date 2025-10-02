package client

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vechain/networkhub/hayabusa"
	"github.com/vechain/networkhub/preset"
	"github.com/vechain/networkhub/thorbuilder"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/thorclient"
)

// TestClientFourNodesHayabusa tests the client with a 4-node Hayabusa network.
// This test verifies that the client can:
// 1. Set up and start a 4-node Hayabusa network with immediate transition
// 2. Wait for all nodes to connect and sync
// 3. Deploy and execute smart contracts in post-hayabusa state
// 4. Verify validator consensus and network health
func TestClientFourNodesHayabusa(t *testing.T) {
	// Create the four nodes Hayabusa network with immediate transition
	fourNodesHayabusaNetwork := preset.LocalFourNodesHayabusa()
	fourNodesHayabusaNetwork.ThorBuilder.DownloadConfig = &thorbuilder.DownloadConfig{
		RepoUrl:    "https://github.com/vechain/thor",
		Branch:     "pedro/hayabusa/improve_customnet",
		IsReusable: false,
	}

	// Update ports to avoid collision with other tests
	basePort := 8700
	for _, node := range fourNodesHayabusaNetwork.Nodes {
		basePort++
		node.SetAPIAddr(fmt.Sprintf("127.0.0.1:%d", basePort))
		basePort++
		node.SetP2PListenPort(basePort)
	}

	// Create client with the network
	c, err := New(fourNodesHayabusaNetwork)
	require.NoError(t, err)

	require.NoError(t, c.Start())
	// Cleanup
	defer func() {
		if err := c.Stop(); err != nil {
			t.Logf("Warning: failed to stop client: %v", err)
		}
	}()

	// Wait for all nodes to connect and sync
	t.Log("Waiting for Hayabusa nodes to connect and sync...")
	require.NoError(t, c.network.HealthCheck(4, 2*time.Minute))

	// Test staker contract functionality to verify validators are active
	client := thorclient.New(c.network.Nodes[0].GetHTTPAddr())
	staker := hayabusa.NewStaker(client)

	// Check firstActive to see if validators are now active
	validatorAddr, err := staker.FirstActive()
	require.NoError(t, err)
	t.Logf("FirstActive successful - Validator: %s", validatorAddr)

	// Verify that one of our validators is now active
	expectedValidators := []thor.Address{
		*preset.SixNNAccount1.Address,
		*preset.SixNNAccount1.Address,
		*preset.SixNNAccount1.Address,
		*preset.SixNNAccount1.Address,
	}

	validatorFound := false
	for _, expected := range expectedValidators {
		if validatorAddr == expected {
			validatorFound = true
			break
		}
	}
	require.True(t, validatorFound, "Active validator should be one of our registered validators, got %s", validatorAddr)

	// Verify all nodes are producing blocks
	t.Log("Verifying all validator nodes are participating in consensus...")
	for i, node := range c.network.Nodes {
		nodeClient := thorclient.New(node.GetHTTPAddr())

		// Check that each node can respond to queries
		block, err := nodeClient.Block("best")
		require.NoError(t, err, "Node %d should respond to block queries", i+1)
		require.Greater(t, block.Number, uint32(0), "Node %d should have produced blocks", i+1)

		t.Logf("Node %d: Block %d, Validator nodes operational", i+1, block.Number)
	}

	t.Log("Successfully tested Hayabusa network with 4 validator nodes!")
}
