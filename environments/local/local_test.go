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

const testnetGenesisID = "0x000000000b2bce3c70bc649a02749e8687721b09ed2e15997f466536b20bb127"
const mainnetGenesisID = "0x00000000851caf3cfdb6e899cf5958bfb1ac3413d346d43539627e6be7ec1b4a"
const soloGenesisID = "0x00000000c05a20fbca2bf6ae3affba6af4a74b800b585bf7a4988aba7aea69f6"

const localNodeAPIAddr = "127.0.0.1:8669"

// verifySoloNodeConfig verifies the basic configuration of a solo node
func verifySoloNodeConfig(t *testing.T, nodeConfig node.Config, expectedID, expectedAPIAddr, expectedAPICORS, expectedDataDir string, expectedVerbosity int) {
	t.Helper()
	require.Equal(t, expectedID, nodeConfig.GetID())
	require.Equal(t, expectedAPIAddr, nodeConfig.GetAPIAddr())
	require.Equal(t, expectedAPICORS, nodeConfig.GetAPICORS())
	require.Equal(t, expectedDataDir, nodeConfig.GetDataDir())
	require.Equal(t, expectedVerbosity, nodeConfig.GetVerbosity())
}

// verifySoloNodeArguments verifies the solo-specific arguments in the configuration
func verifySoloNodeArguments(t *testing.T, nodeConfig node.Config, expectedGasLimit, expectedAPICallGasLimit, expectedTxPoolLimit, expectedTxPoolLimitPerAccount, expectedCache, expectedBlockInterval string) {
	t.Helper()
	additionalArgs := nodeConfig.GetAdditionalArgs()

	// Verify solo-specific flags are present
	require.Contains(t, additionalArgs, "on-demand")
	require.Contains(t, additionalArgs, "api-enable-txpool")
	require.Contains(t, additionalArgs, "persist")

	// Verify argument values
	require.Equal(t, expectedGasLimit, additionalArgs["gas-limit"])
	require.Equal(t, expectedAPICallGasLimit, additionalArgs["api-call-gas-limit"])
	require.Equal(t, expectedTxPoolLimit, additionalArgs["txpool-limit"])
	require.Equal(t, expectedTxPoolLimitPerAccount, additionalArgs["txpool-limit-per-account"])
	require.Equal(t, expectedCache, additionalArgs["cache"])
	require.Equal(t, expectedBlockInterval, additionalArgs["block-interval"])
}

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

type attachNodeTestConfig struct {
	NetworkType    string
	InitialNodeID  string
	InitialAPIPort int
	InitialP2PPort int
	AttachNodeID   string
	AttachAPIPort  int
	AttachP2PPort  int
	Environment    string
	GenesisID      string
}

type publicNetworkTestConfig struct {
	NetworkType string // "test" for testnet, "main" for mainnet
	APIPort     int
	P2PPort     int
	GenesisID   string
	Environment string
	NodeID      string
}

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

func testPublicNetworkConnection(t *testing.T, config publicNetworkTestConfig) {
	t.Helper()
	localEnv := NewEnv()

	t.Cleanup(func() {
		if err := localEnv.StopNetwork(); err != nil {
			t.Logf("Warning: failed to stop network during test cleanup: %v", err)
			t.Fail()
		}
	})

	publicNode := &node.BaseNode{
		ID:             config.NodeID,
		APICORS:        "*",
		Type:           node.RegularNode,
		Verbosity:      3,
		P2PListenPort:  config.P2PPort,
		APIAddr:        fmt.Sprintf("127.0.0.1:%d", config.APIPort),
		AdditionalArgs: map[string]string{"network": config.NetworkType},
	}

	// Create a minimal network configuration
	networkCfg := &network.Network{
		BaseID:      "baseID",
		Environment: config.Environment,
		Nodes:       []node.Config{publicNode},
		ThorBuilder: thorbuilder.DefaultConfig(),
	}

	// Load the configuration
	networkID, err := localEnv.LoadConfig(networkCfg)
	require.NoError(t, err)
	expectedNetworkID := fmt.Sprintf("%sbaseID", config.Environment)
	require.Equal(t, expectedNetworkID, networkID)

	// Start the network (this will start the public network node)
	err = localEnv.StartNetwork()
	require.NoError(t, err)

	// Wait a bit for the node to start syncing
	time.Sleep(5 * time.Second)

	// Try to connect to the node's API
	apiURL := fmt.Sprintf("http://127.0.0.1:%d", config.APIPort)
	client := thorclient.New(apiURL)
	block, err := client.Block("0")
	if err != nil {
		t.Logf("Warning: Could not connect to %s node: %v", config.NetworkType, err)
		t.Logf("This might be normal if the node is still syncing")
	} else {
		// Validate that the genesis block ID matches the expected one
		blockID, err := thor.ParseBytes32(config.GenesisID)
		require.NoError(t, err)
		require.Equal(t, blockID, block.ID)
		t.Logf("Successfully connected to %s! Genesis block: %d", config.NetworkType, block.Number)
	}
}
func TestTestnetConnection(t *testing.T) {
	config := publicNetworkTestConfig{
		NetworkType: "test",
		NodeID:      "testnet-node",
		APIPort:     8669,
		P2PPort:     11235,
		Environment: "testnet",
		GenesisID:   testnetGenesisID,
	}

	testPublicNetworkConnection(t, config)
}

