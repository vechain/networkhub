package genesisbuilder

import (
	"time"

	"github.com/vechain/networkhub/network/node/genesis"
	thorgenesis "github.com/vechain/thor/v2/genesis"
	"github.com/vechain/thor/v2/thor"
)

type Overrider func(genesis *genesis.CustomGenesis)

type Builder struct {
	maxBlockProposers int
	accounts          []thorgenesis.Account
	authority         []thorgenesis.Authority
	params            *thorgenesis.Params
	executor          *thorgenesis.Executor
	forkConfig        *genesis.CustomGenesisForkConfig
	overrider         Overrider
	config            *genesis.Config
}

func New(maxBlockProposers int) *Builder {
	return &Builder{maxBlockProposers: maxBlockProposers}
}

func (b *Builder) Accounts(accounts []thorgenesis.Account) *Builder {
	b.accounts = accounts
	return b
}

func (b *Builder) Authority(authority []thorgenesis.Authority) *Builder {
	b.authority = authority
	return b
}

func (b *Builder) Params(params thorgenesis.Params) *Builder {
	b.params = &params
	return b
}

func (b *Builder) Executor(executor thorgenesis.Executor) *Builder {
	b.executor = &executor
	return b
}

func (b *Builder) ForkConfig(forkConfig *genesis.CustomGenesisForkConfig) *Builder {
	b.forkConfig = forkConfig
	return b
}

func (b *Builder) Overrider(overrider Overrider) *Builder {
	b.overrider = overrider
	return b
}

func (b *Builder) Build() *genesis.CustomGenesis {
	if len(b.accounts) == 0 {
		b.accounts = DefaultAccounts()
	}
	if len(b.authority) == 0 {
		b.authority = DefaultAuthority(b.maxBlockProposers)
	}
	if b.params == nil {
		b.params = DefaultParams(uint64(b.maxBlockProposers))
	}
	if b.executor == nil {
		b.executor = DefaultExecutor()
	}
	if b.forkConfig == nil {
		b.forkConfig = &genesis.CustomGenesisForkConfig{
			ForkConfig: thor.SoloFork,
		}
	}
	if b.config == nil {
		b.config = &genesis.Config{}
	}

	gene := &genesis.CustomGenesis{
		LaunchTime: uint64(time.Now().Unix()),
		GasLimit:   40_000_000,
		ExtraData:  "Custom Genesis",
		Accounts:   b.accounts,
		Authority:  b.authority,
		Params:     b.params,
		Executor:   *b.executor,
		ForkConfig: b.forkConfig,
		Config:     b.config,
	}

	if b.overrider != nil {
		b.overrider(gene)
	}

	return gene
}
