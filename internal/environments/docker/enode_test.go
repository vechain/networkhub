package docker_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vechain/networkhub/internal/environments"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/preset"
)

func TestEnodeGeneration(t *testing.T) {
	// Create a test network configuration with different ports
	genesis := preset.LocalThreeNodesNetworkGenesis()
	presetNetwork := preset.LocalThreeNodesNetwork()

	networkCfg := &network.Network{
		Environment: environments.Docker,
		BaseID:      "enode-test",
		Nodes: []node.Config{
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
				APIAddr:       "0.0.0.0:8546",
				APICORS:       "*",
				P2PListenPort: 30304,
				Key:           presetNetwork.Nodes[1].GetKey(),
				Genesis:       genesis,
			},
			&node.BaseNode{
				ID:            "node3",
				ExecArtifact:  "vechain/thor:latest",
				DataDir:       "/home/thor",
				ConfigDir:     "/home/thor",
				APIAddr:       "0.0.0.0:8547",
				APICORS:       "*",
				P2PListenPort: 30305,
				Key:           presetNetwork.Nodes[2].GetKey(),
				Genesis:       genesis,
			},
		},
	}

	// TODO: Update to use overseer after docker manager is complete
	// env := docker.NewEnvironment(networkCfg)
	// assert.NotNil(t, env)

	// Test that we can generate enodes without starting the network
	// This will test our IP allocation fix
	enodes := make([]string, 0)

	// Generate enodes directly (this should assign IPs)
	for _, node := range networkCfg.Nodes {
		// Get IP that should be assigned during enode generation
		enode, err := node.Enode("192.168.1.10") // Example IP
		assert.NoError(t, err)
		enodes = append(enodes, enode)
		t.Logf("Generated enode for %s: %s", node.GetID(), enode)
	}

	// Verify all enodes are different (they should have different ports)
	assert.Len(t, enodes, 3)
	assert.NotEqual(t, enodes[0], enodes[1], "Node 1 and Node 2 should have different enodes")
	assert.NotEqual(t, enodes[1], enodes[2], "Node 2 and Node 3 should have different enodes")
	assert.NotEqual(t, enodes[0], enodes[2], "Node 1 and Node 3 should have different enodes")

	// Verify enodes contain different ports
	assert.Contains(t, enodes[0], ":30303", "Node 1 should use port 30303")
	assert.Contains(t, enodes[1], ":30304", "Node 2 should use port 30304")
	assert.Contains(t, enodes[2], ":30305", "Node 3 should use port 30305")

	t.Logf("✅ All enodes are unique with different ports!")
}