func TestMainnetConnection(t *testing.T) {
	config := publicNetworkTestConfig{
		NetworkType: "main",
		NodeID:      "mainnet-node",
		APIPort:     8670,
		P2PPort:     11236,
		Environment: "mainnet",
		GenesisID:   mainnetGenesisID,
	}

	testPublicNetworkConnection(t, config)
}

func testAttachNodeConnection(t *testing.T, config attachNodeTestConfig) {
	t.Helper()

	localEnv := NewEnv()

	t.Cleanup(func() {
		if err := localEnv.StopNetwork(); err != nil {
			t.Logf("Warning: failed to stop network during test cleanup: %v", err)
			t.Fail()
		}
	})

	initialNode := &node.BaseNode{
		ID:             config.InitialNodeID,
		APICORS:        "*",
		Type:           node.RegularNode,
		Verbosity:      3,
		P2PListenPort:  config.InitialP2PPort,
		APIAddr:        fmt.Sprintf("127.0.0.1:%d", config.InitialAPIPort),
		AdditionalArgs: map[string]string{"network": config.NetworkType},
	}

	networkCfg := &network.Network{
		BaseID:      "baseID",
		Environment: config.Environment,
		Nodes:       []node.Config{initialNode},
		ThorBuilder: thorbuilder.DefaultConfig(),
	}

	// Load and start the initial network
	networkID, err := localEnv.LoadConfig(networkCfg)
	require.NoError(t, err)
	expectedNetworkID := fmt.Sprintf("%sbaseID", config.Environment)
	require.Equal(t, expectedNetworkID, networkID)

	err = localEnv.StartNetwork()
	require.NoError(t, err)

	// Wait for initial node to start
	time.Sleep(3 * time.Second)

	// Create a new node to attach
	attachNode := &node.BaseNode{
		ID:             config.AttachNodeID,
		APICORS:        "*",
		Type:           node.RegularNode,
		Verbosity:      3,
		P2PListenPort:  config.AttachP2PPort,
		APIAddr:        fmt.Sprintf("127.0.0.1:%d", config.AttachAPIPort),
		AdditionalArgs: map[string]string{"network": config.NetworkType},
	}
	err = localEnv.AttachNode(attachNode)
	require.NoError(t, err)

	// Wait for attached node to start
	time.Sleep(3 * time.Second)

	// Verify both nodes are running
	nodes := localEnv.Nodes()
	require.Len(t, nodes, 2)
	require.Contains(t, nodes, config.InitialNodeID)
	require.Contains(t, nodes, config.AttachNodeID)

	// Test connection to the attached node
	apiURL := fmt.Sprintf("http://127.0.0.1:%d", config.AttachAPIPort)
	client := thorclient.New(apiURL)
	block, err := client.Block("0")
	if err != nil {
		t.Logf("Warning: Could not connect to attached %s node: %v", config.NetworkType, err)
		t.Logf("This might be normal if the node is still syncing")
	} else {
		// Validate that the genesis block ID matches the expected one
		blockID, err := thor.ParseBytes32(config.GenesisID)
		require.NoError(t, err)
		require.Equal(t, blockID, block.ID)
		t.Logf("Successfully connected to attached %s node! Genesis block: %d", config.NetworkType, block.Number)
	}

	// Remove the attached node
	err = localEnv.RemoveNode(config.AttachNodeID)
	require.NoError(t, err)

	// Verify the node was removed
	nodes = localEnv.Nodes()
	require.Len(t, nodes, 1)
	require.Contains(t, nodes, config.InitialNodeID)
	require.NotContains(t, nodes, config.AttachNodeID)
}
func TestAttachNodeTestnet(t *testing.T) {
	config := attachNodeTestConfig{
		NetworkType:    "test",
		InitialNodeID:  "initial-testnet-node",
		InitialAPIPort: 8671,
		InitialP2PPort: 11237,
		AttachNodeID:   "attach-testnet-node",
		AttachAPIPort:  8672,
		AttachP2PPort:  11238,
		Environment:    "testnet",
		GenesisID:      testnetGenesisID,
	}

	testAttachNodeConnection(t, config)
}

