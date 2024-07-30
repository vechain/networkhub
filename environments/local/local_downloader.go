package local

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Downloader struct {
	repoURL     string
	branch      string
	downloadDir string
}

func NewDownloader(branch string) *Downloader {
	if branch == "" {
		branch = "master"
	}
	downloadDir := filepath.Join(os.TempDir(), "thor_repo")
	return &Downloader{
		repoURL:     "https://github.com/vechain/thor",
		branch:      branch,
		downloadDir: downloadDir,
	}
}

func (d *Downloader) CloneAndBuildThor() (string, error) {
	if err := d.DownloadRepo(); err != nil {
		return "", fmt.Errorf("unable to clone repo: %w", err)
	}

	return d.BuildThor()
}

func (d *Downloader) DownloadRepo() error {
	repoDir := filepath.Join(d.downloadDir, "thor")

	// Check if the repo directory already exists
	if _, err := os.Stat(repoDir); !os.IsNotExist(err) {
		fmt.Printf("Repository already exists at %s, skipping download.\n", repoDir)
		return nil
	}

	_, err := git.PlainClone(repoDir, false, &git.CloneOptions{
		URL:           d.repoURL,
		ReferenceName: plumbing.NewBranchReferenceName(d.branch), // git.ReferenceName(fmt.Sprintf("refs/heads/%s", d.branch)),
		SingleBranch:  true,
		Depth:         1,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}
	return nil
}

func (d *Downloader) BuildThor() (string, error) {
	repoDir := filepath.Join(d.downloadDir, "thor")
	thorPath := filepath.Join(repoDir, "bin", "thor") // Assuming 'bin/thor' is the path to the built thor binary

	// Check if the thor binary already exists
	if _, err := os.Stat(thorPath); !os.IsNotExist(err) {
		fmt.Printf("Thor binary already exists at %s, skipping build.\n", thorPath)
		return thorPath, nil
	}

	cmd := exec.Command("make", "thor", "cmd")
	cmd.Dir = repoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to build thor: %w", err)
	}

	return thorPath, nil
}
