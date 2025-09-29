package local_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vechain/networkhub/internal/environments/overseer"
	"github.com/vechain/networkhub/preset"
)

func TestLocalInvalidExecArtifact(t *testing.T) {
	networkCfg := preset.LocalThreeMasterNodesNetwork()

	networkCfg.Nodes[0].SetExecArtifact("/some_fake_dir")

	// Test overseer with local environment
	env, err := overseer.New(networkCfg)
	require.NoError(t, err)

	err = env.StartNetwork()
	require.Error(t, err)

	require.ErrorContains(t, err, "artifact path /some_fake_dir does not exist for node")
}
