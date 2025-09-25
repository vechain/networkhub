package local

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vechain/networkhub/preset"
)

func TestLocalInvalidExecArtifact(t *testing.T) {
	networkCfg := preset.LocalThreeMasterNodesNetwork()

	networkCfg.Nodes[0].SetExecArtifact("/some_fake_dir")

	// Test local environment directly
	env := NewEnvironment(networkCfg)
	err := env.StartNetwork()
	require.Error(t, err)

	require.ErrorContains(t, err, "exec artifact path /some_fake_dir does not exist")
}
