package genesisbuilder

import (
	"math/big"

	"github.com/vechain/networkhub/utils/datagen"
	"github.com/vechain/thor/v2/genesis"
	"github.com/vechain/thor/v2/thor"
)

func DefaultAccounts() []genesis.Account {
	devAccounts := genesis.DevAccounts()
	accounts := make([]genesis.Account, len(devAccounts))

	tenBillion := big.NewInt(10e9)
	tenBillion = tenBillion.Mul(tenBillion, big.NewInt(1e18))

	for i, account := range devAccounts {
		accounts[i] = genesis.Account{
			Address: account.Address,
			Balance: (*genesis.HexOrDecimal256)(tenBillion),
			Energy:  (*genesis.HexOrDecimal256)(tenBillion),
			Code:    "0x",
			Storage: make(map[string]thor.Bytes32),
		}
	}

	return accounts
}

func DefaultAuthority(amount int) []genesis.Authority {
	devAccounts := genesis.DevAccounts()
	accounts := make([]genesis.Authority, amount)

	for i := range amount {
		accounts[i] = genesis.Authority{
			MasterAddress:   devAccounts[i].Address,
			EndorsorAddress: devAccounts[i].Address,
			Identity:        datagen.RandKey(),
		}
	}

	return accounts
}

func DefaultParams(mbp uint64) *genesis.Params {
	executor := thor.MustParseAddress("0x0000000000000000000000004578656375746f72")
	endorsement := big.NewInt(0)
	endorsement.SetString("0xfffffffffffffffffffffffffffffffffff", 16)
	return &genesis.Params{
		RewardRatio:         (*genesis.HexOrDecimal256)(big.NewInt(300000000000000000)),
		BaseGasPrice:        (*genesis.HexOrDecimal256)(big.NewInt(1000000000000000)),
		ProposerEndorsement: (*genesis.HexOrDecimal256)(endorsement),
		ExecutorAddress:     &executor,
		MaxBlockProposers:   &mbp,
	}
}

func DefaultExecutor() *genesis.Executor {
	return &genesis.Executor{
		Approvers: []genesis.Approver{
			{
				Address:  genesis.DevAccounts()[0].Address,
				Identity: datagen.RandKey(),
			},
		},
	}
}
