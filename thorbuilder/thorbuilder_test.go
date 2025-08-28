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

func TestDefaultConfig_WithWorkingDir_PrefersBuildConfig(t *testing.T) {
	t.Setenv("THOR_WORKING_DIR", "/tmp/thor-working-dir")
	t.Setenv("THOR_BRANCH", "release/hayabusa")

	cfg := DefaultConfig()

	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.BuildConfig)
	assert.Equal(t, "/tmp/thor-working-dir", cfg.BuildConfig.ExistingPath)
	assert.False(t, cfg.BuildConfig.DebugBuild)
	assert.Nil(t, cfg.DownloadConfig)
}

func TestDefaultConfig_WithBranch_UsesDownloadConfig(t *testing.T) {
	t.Setenv("THOR_WORKING_DIR", "")
	t.Setenv("THOR_BRANCH", "release/hayabusa")

	cfg := DefaultConfig()

	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.DownloadConfig)
	assert.Equal(t, "https://github.com/vechain/thor", cfg.DownloadConfig.RepoUrl)
	assert.Equal(t, "release/hayabusa", cfg.DownloadConfig.Branch)
	assert.True(t, cfg.DownloadConfig.IsReusable)
	assert.Nil(t, cfg.BuildConfig)
}

func TestDefaultConfig_NoEnv_UsesMasterBranch(t *testing.T) {
	t.Setenv("THOR_WORKING_DIR", "")
	t.Setenv("THOR_BRANCH", "")

	cfg := DefaultConfig()

	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.DownloadConfig)
	assert.Equal(t, "https://github.com/vechain/thor", cfg.DownloadConfig.RepoUrl)
	assert.Equal(t, "master", cfg.DownloadConfig.Branch)
	assert.True(t, cfg.DownloadConfig.IsReusable)
	assert.Nil(t, cfg.BuildConfig)
}

func TestDefaultConfig_WorkingDirOverridesBranch(t *testing.T) {
	t.Setenv("THOR_WORKING_DIR", "/tmp/thor-working-dir")
	t.Setenv("THOR_BRANCH", "release/hayabusa")

	cfg := DefaultConfig()

	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.BuildConfig)
	assert.Equal(t, "/tmp/thor-working-dir", cfg.BuildConfig.ExistingPath)
	assert.Nil(t, cfg.DownloadConfig)
}
