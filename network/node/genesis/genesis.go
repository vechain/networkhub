package genesis

import (
	"github.com/vechain/thor/v2/genesis"
)

// CustomGenesis is user customized genesis
type CustomGenesis struct {
	LaunchTime uint64              `json:"launchTime"`
	GasLimit   uint64              `json:"gaslimit"`
	ExtraData  string              `json:"extraData"`
	Accounts   []genesis.Account   `json:"accounts"`
	Authority  []genesis.Authority `json:"authority"`
	Params     genesis.Params      `json:"params"`
	Executor   genesis.Executor    `json:"executor"`
	ForkConfig *ForkConfig         `json:"forkConfig"`
}

type ForkConfig struct {
	VIP191     uint32
	ETH_CONST  uint32
	BLOCKLIST  uint32
	ETH_IST    uint32
	VIP214     uint32
	FINALITY   uint32
	VIPGASCOEF uint32
}