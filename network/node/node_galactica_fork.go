package node

import (
	"github.com/vechain/networkhub/network/node/genesis"
)

type NodeGalacticaFork struct {
	BaseNode
	Genesis *genesis.GalacticaGenesis `json:"genesis,omitempty"` //TODO would be nice to have validation in this format
}

func (n *NodeGalacticaFork) GetGenesis() any {
	return n.Genesis
}
