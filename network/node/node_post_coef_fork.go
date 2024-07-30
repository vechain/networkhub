package node

import (
	"github.com/vechain/networkhub/network/node/genesis"
)

type NodePostCoefFork struct {
	BaseNode
	Genesis *genesis.PostCoefForkGenesis `json:"genesis,omitempty"` //TODO would be nice to have validation in this format
}

func (n *NodePostCoefFork) GetGenesis() any {
	return n.Genesis
}
