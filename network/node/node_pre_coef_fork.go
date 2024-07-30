package node

import (
	"github.com/vechain/networkhub/network/node/genesis"
)

type NodePreCoefFork struct {
	BaseNode
	Genesis *genesis.PreCoefForkGenesis `json:"genesis,omitempty"` //TODO would be nice to have validation in this format
}

func (n *NodePreCoefFork) GetGenesis() any {
	return n.Genesis
}
