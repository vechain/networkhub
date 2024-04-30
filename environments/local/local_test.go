package local

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/utils/client"
	"github.com/vechain/networkhub/utils/datagen"
)

var networkJSON = `{
  "nodes": [
    {
      "id": "node1",
      "p2pListenPort": 8081,
      "genesis": "/Users/pedro/tmp/multiple-nodes/custom_genesis.json",
      "dataDir": "/Users/pedro/tmp/multiple-nodes/node1",
      "configDir": "/Users/pedro/tmp/multiple-nodes/node1/config",
      "apiAddr": "127.0.0.1:8181",
      "apiCORS": "*",
      "type": "masterNode",
      "key": "01a4107bfb7d5141ec519e75788c34295741a1eefbfe460320efd2ada944071e",
      "enode": "enode://2ac08a2c35f090e5c47fe99bb0b2956d5b3366c61a83ef30719d393b5984227f4a5bb35b42fef94c3c03c1797ddd97546bb6eeb627b040c4c8dd554b4289024d@127.0.0.1:8081"
    },
    {
      "id": "node2",
      "p2pListenPort": 8082,
	  "genesis": "/Users/pedro/tmp/multiple-nodes/custom_genesis.json",
      "dataDir": "/Users/pedro/tmp/multiple-nodes/node2",
      "configDir": "/Users/pedro/tmp/multiple-nodes/node2/config",
      "apiAddr": "127.0.0.1:8182",
      "apiCORS": "*",
      "type": "masterNode",
      "key": "7072249b800ddac1d29a3cd06468cc1a917cbcd110dde358a905d03dad51748d",
      "enode": "enode://ca36cbb2e9ad0ed582350ee04f49408f4fa409a8ca39982a34e4d5bb82418c45f3fd74bc4861f5aaecd986f1697f28010e1f6af7fadf08c6f529188752f47bee@127.0.0.1:8082"
    },
    {
      "id": "node3",
      "p2pListenPort": 8083,
	  "genesis": "/Users/pedro/tmp/multiple-nodes/custom_genesis.json",
      "dataDir": "/Users/pedro/tmp/multiple-nodes/node3",
      "configDir": "/Users/pedro/tmp/multiple-nodes/node3/config",
      "apiAddr": "127.0.0.1:8183",
      "apiCORS": "*",
      "type": "masterNode",
      "key": "c55455943bf026dc44fcf189e8765eb0587c94e66029d580bae795386c0b737a",
      "enode": "enode://2d5b5f39e906dd717d721e3f039326e55163697e99e0a9998193eddfbb42e21a457ab877c355ee89c2bdf2562c86f6946b1e98119e945c091cab1a5ded8ca027@127.0.0.1:8083"
    }
  ]
}`

func TestLocal(t *testing.T) {
	networkCfg, err := network.NewNetwork(
		network.WithJSON(networkJSON),
	)
	require.NoError(t, err)

	fmt.Println(networkCfg)
	localEnv := NewLocalEnv()

	err = localEnv.StartNetwork(networkCfg)
	require.NoError(t, err)

	time.Sleep(30 * time.Second)
	c := client.NewClient("http://" + networkCfg.Nodes[0].APIAddr)
	account, err := c.GetAccount(datagen.RandAccount().Address)
	require.NoError(t, err)

	fmt.Println(account)

	time.Sleep(time.Hour)
	//client1 := network1.GetNode(nodeID).GetClient()
	//
	//account1, err := client1.GetAccount(acc1)
	//require.NoError(t, err)
	//
	err = localEnv.StopNetwork()
	require.NoError(t, err)
}
