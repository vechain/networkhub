package genesisbuilder

import (
	"math/big"

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

func DefaultStakers(n int) []genesis.Validator {
	devAccounts := genesis.DevAccounts()
	stakers := make([]genesis.Validator, n)
	for i := range n {
		stakers[i] = genesis.Validator{
			Master:   devAccounts[i].Address,
			Endorser: devAccounts[i].Address,
		}
	}
	return stakers
}

func DefaultParams(mbp uint64) *genesis.Params {
	endorsement := big.NewInt(0)
	endorsement.SetString("0xfffffffffffffffffffffffffffffffffff", 16)
	return &genesis.Params{
		RewardRatio:         (*genesis.HexOrDecimal256)(big.NewInt(300000000000000000)),
		BaseGasPrice:        (*genesis.HexOrDecimal256)(big.NewInt(1000000000000000)),
		ProposerEndorsement: (*genesis.HexOrDecimal256)(endorsement),
		MaxBlockProposers:   &mbp,
	}
}
