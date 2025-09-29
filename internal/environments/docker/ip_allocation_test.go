package docker_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vechain/networkhub/internal/environments"
	"github.com/vechain/networkhub/internal/environments/docker"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/preset"
)

func TestIPAllocationInEnodes(t *testing.T) {
	// Create a test network configuration
	genesis := preset.LocalThreeNodesNetworkGenesis()
	presetNetwork := preset.LocalThreeNodesNetwork()

	networkCfg := &network.Network{
		Environment: environments.Docker,
		BaseID:      "ip-test",
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
		},
	}

	// TODO: Update to use overseer after docker manager is complete
	// env := docker.NewEnvironment(networkCfg)
	// require.NotNil(t, env)

	// This should use the internal generateEnodes method and assign IPs
	// We can't call it directly since it's private, but we can test through StartNetwork preparation

	// Create an IP manager to check the allocation pattern
	ipManager := docker.NewIPManagerRandom()

	// Manually allocate IPs as the generateEnodes method would do
	ip1, err := ipManager.NextIP("node1")
	require.NoError(t, err)
	ip2, err := ipManager.NextIP("node2")
	require.NoError(t, err)

	t.Logf("Allocated IP for node1: %s", ip1)
	t.Logf("Allocated IP for node2: %s", ip2)

	// Verify IPs are different
	assert.NotEqual(t, ip1, ip2, "Node IPs should be different")

	// Verify IPs are valid and sequential
	assert.True(t, strings.HasSuffix(ip1, ".2"), "First IP should end with .2")
	assert.True(t, strings.HasSuffix(ip2, ".3"), "Second IP should end with .3")

	// Generate enodes with the allocated IPs
	enode1, err := networkCfg.Nodes[0].Enode(ip1)
	require.NoError(t, err)
	enode2, err := networkCfg.Nodes[1].Enode(ip2)
	require.NoError(t, err)

	t.Logf("Enode 1: %s", enode1)
	t.Logf("Enode 2: %s", enode2)

	// Verify enodes contain different IPs and ports
	assert.Contains(t, enode1, ip1+":30303", "Enode 1 should contain allocated IP and port")
	assert.Contains(t, enode2, ip2+":30304", "Enode 2 should contain allocated IP and port")
	assert.NotEqual(t, enode1, enode2, "Enodes should be completely different")

	t.Logf("✅ IP allocation and enode generation working correctly!")
}
