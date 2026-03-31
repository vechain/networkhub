package genesisbuilder

import (
	"time"

	"github.com/vechain/networkhub/network/node/genesis"
	thorgenesis "github.com/vechain/thor/v2/genesis"
	"github.com/vechain/thor/v2/thor"
)

type Overrider func(genesis *genesis.CustomGenesis)

type Builder struct {
	maxBlockProposers     int
	accounts              []thorgenesis.Account
	stakers               []thorgenesis.Validator
	params                *thorgenesis.Params
	forkConfig            *genesis.CustomGenesisForkConfig
	overrider             Overrider
	genesisTimestampDelay time.Duration
	config                *genesis.Config
	gasLimit              uint64
	extraData             string
	extraDataSet          bool
}

func New(maxBlockProposers int) *Builder {
	return &Builder{maxBlockProposers: maxBlockProposers}
}

func (b *Builder) Config(config *genesis.Config) *Builder {
	b.config = config
	return b
}

func (b *Builder) Accounts(accounts []thorgenesis.Account) *Builder {
	b.accounts = accounts
	return b
}

func (b *Builder) Stakers(stakers []thorgenesis.Validator) *Builder {
	b.stakers = stakers
	return b
}

func (b *Builder) Params(params thorgenesis.Params) *Builder {
	b.params = &params
	return b
}

func (b *Builder) ForkConfig(forkConfig *genesis.CustomGenesisForkConfig) *Builder {
	b.forkConfig = forkConfig
	return b
}

func (b *Builder) GasLimit(gasLimit uint64) *Builder {
	b.gasLimit = gasLimit
	return b
}

func (b *Builder) ExtraData(extraData string) *Builder {
	b.extraData = extraData
	b.extraDataSet = true
	return b
}

func (b *Builder) Overrider(overrider Overrider) *Builder {
	b.overrider = overrider
	return b
}

func (b *Builder) GenesisTimestampDelay(d time.Duration) *Builder {
	b.genesisTimestampDelay = d
	return b
}

func (b *Builder) Build() *genesis.CustomGenesis {
	if len(b.accounts) == 0 {
		b.accounts = DefaultAccounts()
	}
	if len(b.stakers) == 0 {
		b.stakers = DefaultStakers(b.maxBlockProposers)
	}
	if b.params == nil {
		b.params = DefaultParams(uint64(b.maxBlockProposers))
	}
	if b.forkConfig == nil {
		b.forkConfig = &genesis.CustomGenesisForkConfig{
			ForkConfig: thor.SoloFork,
		}
	}
	if b.config == nil {
		b.config = &genesis.Config{}
	}
	if b.gasLimit == 0 {
		b.gasLimit = 40_000_000
	}
	extraData := ""
	if b.extraDataSet {
		extraData = b.extraData
	}

	gene := &genesis.CustomGenesis{
		CustomGenesis: &thorgenesis.CustomGenesis{
			LaunchTime: uint64(time.Now().Add(b.genesisTimestampDelay).Unix()),
			GasLimit:   b.gasLimit,
			ExtraData:  extraData,
			Accounts:   b.accounts,
			Stakers:    b.stakers,
			Params:     *b.params,
			Executor:   thorgenesis.Executor{},
		},
		ForkConfig: b.forkConfig,
		Config:     b.config,
	}

	if b.overrider != nil {
		b.overrider(gene)
	}

	return gene
}
