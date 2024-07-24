package node

import (
	"github.com/vechain/thor/v2/genesis"
)

type NodePreCoefFork struct {
	BaseNode
	Genesis *genesis.CustomGenesis `json:"genesis,omitempty"` //TODO would be nice to have validation in this format
}

func (n *NodePreCoefFork) GetGenesis() any {
	return n.Genesis
}
