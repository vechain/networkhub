package preset

import (
	"math/big"

	"github.com/vechain/networkhub/internal/environments"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/network/node/genesis"
	thorgenesis "github.com/vechain/thor/v2/genesis"
	"github.com/vechain/thor/v2/thor"
)

func LocalThreeNodesNetwork() *network.Network {
	genesis := LocalThreeNodesNetworkGenesis()
	return &network.Network{
		BaseID:      "threeMaster",
		Environment: environments.Local,
		Nodes: []node.Config{
			&node.BaseNode{
				ID:            "node1",
				P2PListenPort: 8031,
				APIAddr:       "127.0.0.1:8131",
				Key:           "01a4107bfb7d5141ec519e75788c34295741a1eefbfe460320efd2ada944071e", // 0x61fF580B63D3845934610222245C116E013717ec
				Genesis:       genesis,
			},
			&node.BaseNode{
				ID:            "node2",
				P2PListenPort: 8032,
				APIAddr:       "127.0.0.1:8132",
				Key:           "7072249b800ddac1d29a3cd06468cc1a917cbcd110dde358a905d03dad51748d", // 0x327931085B4cCbCE0baABb5a5E1C678707C51d90
				Genesis:       genesis,
			},
			&node.BaseNode{
				ID:            "node3",
				P2PListenPort: 8033,
				APIAddr:       "127.0.0.1:8133",
				Key:           "c55455943bf026dc44fcf189e8765eb0587c94e66029d580bae795386c0b737a", // 0x084E48c8AE79656D7e27368AE5317b5c2D6a7497
				Genesis:       genesis,
			},
		},
	}
}

var localThreeMasterEndorser = thor.MustParseAddress("0x0000000000000000000000004578656375746f72")

func LocalThreeNodesNetworkGenesis() *genesis.CustomGenesis {
	return &genesis.CustomGenesis{
		LaunchTime: 1703180212,
		GasLimit:   10000000,
		ExtraData:  "",
		Accounts: []thorgenesis.Account{
			{
				Address: thor.MustParseAddress("0x7567d83b7b8d80addcb281a71d54fc7b3364ffed"),
				Balance: convToHexOrDecimal256(LargeBigValue),
				Energy:  convToHexOrDecimal256(LargeBigValue),
			},
			{
				Address: thor.MustParseAddress("0x61fF580B63D3845934610222245C116E013717ec"),
				Balance: convToHexOrDecimal256(LargeBigValue),
				Energy:  convToHexOrDecimal256(LargeBigValue),
			},
			{
				Address: thor.MustParseAddress("0x327931085B4cCbCE0baABb5a5E1C678707C51d90"),
				Balance: convToHexOrDecimal256(LargeBigValue),
				Energy:  convToHexOrDecimal256(LargeBigValue),
			},
			{
				Address: thor.MustParseAddress("0x084E48c8AE79656D7e27368AE5317b5c2D6a7497"),
				Balance: convToHexOrDecimal256(LargeBigValue),
				Energy:  convToHexOrDecimal256(LargeBigValue),
			},
		},
		Authority: []thorgenesis.Authority{
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
		Params: thorgenesis.Params{
			RewardRatio:         convToHexOrDecimal256(big.NewInt(300000000000000000)),
			BaseGasPrice:        convToHexOrDecimal256(big.NewInt(1000000000000000)),
			ProposerEndorsement: convToHexOrDecimal256(LargeBigValue),
		},
		Executor: thorgenesis.Executor{
			Approvers: []thorgenesis.Approver{
				{
					Address:  thor.MustParseAddress("0x199b836d8a57365baccd4f371c1fabb7be77d389"),
					Identity: thor.MustParseBytes32("0x00000000000067656e6572616c20707572706f736520626c6f636b636861696e"),
				},
			},
		},
		ForkConfig: &genesis.CustomGenesisForkConfig{
			ForkConfig: thor.ForkConfig{
				VIP191:    0,
				ETH_CONST: 0,
				BLOCKLIST: 0,
				ETH_IST:   0,
				VIP214:    0,
			},
		},
	}
}
