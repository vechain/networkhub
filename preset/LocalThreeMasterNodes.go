package preset

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/vechain/networkhub/environments"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/thor/v2/genesis"
	"github.com/vechain/thor/v2/thor"
)

var LocalThreeMasterNodesNetwork = &network.Network{
	ID:          "threeMaster",
	Environment: environments.Local,
	Nodes: []*node.Node{
		{
			ID:            "node1",
			ExecArtifact:  "/app/thor",
			Genesis:       localThreeMasterNodesNetworkGenesis,
			P2PListenPort: 8081,
			APIAddr:       "0.0.0.0:8181",
			APICORS:       "*",
			Type:          node.MasterNode,
			Key:           "01a4107bfb7d5141ec519e75788c34295741a1eefbfe460320efd2ada944071e", // 0x61fF580B63D3845934610222245C116E013717ec
			Enode:         "enode://2ac08a2c35f090e5c47fe99bb0b2956d5b3366c61a83ef30719d393b5984227f4a5bb35b42fef94c3c03c1797ddd97546bb6eeb627b040c4c8dd554b4289024d@127.0.0.1:8081",
		},
		{
			ID:            "node2",
			ExecArtifact:  "/app/thor",
			Genesis:       localThreeMasterNodesNetworkGenesis,
			P2PListenPort: 8082,
			APIAddr:       "0.0.0.0:8182",
			APICORS:       "*",
			Type:          node.MasterNode,
			Key:           "7072249b800ddac1d29a3cd06468cc1a917cbcd110dde358a905d03dad51748d", // 0x327931085B4cCbCE0baABb5a5E1C678707C51d90
			Enode:         "enode://ca36cbb2e9ad0ed582350ee04f49408f4fa409a8ca39982a34e4d5bb82418c45f3fd74bc4861f5aaecd986f1697f28010e1f6af7fadf08c6f529188752f47bee@127.0.0.1:8082",
		},
		{
			ID:            "node3",
			ExecArtifact:  "/app/thor",
			Genesis:       localThreeMasterNodesNetworkGenesis,
			P2PListenPort: 8083,
			APIAddr:       "0.0.0.0:8183",
			APICORS:       "*",
			Type:          node.MasterNode,
			Key:           "c55455943bf026dc44fcf189e8765eb0587c94e66029d580bae795386c0b737a", // 0x084E48c8AE79656D7e27368AE5317b5c2D6a7497
			Enode:         "enode://2d5b5f39e906dd717d721e3f039326e55163697e99e0a9998193eddfbb42e21a457ab877c355ee89c2bdf2562c86f6946b1e98119e945c091cab1a5ded8ca027@127.0.0.1:8083",
		},
	},
}

var localThreeMasterEndorser = thor.MustParseAddress("0x0000000000000000000000004578656375746f72")

func convToHexOrDecimal256(i *big.Int) *genesis.HexOrDecimal256 {
	tmp := genesis.HexOrDecimal256(*i)
	return &tmp
}

var localThreeMasterNodesNetworkGenesis = &genesis.CustomGenesis{
	LaunchTime: 1703180212,
	GasLimit:   10000000,
	ExtraData:  "",
	Accounts: []genesis.Account{
		{
			Address: thor.MustParseAddress("0x7567d83b7b8d80addcb281a71d54fc7b3364ffed"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(big.NewInt(0)),
			Code:    "0x6060604052600256",
			Storage: map[string]thor.Bytes32{
				"0x0000000000000000000000000000000000000000000000000000000000000001": thor.MustParseBytes32("0x0000000000000000000000000000000000000000000000000000000000000002"),
			},
		},
		{
			Address: thor.MustParseAddress("0x61fF580B63D3845934610222245C116E013717ec"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		},
		{
			Address: thor.MustParseAddress("0x327931085B4cCbCE0baABb5a5E1C678707C51d90"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		},
		{
			Address: thor.MustParseAddress("0x084E48c8AE79656D7e27368AE5317b5c2D6a7497"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		},
	},
	Authority: []genesis.Authority{
		{
			MasterAddress:   thor.MustParseAddress("0x61fF580B63D3845934610222245C116E013717ec"),
			EndorsorAddress: thor.MustParseAddress("0x7567d83b7b8d80addcb281a71d54fc7b3364ffed"),
			Identity:        thor.MustParseBytes32("0x000000000000000068747470733a2f2f636f6e6e65782e76656368612e696e2f"),
		},
		{
			MasterAddress:   thor.MustParseAddress("0x327931085B4cCbCE0baABb5a5E1C678707C51d90"),
			EndorsorAddress: thor.MustParseAddress("0x7567d83b7b8d80addcb281a71d54fc7b3364ffed"),
			Identity:        thor.MustParseBytes32("0x000000000000000068747470733a2f2f656e762e7665636861696e2e6f72672f"),
		},
		{
			MasterAddress:   thor.MustParseAddress("0x084E48c8AE79656D7e27368AE5317b5c2D6a7497"),
			EndorsorAddress: thor.MustParseAddress("0x7567d83b7b8d80addcb281a71d54fc7b3364ffed"),
			Identity:        thor.MustParseBytes32("0x0000000000000068747470733a2f2f617070732e7665636861696e2e6f72672f"),
		},
	},
	Params: genesis.Params{
		RewardRatio:         convToHexOrDecimal256(big.NewInt(300000000000000000)),
		BaseGasPrice:        convToHexOrDecimal256(big.NewInt(1000000000000000)),
		ProposerEndorsement: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		ExecutorAddress:     &localThreeMasterEndorser,
	},
	Executor: genesis.Executor{
		Approvers: []genesis.Approver{
			{
				Address:  thor.MustParseAddress("0x199b836d8a57365baccd4f371c1fabb7be77d389"),
				Identity: thor.MustParseBytes32("0x00000000000067656e6572616c20707572706f736520626c6f636b636861696e"),
			},
		},
	},
	ForkConfig: &thor.ForkConfig{
		VIP191:    0,
		ETH_CONST: 0,
		BLOCKLIST: 0,
		ETH_IST:   0,
		VIP214:    0,
	},
}
