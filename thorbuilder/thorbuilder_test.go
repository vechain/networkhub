package thorbuilder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder(t *testing.T) {
	// TODO fix: gh runners are not playing well with this particular test even when the same reusable feature is used on other tests
	// t.Run("Test Build Reusable", func(t *testing.T) {
	//	branch := "master"
	//	builder := New(branch, true)
	//
	//	// First download
	//	err := builder.Download()
	//	require.NoError(t, err)
	//
	//	// First build
	//	thorBinaryPath, err := builder.Build()
	//	require.NoError(t, err)
	//
	//	_, err = os.Stat(thorBinaryPath)
	//	require.NoError(t, err)
	//	assert.Equal(t, filepath.Join(builder.downloadPath, "bin", "thor"), thorBinaryPath)
	//
	//	// Second download should skip cloning
	//	err = builder.Download()
	//	require.NoError(t, err)
	//
	//	// Second build should skip building if the binary exists
	//	thorBinaryPath, err = builder.Build()
	//	require.NoError(t, err)
	//	assert.Equal(t, filepath.Join(builder.downloadPath, "bin", "thor"), thorBinaryPath)
	//})

	downloadedPath := ""

	t.Run("Test Build Non-Reusable", func(t *testing.T) {
		builder := New(DefaultConfig())

		err := builder.Download()
		require.NoError(t, err)

		thorBinaryPath, err := builder.Build()
		require.NoError(t, err)

		_, err = os.Stat(thorBinaryPath)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(builder.DownloadPath, "bin", "thor"), thorBinaryPath)
		downloadedPath = builder.DownloadPath
	})

	t.Run("Invalid Branch", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.DownloadConfig.Branch = "invalid-branch"
		builder := New(cfg)

		err := builder.Download()
		assert.Error(t, err)
	})

	t.Run("Build Without Download", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.BuildConfig = &BuildConfig{
			ExistingPath: downloadedPath,
			DebugBuild:   true,
		}
		builder := New(cfg)

		thorBinaryPath, err := builder.Build()
		assert.NoError(t, err)

		_, err = os.Stat(thorBinaryPath)
		assert.NoError(t, err)
	})
}
