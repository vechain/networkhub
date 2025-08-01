package local

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/preset"
	"github.com/vechain/networkhub/thorbuilder"
	"github.com/vechain/networkhub/utils/common"
	"github.com/vechain/networkhub/utils/datagen"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/thorclient"
	"github.com/vechain/thor/v2/tx"
)

var genesis = `{
        "launchTime": 1703180212,
        "gasLimit": 10000000,
        "forkConfig": {
          "VIP191": 0,
          "ETH_CONST": 0,
          "BLOCKLIST": 0,
          "ETH_IST": 0,
          "VIP214": 0,
          "FINALITY": 0
        },
        "accounts": [
          {
            "address": "0x7567d83b7b8d80addcb281a71d54fc7b3364ffed",
            "balance": "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
            "energy": 0,
            "code": "0x6060604052600256",
            "storage": {
              "0x0000000000000000000000000000000000000000000000000000000000000001": "0x0000000000000000000000000000000000000000000000000000000000000002"
            }
          },
          {
            "address": "0x61fF580B63D3845934610222245C116E013717ec",
            "balance": "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
            "energy": "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
          },
          {
            "address": "0x327931085B4cCbCE0baABb5a5E1C678707C51d90",
            "balance": "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
            "energy": "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
          },
          {
            "address": "0x084E48c8AE79656D7e27368AE5317b5c2D6a7497",
            "balance": "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
            "energy": "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
          }
        ],
        "authority": [
          {
            "masterAddress": "0x61fF580B63D3845934610222245C116E013717ec",
            "endorsorAddress": "0x7567d83b7b8d80addcb281a71d54fc7b3364ffed",
            "identity": "0x000000000000000068747470733a2f2f636f6e6e65782e76656368612e696e2f"
          },
          {
            "masterAddress": "0x327931085B4cCbCE0baABb5a5E1C678707C51d90",
            "endorsorAddress": "0x7567d83b7b8d80addcb281a71d54fc7b3364ffed",
            "identity": "0x000000000000000068747470733a2f2f656e762e7665636861696e2e6f72672f"
          },
          {
            "masterAddress": "0x084E48c8AE79656D7e27368AE5317b5c2D6a7497",
            "endorsorAddress": "0x7567d83b7b8d80addcb281a71d54fc7b3364ffed",
            "identity": "0x0000000000000068747470733a2f2f617070732e7665636861696e2e6f72672f"
          }
        ],
        "params": {
          "rewardRatio": 300000000000000000,
          "baseGasPrice": 1000000000000000,
          "proposerEndorsement": "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
          "executorAddress": "0x0000000000000000000000004578656375746f72"
        },
        "executor": {
          "approvers": [
            {
              "address": "0x199b836d8a57365baccd4f371c1fabb7be77d389",
              "identity": "0x00000000000067656e6572616c20707572706f736520626c6f636b636861696e"
            }
          ]
        }
      }`
var networkJSON = fmt.Sprintf(`{
  "baseId": "local_test",
  "nodes": [
    {
      "id": "node1",
      "execArtifact": "/Users/pedro/go/src/github.com/vechain/thor/bin/thor",
      "p2pListenPort": 8081,
      "host": "127.0.0.1",
      "apiAddr": "127.0.0.1:8181",
      "apiCORS": "*",
      "type": "masterNode",
      "key": "01a4107bfb7d5141ec519e75788c34295741a1eefbfe460320efd2ada944071e",
      "genesis": %s
    },
    {
      "id": "node2",
	  "execArtifact": "/Users/pedro/go/src/github.com/vechain/thor/bin/thor",
      "p2pListenPort": 8082,
      "host": "127.0.0.1",
      "apiAddr": "127.0.0.1:8182",
      "apiCORS": "*",
      "type": "masterNode",
      "key": "7072249b800ddac1d29a3cd06468cc1a917cbcd110dde358a905d03dad51748d",
      "genesis": %s
    },
    {
      "id": "node3",
	  "execArtifact": "/Users/pedro/go/src/github.com/vechain/thor/bin/thor",
      "p2pListenPort": 8083,
      "host": "127.0.0.1",
      "apiAddr": "127.0.0.1:8183",
      "apiCORS": "*",
      "type": "masterNode",
      "key": "c55455943bf026dc44fcf189e8765eb0587c94e66029d580bae795386c0b737a",
	  "genesis": %s
    }
  ]
}`, genesis, genesis, genesis)

