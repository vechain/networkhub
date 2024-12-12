package thorbuilder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder(t *testing.T) {
	t.Run("Test Build", func(t *testing.T) {
		branch := "master"
		builder := New(branch)

		err := builder.Download()
		require.NoError(t, err)

		thorBinaryPath, err := builder.Build()
		require.NoError(t, err)

		_, err = os.Stat(thorBinaryPath)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(builder.downloadPath, "bin", "thor"), thorBinaryPath)
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("Invalid Branch", func(t *testing.T) {
		branch := "invalid-branch"
		builder := New(branch)

		err := builder.Download()
		assert.Error(t, err)
	})

	t.Run("Build Without Download", func(t *testing.T) {
		branch := "main"
		builder := New(branch)

		_, err := builder.Build()
		assert.Error(t, err)
	})
}
```

## Project Structure
- **Entrypoints**: Interface to interact with the framework. Currently implemented as an HTTP API server.
- **Actions**: Domain-specific language (DSL) that allows users to configure, start, stop, and request information on networks.
- **Environments**: Interface for running networks in different environments, with the Local environment currently implemented.
- **Thorbuilder Package**: Provides tools for building and managing the Thor binary from a given branch.
