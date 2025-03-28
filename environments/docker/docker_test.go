package docker_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vechain/networkhub/environments/docker"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/preset"
	"github.com/vechain/networkhub/thorbuilder"
)

func TestDockerNetwork(t *testing.T) {
	genesis := preset.LocalThreeMasterNodesNetworkGenesis()
	presetNetwork := preset.LocalThreeMasterNodesNetwork()
	// Create a mock network configuration
	networkCfg := &network.Network{
		Environment: "docker",
		ID:          "test-id",
		Nodes: []node.Node{
			&node.BaseNode{
				ID:            "node1",
				ExecArtifact:  "vechain/thor:latest",
				DataDir:       "/home/thor",
				ConfigDir:     "/home/thor",
				APIAddr:       "0.0.0.0:8545",
				APICORS:       "*",
				P2PListenPort: 30303,
				Key:           presetNetwork.Nodes[0].GetKey(),
				Genesis:       genesis,
			},
			&node.BaseNode{
				ID:            "node2",
				ExecArtifact:  "vechain/thor:latest",
				DataDir:       "/home/thor",
				ConfigDir:     "/home/thor",
				APIAddr:       "0.0.0.0:8545",
				APICORS:       "*",
				P2PListenPort: 30303,
				Key:           presetNetwork.Nodes[1].GetKey(),
				Genesis:       genesis,
			},
			&node.BaseNode{
				ID:            "node3",
				ExecArtifact:  "vechain/thor:latest",
				DataDir:       "/home/thor",
				ConfigDir:     "/home/thor",
				APIAddr:       "0.0.0.0:8545",
				APICORS:       "*",
				P2PListenPort: 30303,
				Key:           presetNetwork.Nodes[2].GetKey(),
				Genesis:       genesis,
			},
		},
	}

	// Initialize Docker environment
	dockerEnv := docker.NewDockerEnv()
	assert.NotNil(t, dockerEnv)

	// Load configuration
	id, err := dockerEnv.LoadConfig(networkCfg)
	assert.NoError(t, err)
	assert.Equal(t, "dockertest-id", id)

	// Start network
	err = dockerEnv.StartNetwork()
	assert.NoError(t, err)

	time.Sleep(time.Minute)
	// Stop network
	err = dockerEnv.StopNetwork()
	assert.NoError(t, err)
}

func TestDockerHayabusaNetwork(t *testing.T) {
	repoUrl := "https://github.com/vechain/hayabusa"
	branch := "release/hayabusa"
	genesisUrl := "https://vechain.github.io/hayabusa-devnet/genesis.json"
	builder := thorbuilder.NewWithRepo(repoUrl, branch, true)
	dockerImage, err := builder.BuildDockerImage()
	assert.NoError(t, err)

	genesisJson, err := thorbuilder.FetchCustomGenesisFile(genesisUrl)
	assert.NoError(t, err)

	genesis, err := preset.LocalSixNodesNetworkCustomGenesis(*genesisJson)
	assert.NoError(t, err)
	presetNetwork := preset.LocalSixNodesHayabusaNetwork(*genesisJson, repoUrl)
	// Create a mock network configuration
	networkCfg := &network.Network{
		Environment: "docker",
		ID:          "test-id",
		Nodes: []node.Node{
			&node.BaseNode{
				ID:            "node1",
				ExecArtifact:  dockerImage,
				DataDir:       "/home/thor",
				ConfigDir:     "/home/thor",
				APIAddr:       "0.0.0.0:8545",
				APICORS:       "*",
				P2PListenPort: 30303,
				Key:           presetNetwork.Nodes[0].GetKey(),
				Genesis:       genesis,
			},
			&node.BaseNode{
				ID:            "node2",
				ExecArtifact:  dockerImage,
				DataDir:       "/home/thor",
				ConfigDir:     "/home/thor",
				APIAddr:       "0.0.0.0:8545",
				APICORS:       "*",
				P2PListenPort: 30303,
				Key:           presetNetwork.Nodes[1].GetKey(),
				Genesis:       genesis,
			},
			&node.BaseNode{
				ID:            "node3",
				ExecArtifact:  dockerImage,
				DataDir:       "/home/thor",
				ConfigDir:     "/home/thor",
				APIAddr:       "0.0.0.0:8545",
				APICORS:       "*",
				P2PListenPort: 30303,
				Key:           presetNetwork.Nodes[2].GetKey(),
				Genesis:       genesis,
			},
		},
	}

	// Initialize Docker environment
	dockerEnv := docker.NewDockerEnv()
	assert.NotNil(t, dockerEnv)

	// Load configuration
	id, err := dockerEnv.LoadConfig(networkCfg)
	assert.NoError(t, err)
	assert.Equal(t, "dockertest-id", id)

	// Start network
	err = dockerEnv.StartNetwork()
	assert.NoError(t, err)

	time.Sleep(time.Minute)
	// Stop network
	err = dockerEnv.StopNetwork()
	assert.NoError(t, err)
}
