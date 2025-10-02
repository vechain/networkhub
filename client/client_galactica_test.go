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
	"github.com/vechain/thor/v2/tx"
)

// TestClientSixNodesGalactica tests the client with a 6-node Galactica network.
// This test verifies that the client can:
// 1. Set up and start a 6-node Galactica network
// 2. Wait for all nodes to connect (5 peers each)
// 3. Deploy and execute Shanghai-compatible smart contracts
func TestClientSixNodesGalactica(t *testing.T) {
	// Create the six nodes Galactica network
	sixNodesGalacticaNetwork := preset.LocalSixNodesNetwork()

	// Configure thor builder for reusable builds
	cfg := thorbuilder.DefaultConfig()
	cfg.DownloadConfig.IsReusable = false
	sixNodesGalacticaNetwork.ThorBuilder = cfg

	// Create client with the network (automatically starts)
	c, err := New(sixNodesGalacticaNetwork)
	require.NoError(t, err)

	require.NoError(t, c.Start())
	// Cleanup
	defer func() {
		if err := c.Stop(); err != nil {
			t.Logf("Warning: failed to stop client: %v", err)
		}
	}()

	// Wait for all nodes to connect to each other
	t.Log("Waiting for nodes to connect...")
	require.NoError(t, c.network.HealthCheck(2, time.Minute))

	// Deploy and test Shanghai contract to verify network functionality
	t.Log("Deploying Shanghai contract to test Galactica network...")
	deployAndAssertShanghaiContract(t, thorclient.New(c.network.Nodes[0].GetHTTPAddr()), preset.SixNNAccount1)

	t.Log("Successfully tested Galactica network with Shanghai contract deployment!")
}

// deployAndAssertShanghaiContract deploys a Shanghai-compatible smart contract to test network functionality
func deployAndAssertShanghaiContract(t *testing.T, client *thorclient.Client, acc *common.Account) {
	tag, err := client.ChainTag()
	require.NoError(t, err)

	// Build the transaction using the bytecode
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

	// Simulating the contract deployment transaction before deploying it
	depContractInspectResults, err := client.InspectTxClauses(deployContractTx, acc.Address)
	require.NoError(t, err)
	for _, respClause := range depContractInspectResults {
		require.False(t, respClause.Reverted || respClause.VMError != "")
	}

	// Send a transaction
	signedTxHash, err := crypto.Sign(deployContractTx.SigningHash().Bytes(), acc.PrivateKey)
	require.NoError(t, err)
	issuedTx, err := client.SendTransaction(deployContractTx.WithSignature(signedTxHash))
	require.NoError(t, err)

	// Retrieve transaction receipt - GET /transactions/{id}/receipt
	var contractAddr *thor.Address
	const retryPeriod = 3 * time.Second
	const maxRetries = 8
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
	}, retryPeriod, maxRetries)

	require.NoError(t, err)
	require.NotNil(t, contractAddr)
	t.Logf("Successfully deployed Shanghai contract at address: %s", contractAddr.String())
}

// https://github.com/vechain/thor-e2e-tests/blob/main/contracts/shanghai/SimpleCounterShanghai.sol
const shanghaiContractBytecode = "0x608060405234801561000f575f80fd5b505f805561016e806100205f395ff3fe608060405234801561000f575f80fd5b506004361061003f575f3560e01c80635b34b966146100435780638ada066e1461004d5780638bb5d9c314610061575b5f80fd5b61004b610074565b005b5f5460405190815260200160405180910390f35b61004b61006f3660046100fd565b6100c3565b5f8054908061008283610114565b91905055507f3cf8b50771c17d723f2cb711ca7dadde485b222e13c84ba0730a14093fad6d5c5f546040516100b991815260200190565b60405180910390a1565b5f8190556040518181527f3cf8b50771c17d723f2cb711ca7dadde485b222e13c84ba0730a14093fad6d5c9060200160405180910390a150565b5f6020828403121561010d575f80fd5b5035919050565b5f6001820161013157634e487b7160e01b5f52601160045260245ffd5b506001019056fea2646970667358221220aa73e6082b52bca8243902c639e5386b481c2183e8400f34731c4edb93d87f6764736f6c63430008180033"

func decodedShanghaiContract(t *testing.T) []byte {
	contractBytecode, err := hexutil.Decode(shanghaiContractBytecode)
	require.NoError(t, err)
	return contractBytecode
}
