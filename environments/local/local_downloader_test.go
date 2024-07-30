package local

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDownloader(t *testing.T) {
	t.Skip()
	downloader := NewDownloader("") // Pass a branch name here if needed

	err := downloader.DownloadRepo()
	require.NoError(t, err)

	thorPath, err := downloader.BuildThor()
	require.NoError(t, err)

	fmt.Printf("Built thor app at: %s\n", thorPath)
}
