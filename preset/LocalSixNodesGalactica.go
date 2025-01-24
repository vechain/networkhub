package preset

import (
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/thorbuilder"
)

func LocalSixNodesGalacticaNetwork() *network.Network {
	thorBuilder := thorbuilder.New("release/galactica", true)
	err := thorBuilder.Download()
	if err != nil {
		panic(err)
	}
	thorBinPath, err := thorBuilder.Build()

	if err != nil {
		panic(err)
	}

	sixNodesNetwork := LocalSixNodesNetwork()

	sixNodesGalacticaGenesis := LocalSixNodesNetworkGenesis()
	sixNodesGalacticaGenesis.ForkConfig.AddField("GALACTICA", 0)
	// ensure the artifact path is set
	for _, node := range sixNodesNetwork.Nodes {
		node.SetGenesis(sixNodesGalacticaGenesis)
		node.SetExecArtifact(thorBinPath)
	}

	return sixNodesNetwork
}
