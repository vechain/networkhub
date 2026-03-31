package client

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"github.com/vechain/networkhub/preset"
	"github.com/vechain/networkhub/thorbuilder"
	"github.com/vechain/networkhub/utils/common"
	"github.com/vechain/networkhub/utils/datagen"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/thorclient"
	"github.com/vechain/thor/v2/thorclient/builtin"
	"github.com/vechain/thor/v2/tx"
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
		Branch:     "master",
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
	require.NoError(t, c.network.HealthCheck(4, 4*time.Minute))

	// Test staker contract functionality to verify validators are active
	client := thorclient.New(c.network.Nodes[0].GetHTTPAddr())
	staker, err := builtin.NewStaker(client)
	require.NoError(t, err)

	// Check firstActive to see if validators are now active
	_, validatorAddr, err := staker.FirstActive()
	require.NoError(t, err)
	t.Logf("FirstActive successful - Validator: %s", validatorAddr)

	// Verify that one of our validators is now active
	expectedValidators := []thor.Address{
		*preset.SixNNAccount1.Address,
		*preset.SixNNAccount2.Address,
		*preset.SixNNAccount3.Address,
		*preset.SixNNAccount4.Address,
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

	// Deploy and test Shanghai contract to verify Galactica compatibility
	t.Log("Deploying Shanghai contract to test EVM compatibility...")
	deployAndAssertShanghaiContract(t, thorclient.New(c.network.Nodes[0].GetHTTPAddr()), preset.SixNNAccount1)

	t.Log("Successfully tested Hayabusa network with 4 validator nodes!")
}

func deployAndAssertShanghaiContract(t *testing.T, client *thorclient.Client, acc *common.Account) {
	tag, err := client.ChainTag()
	require.NoError(t, err)

	contractData := decodedShanghaiContract(t)

	deployContractTx := new(tx.Builder).
		ChainTag(tag).
		Expiration(math.MaxUint32).
		Gas(10_000_000).
		GasPriceCoef(128).
		BlockRef(tx.NewBlockRef(0)).
		Nonce(datagen.RandUInt64()).
		Clause(
			tx.NewClause(nil).WithData(contractData),
		).Build()

	depContractInspectResults, err := client.InspectTxClauses(deployContractTx, acc.Address)
	require.NoError(t, err)
	for _, respClause := range depContractInspectResults {
		require.False(t, respClause.Reverted || respClause.VMError != "")
	}

	signedTxHash, err := crypto.Sign(deployContractTx.SigningHash().Bytes(), acc.PrivateKey)
	require.NoError(t, err)
	issuedTx, err := client.SendTransaction(deployContractTx.WithSignature(signedTxHash))
	require.NoError(t, err)

	var contractAddr *thor.Address
	err = common.Retry(func() error {
		receipt, err := client.TransactionReceipt(issuedTx.ID)
		if err != nil {
			return fmt.Errorf("unable to retrieve tx receipt - %w", err)
		}
		if receipt.Reverted {
			return fmt.Errorf("transaction was reverted - %+v", receipt)
		}
		contractAddr = receipt.Outputs[0].ContractAddress
		return nil
	}, 3*time.Second, 8)

	require.NoError(t, err)
	require.NotNil(t, contractAddr)
	t.Logf("Successfully deployed Shanghai contract at address: %s", contractAddr.String())
}

// https://github.com/vechain/thor-e2e-tests/blob/main/contracts/shanghai/SimpleCounterShanghai.sol
const shanghaiContractBytecode = "0x608060405234801561000f575f80fd5b505f805561016e806100205f395ff3fe608060405234801561000f575f80fd5b506004361061003f575f3560e01c80635b34b966146100435780638ada066e1461004d5780638bb5d9c314610061575b5f80fd5b61004b610074565b005b5f5460405190815260200160405180910390f35b61004b61006f3660046100fd565b6100c3565b5f8054908061008283610114565b91905055507f3cf8b50771c17d723f2cb711ca7dadde485b222e13c84ba0730a14093fad6d5c5f546040516100b991815260200160405180910390a1565b5f8190556040518181527f3cf8b50771c17d723f2cb711ca7dadde485b222e13c84ba0730a14093fad6d5c9060200160405180910390a150565b5f6020828403121561010d575f80fd5b5035919050565b5f6001820161013157634e487b7160e01b5f52601160045260245ffd5b506001019056fea2646970667358221220aa73e6082b52bca8243902c639e5386b481c2183e8400f34731c4edb93d87f6764736f6c63430008180033"

func decodedShanghaiContract(t *testing.T) []byte {
	contractBytecode, err := hexutil.Decode(shanghaiContractBytecode)
	require.NoError(t, err)
	return contractBytecode
}