func TestLocalInvalidExecArtifact(t *testing.T) {
	networkCfg, err := network.NewNetwork(
		network.WithJSON(networkJSON),
	)
	require.NoError(t, err)

	networkCfg.Nodes[0].SetExecArtifact("/some_fake_dir")

	localEnv := NewEnv()
	_, err = localEnv.LoadConfig(networkCfg)
	require.Error(t, err)

	require.ErrorContains(t, err, "exec artifact path /some_fake_dir does not exist")
}

func TestLocal(t *testing.T) {
	t.Skip()
	networkCfg, err := network.NewNetwork(
		network.WithJSON(networkJSON),
	)
	require.NoError(t, err)

	slog.Info(networkJSON)
	localEnv := NewEnv()
	_, err = localEnv.LoadConfig(networkCfg)
	require.NoError(t, err)

	require.NoError(t, localEnv.StartNetwork())

	time.Sleep(30 * time.Second)
	c := thorclient.New(networkCfg.Nodes[0].GetHTTPAddr())
	account, err := c.Account(datagen.RandAccount().Address)
	require.NoError(t, err)

	slog.Info("Account", "acc", account)

	time.Sleep(time.Minute)
	err = localEnv.StopNetwork()
	require.NoError(t, err)
}

func TestThreeNodes(t *testing.T) {
	var err error
	cfg := thorbuilder.DefaultConfig()
	cfg.DownloadConfig.IsReusable = false
	thorBuilder := thorbuilder.New(cfg)
	require.NoError(t, thorBuilder.Download())
	thorBinPath, err := thorBuilder.Build()
	require.NoError(t, err)

	networkCfg := preset.LocalThreeMasterNodesNetwork()

	// ensure the artifact path is set
	for _, node := range networkCfg.Nodes {
		node.SetExecArtifact(thorBinPath)
	}
	localEnv := NewEnv()
	_, err = localEnv.LoadConfig(networkCfg)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, localEnv.StopNetwork())
	})
	require.NoError(t, localEnv.StartNetwork())

	err = networkCfg.HealthCheck(0, time.Second*30)
	require.NoError(t, err)

	c := thorclient.New(networkCfg.Nodes[0].GetHTTPAddr())
	account, err := c.Account(datagen.RandAccount().Address)
	require.NoError(t, err)

	slog.Info("account:", "acc", account)
}

func TestSixNodes(t *testing.T) {
	sixNodeJson, err := json.Marshal(preset.LocalSixNodesNetwork())
	require.NoError(t, err)

	networkCfg, err := network.NewNetwork(
		network.WithJSON(string(sixNodeJson)),
	)
	require.NoError(t, err)

	thorBuilder := thorbuilder.New(thorbuilder.DefaultConfig())
	require.NoError(t, thorBuilder.Download())
	thorBinPath, err := thorBuilder.Build()
	require.NoError(t, err)

	// ensure the artifact path is set
	for _, node := range networkCfg.Nodes {
		node.SetExecArtifact(thorBinPath)
	}

	localEnv := NewEnv()
	_, err = localEnv.LoadConfig(networkCfg)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, localEnv.StopNetwork())
	})
	require.NoError(t, localEnv.StartNetwork())
	assert.NoError(t, networkCfg.HealthCheck(0, time.Second*20))

	pollingWhileConnectingPeers(t, networkCfg.Nodes, 5)
}

func TestSixNodesGalactica(t *testing.T) {
	t.Skip()
	var sixNodesGalacticaNetwork *network.Network
	require.NotPanics(t, func() { sixNodesGalacticaNetwork = preset.LocalSixNodesGalacticaNetwork() })

	localEnv := NewEnv()
	_, err := localEnv.LoadConfig(sixNodesGalacticaNetwork)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, localEnv.StopNetwork())
	})
	require.NoError(t, localEnv.StartNetwork())

	clients := pollingWhileConnectingPeers(t, sixNodesGalacticaNetwork.Nodes, 5)

	deployAndAssertShanghaiContract(t, clients[0], preset.SixNNAccount1)
}

