package node

import (
	"github.com/vechain/networkhub/network/node/genesis"
	"github.com/vechain/networkhub/thorbuilder"
)

type NodeGalacticaFork struct {
	BaseNode
	Genesis *genesis.GalacticaForkGenesis `json:"genesis,omitempty"` //TODO would be nice to have validation in this format
	ExecArtifact  string `json:"execArtifact"` // Enforcing the release/galactica branch
}

func (n *NodeGalacticaFork) GetGenesis() any {
	return n.Genesis
}

func (n *NodeGalacticaFork) GetExecArtifact() string {
	thorBuilder := thorbuilder.New("release/galactica")
	if err := thorBuilder.Download(); err != nil {
		panic(err)
	}

	path, err := thorBuilder.Build()
	if err != nil {
		panic(err)
	}

	return path
}