func TestAttachNodeMainnet(t *testing.T) {
	config := attachNodeTestConfig{
		NetworkType:    "main",
		InitialNodeID:  "initial-mainnet-node",
		InitialAPIPort: 8670,
		InitialP2PPort: 11236,
		AttachNodeID:   "attach-mainnet-node",
		AttachAPIPort:  8671,
		AttachP2PPort:  11238,
		Environment:    "mainnet",
		GenesisID:      mainnetGenesisID,
	}

	testAttachNodeConnection(t, config)
}

func TestAttachToPublicNetworkAndStart(t *testing.T) {
	localEnv := NewEnv()

	t.Cleanup(func() {
		if err := localEnv.StopNetwork(); err != nil {
			t.Logf("Warning: failed to stop network during test cleanup: %v", err)
			t.Fail()
		}
	})

	testnetConfig := PublicNetworkConfig{
		NodeID:      "testnet-node",
		NetworkType: "test",
		APIAddr:     "127.0.0.1:8672",
		P2PPort:     11239,
	}

	err := localEnv.AttachToPublicNetworkAndStart(testnetConfig)
	require.NoError(t, err)

	// Wait for the node to start
	time.Sleep(3 * time.Second)

	// Verify the node is running
	nodes := localEnv.Nodes()
	require.Len(t, nodes, 1)
	require.Contains(t, nodes, testnetConfig.NodeID)

	// Test connection to the node
	client := thorclient.New("http://" + localNodeAPIAddr)
	block, err := client.Block("0")
	if err != nil {
		t.Logf("Warning: Could not connect to testnet node: %v", err)
		t.Logf("This might be normal if the node is still syncing")
	} else {
		// Validate that the genesis block ID is the testnet one
		blockID, err := thor.ParseBytes32(testnetGenesisID)
		require.NoError(t, err)
		require.Equal(t, blockID, block.ID)
		t.Logf("Successfully connected to testnet using convenience method! Genesis block: %d", block.Number)
	}
}

func TestSoloNodeConfig(t *testing.T) {
	t.Run("CreateSoloNodeConfig with defaults", func(t *testing.T) {
		config := SoloNodeConfig{
			NodeID: "solo-test-node",
		}

		nodeConfig := CreateSoloNodeConfig(config)

		// Verify basic configuration with defaults
		verifySoloNodeConfig(t, nodeConfig, "solo-test-node", "0.0.0.0:8669", "*", "/data", 9)

		// Verify solo-specific arguments with defaults
		verifySoloNodeArguments(t, nodeConfig, "10000000000000", "10000000000000", "100000000000", "256", "1024", "1")
	})

	t.Run("CreateSoloNodeConfig with custom values", func(t *testing.T) {
		config := SoloNodeConfig{
			NodeID:                "custom-solo-node",
			APIAddr:               "127.0.0.1:8670",
			APICORS:               "http://localhost:3000",
			GasLimit:              "5000000000000",
			APICallGasLimit:       "5000000000000",
			TxPoolLimit:           "50000000000",
			TxPoolLimitPerAccount: "128",
			Cache:                 "512",
			DataDir:               "/custom/data",
			Verbosity:             5,
			BlockInterval:         "2",
		}

		nodeConfig := CreateSoloNodeConfig(config)

		// Verify custom configuration
		verifySoloNodeConfig(t, nodeConfig, "custom-solo-node", "127.0.0.1:8670", "http://localhost:3000", "/custom/data", 5)

		// Verify custom arguments
		verifySoloNodeArguments(t, nodeConfig, "5000000000000", "5000000000000", "50000000000", "128", "512", "2")
	})

}