func TestThreeNodes_Healthcheck(t *testing.T) {
	networkCfg := preset.LocalThreeMasterNodesNetwork()

	thorBuilder := thorbuilder.New(thorbuilder.DefaultConfig())
	require.NoError(t, thorBuilder.Download())
	thorBinPath, err := thorBuilder.Build()
	require.NoError(t, err)

	// ensure the artifact path is set
	for _, node := range networkCfg.Nodes {
		node.SetExecArtifact(thorBinPath)
	}

	localEnv := NewEnv()
	_, err = localEnv.LoadConfig(networkCfg)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, localEnv.StopNetwork())
	})
	require.NoError(t, localEnv.StartNetwork())

	assert.NoError(t, networkCfg.HealthCheck(0, time.Second*20))
}

func TestThreeNodes_AdditionalArgs(t *testing.T) {
	networkCfg := preset.LocalThreeMasterNodesNetwork()

	thorBuilder := thorbuilder.New(thorbuilder.DefaultConfig())
	require.NoError(t, thorBuilder.Download())
	thorBinPath, err := thorBuilder.Build()
	require.NoError(t, err)

	// ensure the artifact path is set
	for _, node := range networkCfg.Nodes {
		node.SetExecArtifact(thorBinPath)
		node.SetAdditionalArgs(map[string]string{
			"api-allowed-tracers": "call",
		})
	}

	localEnv := NewEnv()
	_, err = localEnv.LoadConfig(networkCfg)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, localEnv.StopNetwork())
	})
	require.NoError(t, localEnv.StartNetwork())

	assert.NoError(t, networkCfg.HealthCheck(0, time.Second*20))

	client := thorclient.New(networkCfg.Nodes[0].GetHTTPAddr())
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
	body := string(res)
	println("Response body: %s", body)
}

func pollingWhileConnectingPeers(t *testing.T, nodes []node.Config, expectedPeersLen int) []*thorclient.Client {
	// Polling approach with timeout
	timeout := time.After(1 * time.Minute)
	tick := time.Tick(5 * time.Second)

	clients := make([]*thorclient.Client, 0)
	for {
		select {
		case <-timeout:
			t.Fatal("timed out waiting for nodes to connect")
		case <-tick:
			allConnected := true
			for _, node := range nodes {
				c := thorclient.New(node.GetHTTPAddr())
				peers, err := c.Peers()
				require.NoError(t, err)
				if len(peers) != expectedPeersLen {
					allConnected = false
					clients = clients[:0]
					break
				}
				clients = append(clients, c)
			}
			if allConnected {
				return clients
			}
		}
	}
}

// https://github.com/vechain/thor-e2e-tests/blob/main/contracts/shanghai/SimpleCounterShanghai.sol
const shanghaiContractBytecode = "0x608060405234801561000f575f80fd5b505f805561016e806100205f395ff3fe608060405234801561000f575f80fd5b506004361061003f575f3560e01c80635b34b966146100435780638ada066e1461004d5780638bb5d9c314610061575b5f80fd5b61004b610074565b005b5f5460405190815260200160405180910390f35b61004b61006f3660046100fd565b6100c3565b5f8054908061008283610114565b91905055507f3cf8b50771c17d723f2cb711ca7dadde485b222e13c84ba0730a14093fad6d5c5f546040516100b991815260200190565b60405180910390a1565b5f8190556040518181527f3cf8b50771c17d723f2cb711ca7dadde485b222e13c84ba0730a14093fad6d5c9060200160405180910390a150565b5f6020828403121561010d575f80fd5b5035919050565b5f6001820161013157634e487b7160e01b5f52601160045260245ffd5b506001019056fea2646970667358221220aa73e6082b52bca8243902c639e5386b481c2183e8400f34731c4edb93d87f6764736f6c63430008180033"

func decodedShanghaiContract(t *testing.T) []byte {
	contractBytecode, err := hexutil.Decode(shanghaiContractBytecode)
	require.NoError(t, err)
	return contractBytecode
}

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
}
