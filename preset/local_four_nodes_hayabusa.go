package preset

import (
	"math/big"
	"time"

	"github.com/vechain/networkhub/genesisbuilder"
	"github.com/vechain/networkhub/internal/environments"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/network/node/genesis"
	"github.com/vechain/networkhub/thorbuilder"
	thorgenesis "github.com/vechain/thor/v2/genesis"
	"github.com/vechain/thor/v2/thor"
)

func LocalFourNodesHayabusaGenesis() *genesis.CustomGenesis {
	hayabusaTP := uint32(0)
	mbp := uint64(10)

	endorsement := new(big.Int)
	endorsement.SetString("fffffffffffffffffffffffffffffffffff", 16)

	return genesisbuilder.New(4).
		Accounts([]thorgenesis.Account{
			{Address: *SixNNAccount1.Address, Balance: convToHexOrDecimal256(LargeBigValue), Energy: convToHexOrDecimal256(LargeBigValue)},
			{Address: *SixNNAccount2.Address, Balance: convToHexOrDecimal256(LargeBigValue), Energy: convToHexOrDecimal256(LargeBigValue)},
			{Address: *SixNNAccount3.Address, Balance: convToHexOrDecimal256(LargeBigValue), Energy: convToHexOrDecimal256(LargeBigValue)},
			{Address: *SixNNAccount4.Address, Balance: convToHexOrDecimal256(LargeBigValue), Energy: convToHexOrDecimal256(LargeBigValue)},
		}).
		Stakers([]thorgenesis.Validator{
			{Master: *SixNNAccount1.Address, Endorser: *SixNNAccount1.Address},
			{Master: *SixNNAccount2.Address, Endorser: *SixNNAccount2.Address},
			{Master: *SixNNAccount3.Address, Endorser: *SixNNAccount3.Address},
			{Master: *SixNNAccount4.Address, Endorser: *SixNNAccount4.Address},
		}).
		Params(thorgenesis.Params{
			RewardRatio:         convToHexOrDecimal256(big.NewInt(300000000000000000)),
			BaseGasPrice:        convToHexOrDecimal256(big.NewInt(1000000000000000)),
			ProposerEndorsement: convToHexOrDecimal256(endorsement),
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
		ExtraData("Local Four Nodes Network").
		GenesisTimestampDelay(5 * time.Second).
		Build()
}

func LocalFourNodesHayabusa() *network.Network {
	thorBuilderCfg := thorbuilder.DefaultConfig()
	thorBuilderCfg.DownloadConfig.Branch = "master"
	thorBuilderCfg.BuildConfig.ReuseBinary = false

	gen := LocalFourNodesHayabusaGenesis()

	netwk := &network.Network{
		Environment: environments.Local,
		BaseID:      "hayabusa-four-nodes",
		ThorBuilder: thorBuilderCfg,
		Nodes: []node.Config{
			&node.BaseNode{
				ID:            "hayabusa-node-1",
				Key:           SixNNAccount1.PrivateKeyString(),
				Genesis:       gen,
				FakeExecution: false,
			},
			&node.BaseNode{
				ID:      "hayabusa-node-2",
				Key:     SixNNAccount2.PrivateKeyString(),
				Genesis: gen,
			},
			&node.BaseNode{
				ID:      "hayabusa-node-3",
				Key:     SixNNAccount3.PrivateKeyString(),
				Genesis: gen,
			},
			&node.BaseNode{
				ID:      "hayabusa-node-4",
				Key:     SixNNAccount4.PrivateKeyString(),
				Genesis: gen,
			},
		},
	}
	return netwk
}
