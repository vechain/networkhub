package docker_test

import (
	"testing"
	"time"

	"github.com/vechain/thor/v2/thorclient"

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
				ID:             "node1",
				ExecArtifact:   "vechain/thor:latest",
				DataDir:        "/home/thor",
				ConfigDir:      "/home/thor",
				APIAddr:        "0.0.0.0:8545",
				APICORS:        "*",
				P2PListenPort:  30303,
				Key:            presetNetwork.Nodes[0].GetKey(),
				Genesis:        genesis,
				AdditionalArgs: map[string]string{"api-allowed-tracers": "all"},
			},
			&node.BaseNode{
				ID:             "node2",
				ExecArtifact:   "vechain/thor:latest",
				DataDir:        "/home/thor",
				ConfigDir:      "/home/thor",
				APIAddr:        "0.0.0.0:8545",
				APICORS:        "*",
				P2PListenPort:  30303,
				Key:            presetNetwork.Nodes[1].GetKey(),
				Genesis:        genesis,
				AdditionalArgs: map[string]string{"api-allowed-tracers": "all"},
			},
			&node.BaseNode{
				ID:             "node3",
				ExecArtifact:   "vechain/thor:latest",
				DataDir:        "/home/thor",
				ConfigDir:      "/home/thor",
				APIAddr:        "0.0.0.0:8545",
				APICORS:        "*",
				P2PListenPort:  30303,
				Key:            presetNetwork.Nodes[2].GetKey(),
				Genesis:        genesis,
				AdditionalArgs: map[string]string{"api-allowed-tracers": "all"},
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

	t.Cleanup(func() {
		time.Sleep(time.Minute)
		// Stop network
		err = dockerEnv.StopNetwork()
		assert.NoError(t, err)
	})

	// Start network
	err = dockerEnv.StartNetwork()
	assert.NoError(t, err)

	err = networkCfg.HealthCheck(1, time.Minute)
	assert.NoError(t, err)

	// test additional args
	client := thorclient.New(networkCfg.Nodes[0].GetHTTPAddr())
	_, statusCode, err := client.RawHTTPClient().RawHTTPPost("/debug/tracers/call", []byte(`{
  "value": "0x0",
  "to": "0x0000000000000000000000000000456E65726779",
  "data": "0xa9059cbb0000000000000000000000000f872421dc479f3c11edd89512731814d0598db50000000000",
  "gas": 50000,
  "gasPrice": "1000000000000000",
  "caller": "0x7567d83b7b8d80addcb281a71d54fc7b3364ffed",
  "provedWork": "1000",
  "gasPayer": "0xd3ae78222beadb038203be21ed5ce7c9b1bff602",
  "expiration": 1000,
  "blockRef": "0x00000000851caf3c",
  "name": "call"
}`))
	assert.Equal(t, 200, statusCode)
}

func TestDockerHayabusaNetwork(t *testing.T) {
	t.Skip()
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
