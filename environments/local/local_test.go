package local

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/preset"
	"github.com/vechain/networkhub/utils/client"
	"github.com/vechain/networkhub/utils/datagen"
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
  "id": "local_test",
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

	localEnv := NewLocalEnv()
	_, err = localEnv.LoadConfig(networkCfg)
	require.Error(t, err)

	require.True(t, strings.Contains(err.Error(), "does not exist at path"))
}

func TestLocal(t *testing.T) {
	t.Skip()
	networkCfg, err := network.NewNetwork(
		network.WithJSON(networkJSON),
	)
	require.NoError(t, err)

	fmt.Println(networkJSON)
	localEnv := NewLocalEnv()
	_, err = localEnv.LoadConfig(networkCfg)
	require.NoError(t, err)

	err = localEnv.StartNetwork()
	require.NoError(t, err)

	time.Sleep(30 * time.Second)
	c := client.NewClient("http://" + networkCfg.Nodes[0].GetAPIAddr())
	account, err := c.GetAccount(datagen.RandAccount().Address)
	require.NoError(t, err)

	fmt.Println(account)

	time.Sleep(time.Minute)
	err = localEnv.StopNetwork()
	require.NoError(t, err)
}

func TestThreeNodes(t *testing.T) {
	var err error

	downloader := NewDownloader("")
	thorBinPath, err := downloader.CloneAndBuildThor()
	require.NoError(t, err)

	networkCfg := preset.LocalThreeMasterNodesNetwork

	// ensure the artifact path is set
	for _, node := range networkCfg.Nodes {
		node.SetExecArtifact(thorBinPath)
	}
	localEnv := NewLocalEnv()
	_, err = localEnv.LoadConfig(networkCfg)
	require.NoError(t, err)

	err = localEnv.StartNetwork()
	require.NoError(t, err)

	time.Sleep(30 * time.Second)
	c := client.NewClient(networkCfg.Nodes[0].GetHTTPAddr())
	account, err := c.GetAccount(datagen.RandAccount().Address)
	require.NoError(t, err)

	fmt.Println(account)

	time.Sleep(30 * time.Second)
	err = localEnv.StopNetwork()
	require.NoError(t, err)
}

func TestSixNode(t *testing.T) {
	t.Skip()
	sixNodeJson, err := json.Marshal(preset.LocalSixNodesNetwork)
	require.NoError(t, err)

	networkCfg, err := network.NewNetwork(
		network.WithJSON(string(sixNodeJson)),
	)
	require.NoError(t, err)

	// ensure the artifact path is set
	for _, node := range networkCfg.Nodes {
		node.SetExecArtifact("/Users/user/go/src/github.com/vechain/thor/bin/thor")
	}

	localEnv := NewLocalEnv()
	_, err = localEnv.LoadConfig(networkCfg)
	require.NoError(t, err)

	err = localEnv.StartNetwork()
	require.NoError(t, err)

	time.Sleep(30 * time.Second) // TODO: change this to a polling approach
	for _, node := range networkCfg.Nodes {
		c := client.NewClient("http://" + node.GetAPIAddr())
		peers, err := c.GetPeers()
		require.NoError(t, err)

		require.GreaterOrEqual(t, len(peers), 0)
	}
	err = localEnv.StopNetwork()
	require.NoError(t, err)
}

func TestFourNodesGalactica(t *testing.T) {
	fourNodesGalacticaJson, err := json.Marshal(preset.LocalFourNodesGalacticaNetwork)
	require.NoError(t, err)

	networkCfg, err := network.NewNetwork(
		network.WithJSON(string(fourNodesGalacticaJson)),
	)
	require.NoError(t, err)

	localEnv := NewLocalEnv()
	_, err = localEnv.LoadConfig(networkCfg)
	require.NoError(t, err)

	err = localEnv.StartNetwork()
	require.NoError(t, err)

	time.Sleep(30 * time.Second) // TODO: change this to a polling approach
	for _, node := range networkCfg.Nodes {
		c := client.NewClient("http://" + node.GetAPIAddr())
		peers, err := c.GetPeers()
		require.NoError(t, err)

		require.GreaterOrEqual(t, len(peers), 0)
	}
	err = localEnv.StopNetwork()
	require.NoError(t, err)
}
