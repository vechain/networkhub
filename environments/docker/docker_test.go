package docker_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vechain/networkhub/environments/docker"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/preset"
)

func TestDockerNetwork(t *testing.T) {
	t.Skip()
	// Create a mock network configuration
	networkCfg := &network.Network{
		Environment: "docker",
		ID:          "test-id",
		Nodes: []node.Node{
			&node.NodePreCoefFork{
				BaseNode: node.BaseNode{
					ID:            "node1",
					ExecArtifact:  "vechain/thor:latest",
					DataDir:       "/home/thor",
					ConfigDir:     "/home/thor",
					APIAddr:       "0.0.0.0:8545",
					APICORS:       "*",
					P2PListenPort: 30303,
					Key:           preset.LocalThreeMasterNodesNetwork.Nodes[0].GetKey(),
				},
				Genesis: preset.LocalThreeMasterNodesNetworkGenesis,
			},
			&node.NodePreCoefFork{
				BaseNode: node.BaseNode{
					ID:            "node2",
					ExecArtifact:  "vechain/thor:latest",
					DataDir:       "/home/thor",
					ConfigDir:     "/home/thor",
					APIAddr:       "0.0.0.0:8545",
					APICORS:       "*",
					P2PListenPort: 30303,
					Key:           preset.LocalThreeMasterNodesNetwork.Nodes[1].GetKey(),
				},
				Genesis: preset.LocalThreeMasterNodesNetworkGenesis,
			},
			&node.NodePreCoefFork{
				BaseNode: node.BaseNode{
					ID:            "node3",
					ExecArtifact:  "vechain/thor:latest",
					DataDir:       "/home/thor",
					ConfigDir:     "/home/thor",
					APIAddr:       "0.0.0.0:8545",
					APICORS:       "*",
					P2PListenPort: 30303,
					Key:           preset.LocalThreeMasterNodesNetwork.Nodes[2].GetKey(),
				},
				Genesis: preset.LocalThreeMasterNodesNetworkGenesis,
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

	time.Sleep(time.Hour)
	// Stop network
	err = dockerEnv.StopNetwork()
	assert.NoError(t, err)
}
