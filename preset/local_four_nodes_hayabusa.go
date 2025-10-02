package preset

import (
	"github.com/vechain/networkhub/internal/environments"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/network/node/genesis"
	"github.com/vechain/networkhub/thorbuilder"
	thorgenesis "github.com/vechain/thor/v2/genesis"
	"github.com/vechain/thor/v2/thor"
)

func LocalFourNodesHayabusaGenesis() *genesis.CustomGenesis {
	zero := uint32(0)
	return &genesis.CustomGenesis{
		LaunchTime: 1703180212,
		GasLimit:   10_000_000,
		ExtraData:  "Local Four Nodes Network",
		Accounts: []thorgenesis.Account{
			{
				Address: *SixNNAccount1.Address,
				Balance: convToHexOrDecimal256(LargeBigValue),
				Energy:  convToHexOrDecimal256(LargeBigValue),
			},
			{
				Address: *SixNNAccount2.Address,
				Balance: convToHexOrDecimal256(LargeBigValue),
				Energy:  convToHexOrDecimal256(LargeBigValue),
			},
			{
				Address: *SixNNAccount3.Address,
				Balance: convToHexOrDecimal256(LargeBigValue),
				Energy:  convToHexOrDecimal256(LargeBigValue),
			},
			{
				Address: *SixNNAccount4.Address,
				Balance: convToHexOrDecimal256(LargeBigValue),
				Energy:  convToHexOrDecimal256(LargeBigValue),
			},
		},
		Authority: []thorgenesis.Authority{
			{
				MasterAddress:   *SixNNAccount1.Address,
				EndorsorAddress: *SixNNAccount1.Address,
				Identity:        thor.MustParseBytes32("0x0000000000000068747470733a2f2f617070732e7665636861696e2e6f72672f"),
			},
			{
				MasterAddress:   *SixNNAccount2.Address,
				EndorsorAddress: *SixNNAccount2.Address,
				Identity:        thor.MustParseBytes32("0x0000000000000068747470733a2f2f617070732e7665636861696e2e6f72672f"),
			},
			{
				MasterAddress:   *SixNNAccount3.Address,
				EndorsorAddress: *SixNNAccount3.Address,
				Identity:        thor.MustParseBytes32("0x0000000000000068747470733a2f2f617070732e7665636861696e2e6f72672f"),
			},
			{
				MasterAddress:   *SixNNAccount4.Address,
				EndorsorAddress: *SixNNAccount4.Address,
				Identity:        thor.MustParseBytes32("0x0000000000000068747470733a2f2f617070732e7665636861696e2e6f72672f"),
			},
		},
		Params: thorgenesis.Params{},
		ForkConfig: &genesis.CustomGenesisForkConfig{
			ForkConfig: thor.ForkConfig{
				VIP191:    0,
				ETH_CONST: 0,
				BLOCKLIST: 0,
				ETH_IST:   0,
				VIP214:    0,
			},
			AdditionalFields: map[string]uint32{
				"HAYABUSA": 0,
			},
		},
		Config: &genesis.Config{
			BlockInterval:              10,
			EpochLength:                10,
			SeederInterval:             10,
			ValidatorEvictionThreshold: 40,
			EvictionCheckInterval:      10,
			LowStakingPeriod:           10,
			MediumStakingPeriod:        20,
			HighStakingPeriod:          40,
			CooldownPeriod:             10,
			HayabusaTP:                 &zero,
		},
	}
}

func LocalFourNodesHayabusa() *network.Network {
	thorBuilderCfg := thorbuilder.DefaultConfig()
	thorBuilderCfg.DownloadConfig.Branch = "release/hayabusa"
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
