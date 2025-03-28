package preset

import (
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/thorbuilder"
)

func LocalSixNodesHayabusaNetwork(customGenesisJson string, repoUrl string) *network.Network {
	thorBuilder := thorbuilder.NewWithRepo(repoUrl, "release/hayabusa", true)
	err := thorBuilder.Download()
	if err != nil {
		panic(err)
	}
	thorBinPath, err := thorBuilder.Build()
	if err != nil {
		panic(err)
	}

	sixNodesHayabusaGenesis, err := LocalSixNodesNetworkCustomGenesis(customGenesisJson)
	if err != nil {
		panic(err)
	}

	sixNodesHayabusaGenesis.ForkConfig.AddField("FINALITY", 0)
	sixNodesHayabusaGenesis.ForkConfig.AddField("HAYABUSA", 12)
	sixNodesHayabusaGenesis.ForkConfig.AddField("HAYABUSA_TP", 12)
	sixNodesHayabusaNetwork := LocalSixNodesNetworkWithGenesis(sixNodesHayabusaGenesis)

	for _, node := range sixNodesHayabusaNetwork.Nodes {
		node.SetExecArtifact(thorBinPath)
	}

	return sixNodesHayabusaNetwork
}
