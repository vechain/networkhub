package preset

import (
	"math/big"

	"github.com/vechain/networkhub/consts"
	"github.com/vechain/networkhub/environments"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/network/node/genesis"
	"github.com/vechain/thor/v2/thor"

	thorgenesis "github.com/vechain/thor/v2/genesis"
)

var LocalFourNodesGalacticaNetwork = &network.Network{
	ID:          "fourNodesGalacticaNetwork",
	Environment: environments.Local,
	Nodes: []node.Node{
		&node.NodeGalacticaFork{
			BaseNode: node.BaseNode{
				ID:            "node1",
				P2PListenPort: 8081,
				APIAddr:       "127.0.0.1:8181",
				APICORS:       "*",
				Type:          node.MasterNode,
				Verbosity:     4,
				Key:           "b2c859e115ef4a3f5e4d32228b41de4c661c527a32f723ac37745bf860fd09cb", // 0x5F90f56c7b87E3d1acf9437f0E43E4d687AcEB7e
			},
			Genesis: localFourNodesNetworkGenesis,
		},
		&node.NodeGalacticaFork{
			BaseNode: node.BaseNode{
				ID:            "node2",
				P2PListenPort: 8082,
				APIAddr:       "127.0.0.1:8182",
				APICORS:       "*",
				Type:          node.MasterNode,
				Verbosity:     4,
				Key:           "4de650ca1c8beae4ed6a4358087f50c01b51f5c0002ae9836c55039ca9818d0c", // 0x5c29518F6a6124a2BeE89253347c8295f604710A
			},
			Genesis: localFourNodesNetworkGenesis,
		},
		&node.NodeGalacticaFork{
			BaseNode: node.BaseNode{
				ID:            "node3",
				P2PListenPort: 8083,
				APIAddr:       "127.0.0.1:8183",
				APICORS:       "*",
				Type:          node.RegularNode,
				Key:           "1b310ea04afd6d14a8f142158873fc70bfd4ba12a19138cc5b309fce7c77105e", // 0x1b1c0055065b3ADee4B9a9e8297142Ba2cD34EfE
			},
			Genesis: localFourNodesNetworkGenesis,
		},
		&node.NodeGalacticaFork{
			BaseNode: node.BaseNode{
				ID:            "node4",
				P2PListenPort: 8084,
				APIAddr:       "127.0.0.1:8184",
				APICORS:       "*",
				Type:          node.MasterNode,
				Verbosity:     4,
				Key:           "c70dda88e779df10abbc7c5d37fbb3478c5cf8df2a70d6b0bfc551a5a9a17359", // 0x042306e116Dc301ecd7b83a04F4c8277Fbe41b6c
			},
			Genesis: localFourNodesNetworkGenesis,
		},
	},
}

var localFourNodesNetworkGenesis = &genesis.GalacticaForkGenesis{
	LaunchTime: 1703180212,
	GasLimit:   10000000,
	ExtraData:  "Local Four Nodes Network (Galactica)",
	Accounts: []thorgenesis.Account{
		{
			Address: thor.MustParseAddress("0x7567d83b7b8d80addcb281a71d54fc7b3364ffed"),
			Balance: convToHexOrDecimal256(consts.LargeBigValue),
			Energy:  convToHexOrDecimal256(consts.LargeBigValue),
		},
		{
			Address: *Account1.Address,
			Balance: convToHexOrDecimal256(consts.LargeBigValue),
			Energy:  convToHexOrDecimal256(consts.LargeBigValue),
		},
		{
			Address: *Account2.Address,
			Balance: convToHexOrDecimal256(consts.LargeBigValue),
			Energy:  convToHexOrDecimal256(consts.LargeBigValue),
		},
		{
			Address: *Account3.Address,
			Balance: convToHexOrDecimal256(consts.LargeBigValue),
			Energy:  convToHexOrDecimal256(consts.LargeBigValue),
		},
		{
			Address: *Account4.Address,
			Balance: convToHexOrDecimal256(consts.LargeBigValue),
			Energy:  convToHexOrDecimal256(consts.LargeBigValue),
		},
	},
	Authority: []thorgenesis.Authority{
		{
			MasterAddress:   *Account1.Address,
			EndorsorAddress: thor.MustParseAddress("0x7567d83b7b8d80addcb281a71d54fc7b3364ffed"),
			Identity:        thor.MustParseBytes32("0x000000000000000068747470733a2f2f636f6e6e65782e76656368612e696e2f"),
		},
		{
			MasterAddress:   thor.MustParseAddress("0x5c29518F6a6124a2BeE89253347c8295f604710A"),
			EndorsorAddress: thor.MustParseAddress("0x7567d83b7b8d80addcb281a71d54fc7b3364ffed"),
			Identity:        thor.MustParseBytes32("0x000000000000000068747470733a2f2f656e762e7665636861696e2e6f72672f"),
		},
		{
			MasterAddress:   thor.MustParseAddress("0x042306e116Dc301ecd7b83a04F4c8277Fbe41b6c"),
			EndorsorAddress: thor.MustParseAddress("0x7567d83b7b8d80addcb281a71d54fc7b3364ffed"),
			Identity:        thor.MustParseBytes32("0x0000000000000068747470733a2f2f617070732e7665636861696e2e6f72672f"),
		},
		{
			MasterAddress:   thor.MustParseAddress("0x0aeC31606e217895696771961de416Efa185Be66"),
			EndorsorAddress: thor.MustParseAddress("0x7567d83b7b8d80addcb281a71d54fc7b3364ffed"),
			Identity:        thor.MustParseBytes32("0x0000000000000068747470733a2f2f617070732e7665636861696e2e6f72672f"),
		},
	},
	Params: thorgenesis.Params{
		RewardRatio:         convToHexOrDecimal256(big.NewInt(300000000000000000)),
		BaseGasPrice:        convToHexOrDecimal256(big.NewInt(1000000000000000)),
		ProposerEndorsement: convToHexOrDecimal256(consts.LargeBigValue),
		ExecutorAddress:     &localThreeMasterEndorser,
	},
	Executor: thorgenesis.Executor{
		Approvers: []thorgenesis.Approver{
			{
				Address:  thor.MustParseAddress("0x199b836d8a57365baccd4f371c1fabb7be77d389"),
				Identity: thor.MustParseBytes32("0x00000000000067656e6572616c20707572706f736520626c6f636b636861696e"),
			},
		},
	},
	ForkConfig: &genesis.GalacticaForkConfig{
		VIP191:    0,
		ETH_CONST: 0,
		BLOCKLIST: 0,
		ETH_IST:   0,
		VIP214:    0,
		FINALITY:  0,
		GALACTICA: 0,
	},
}
