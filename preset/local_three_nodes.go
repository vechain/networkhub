package preset

import (
	"math/big"
	"time"

	"github.com/vechain/networkhub/genesisbuilder"
	"github.com/vechain/networkhub/internal/environments"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/network/node/genesis"
	thorgenesis "github.com/vechain/thor/v2/genesis"
	"github.com/vechain/thor/v2/thor"
)

func LocalThreeNodesNetwork() *network.Network {
	gen := LocalThreeNodesNetworkGenesis()
	return &network.Network{
		BaseID:      "threeMaster",
		Environment: environments.Local,
		Nodes: []node.Config{
			&node.BaseNode{
				ID:            "node1",
				P2PListenPort: 8031,
				APIAddr:       "127.0.0.1:8131",
				Key:           "01a4107bfb7d5141ec519e75788c34295741a1eefbfe460320efd2ada944071e", // 0x61fF580B63D3845934610222245C116E013717ec
				Genesis:       gen,
			},
			&node.BaseNode{
				ID:            "node2",
				P2PListenPort: 8032,
				APIAddr:       "127.0.0.1:8132",
				Key:           "7072249b800ddac1d29a3cd06468cc1a917cbcd110dde358a905d03dad51748d", // 0x327931085B4cCbCE0baABb5a5E1C678707C51d90
				Genesis:       gen,
			},
			&node.BaseNode{
				ID:            "node3",
				P2PListenPort: 8033,
				APIAddr:       "127.0.0.1:8133",
				Key:           "c55455943bf026dc44fcf189e8765eb0587c94e66029d580bae795386c0b737a", // 0x084E48c8AE79656D7e27368AE5317b5c2D6a7497
				Genesis:       gen,
			},
		},
	}
}

func LocalThreeNodesNetworkGenesis() *genesis.CustomGenesis {
	hayabusaTP := uint32(0)
	mbp := uint64(3)

	node1 := thor.MustParseAddress("0x61fF580B63D3845934610222245C116E013717ec")
	node2 := thor.MustParseAddress("0x327931085B4cCbCE0baABb5a5E1C678707C51d90")
	node3 := thor.MustParseAddress("0x084E48c8AE79656D7e27368AE5317b5c2D6a7497")

	return genesisbuilder.New(3).
		Accounts([]thorgenesis.Account{
			{
				Address: thor.MustParseAddress("0x7567d83b7b8d80addcb281a71d54fc7b3364ffed"),
				Balance: convToHexOrDecimal256(LargeBigValue),
				Energy:  convToHexOrDecimal256(LargeBigValue),
			},
			{Address: node1, Balance: convToHexOrDecimal256(LargeBigValue), Energy: convToHexOrDecimal256(LargeBigValue)},
			{Address: node2, Balance: convToHexOrDecimal256(LargeBigValue), Energy: convToHexOrDecimal256(LargeBigValue)},
			{Address: node3, Balance: convToHexOrDecimal256(LargeBigValue), Energy: convToHexOrDecimal256(LargeBigValue)},
		}).
		Stakers([]thorgenesis.Validator{
			{Master: node1, Endorser: node1},
			{Master: node2, Endorser: node2},
			{Master: node3, Endorser: node3},
		}).
		Params(thorgenesis.Params{
			RewardRatio:         convToHexOrDecimal256(big.NewInt(300000000000000000)),
			BaseGasPrice:        convToHexOrDecimal256(big.NewInt(1000000000000000)),
			ProposerEndorsement: convToHexOrDecimal256(LargeBigValue),
			MaxBlockProposers:   &mbp,
		}).
		ForkConfig(&genesis.CustomGenesisForkConfig{ForkConfig: thor.SoloFork}).
		Config(&genesis.Config{
			BlockInterval:              10,
			EpochLength:                10,
			SeederInterval:             10,
			ValidatorEvictionThreshold: 40,
			EvictionCheckInterval:      10,
			LowStakingPeriod:           10,
			MediumStakingPeriod:        20,
			HighStakingPeriod:          40,
			CooldownPeriod:             10,
			HayabusaTP:                 &hayabusaTP,
		}).
		GasLimit(10_000_000).
		GenesisTimestampDelay(5 * time.Second).
		Build()
}