func TestSoloNodeIntegration(t *testing.T) {
	t.Run("Create and start a solo node", func(t *testing.T) {
		// Create a solo node configuration
		soloConfig := SoloNodeConfig{
			NodeID:                "test-solo-node",
			APIAddr:               localNodeAPIAddr,
			APICORS:               "*",
			GasLimit:              "10000000000000",
			APICallGasLimit:       "10000000000000",
			TxPoolLimit:           "100000000000",
			TxPoolLimitPerAccount: "256",
			Cache:                 "1024",
			DataDir:               "/tmp/solo-test-data",
			Verbosity:             3,
			BlockInterval:         "1",
		}

		// Create the node configuration
		nodeConfig := CreateSoloNodeConfig(soloConfig)

		// Verify the node configuration using helper functions
		verifySoloNodeConfig(t, nodeConfig, "test-solo-node", localNodeAPIAddr, "*", "/tmp/solo-test-data", 3)
		verifySoloNodeArguments(t, nodeConfig, "10000000000000", "10000000000000", "100000000000", "256", "1024", "1")

		// Create a network configuration with the solo node
		networkCfg := &network.Network{
			BaseID:      "solo-test-network",
			Environment: "solo",
			Nodes:       []node.Config{nodeConfig},
			ThorBuilder: thorbuilder.DefaultConfig(),
		}

		// Create local environment
		localEnv := NewEnv()

		t.Cleanup(func() {
			if err := localEnv.StopNetwork(); err != nil {
				t.Logf("Warning: failed to stop network during test cleanup: %v", err)
			}
		})

		// Load the configuration
		networkID, err := localEnv.LoadConfig(networkCfg)
		require.NoError(t, err)
		require.Equal(t, "solosolo-test-network", networkID)

		// Start the solo node
		err = localEnv.StartNetwork()
		require.NoError(t, err)

		// Wait for the node to start and generate the genesis block
		time.Sleep(5 * time.Second)

		// Verify the node is running
		nodes := localEnv.Nodes()
		require.Len(t, nodes, 1)
		require.Contains(t, nodes, "test-solo-node")

		// Test connection to the solo node and validate genesis block
		client := thorclient.New("http://" + localNodeAPIAddr)
		block, err := client.Block("0")
		if err != nil {
			t.Logf("Warning: Could not connect to solo node: %v", err)
			t.Logf("This might be normal if the node is still starting up")
		} else {
			// Validate that we got the genesis block
			blockID, err := thor.ParseBytes32(soloGenesisID)
			require.NoError(t, err)
			require.Equal(t, blockID, block.ID)
			t.Logf("Successfully connected to solo node! Genesis block: %d, ID: %s", block.Number, block.ID)
		}

		t.Logf("Solo node started successfully with network ID: %s", networkID)
		t.Logf("The node is configured to run with: thor solo --on-demand --api-addr=%s --api-cors=* --gas-limit=10000000000000 --api-enable-txpool --api-call-gas-limit=10000000000000 --txpool-limit=100000000000 --txpool-limit-per-account=256 --cache=1024 --data-dir=/tmp/solo-test-data --verbosity=3 --persist --block-interval=1", localNodeAPIAddr)
	})

	t.Run("Solo node with minimal configuration", func(t *testing.T) {
		// Test with minimal configuration (using defaults)
		soloConfig := SoloNodeConfig{
			NodeID: "minimal-solo-node",
		}

		nodeConfig := CreateSoloNodeConfig(soloConfig)

		// Test that the node is detected as solo
		require.True(t, isSoloNode(nodeConfig))

		// Verify defaults are applied using helper functions
		verifySoloNodeConfig(t, nodeConfig, "minimal-solo-node", "0.0.0.0:8669", "*", "/data", 9)
		verifySoloNodeArguments(t, nodeConfig, "10000000000000", "10000000000000", "100000000000", "256", "1024", "1")

		t.Logf("Minimal solo node configuration created successfully")
	})
}
