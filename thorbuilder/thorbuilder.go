package thorbuilder

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

type Builder struct {
	branch       string
	downloadPath string
	reusable     bool
	repoUrl      string
}

// New creates a new Builder instance for the specified branch.
// If reusable is true, it skips cloning if the directory exists and checks for the binary.
func New(branch string, reusable bool) *Builder {
	return NewWithRepo("https://github.com/vechain/thor", branch, reusable)
}

// NewWithRepo creates a new Builder instance for the specified branch and repo.
// If reusable is true, it skips cloning if the directory exists and checks for the binary.
func NewWithRepo(repoUrl string, branch string, reusable bool) *Builder {
	suffix := generateRandomSuffix(4)

	downloadPath := filepath.Join(os.TempDir(), fmt.Sprintf("thor_%s_%d_%s", branch, os.Getpid(), suffix))
	if reusable {
		downloadPath = filepath.Join(os.TempDir(), fmt.Sprintf("thor_%s_reusable", branch))
	}
	return &Builder{
		branch:       branch,
		reusable:     reusable,
		downloadPath: downloadPath,
		repoUrl:      repoUrl,
	}
}

// NewWithRepoPath allows for local testing and quicker builds
func NewWithRepoPath(repoUrl string, downloadPath string) *Builder {
	return &Builder{
		reusable:     true,
		downloadPath: downloadPath,
		repoUrl:      repoUrl,
	}
}

// Download clones the specified branch of the Thor repository into the downloadPath.
func (b *Builder) Download() error {
	if b.reusable {
		// Check if the folder exists and ensure it contains a cloned repository
		if _, err := os.Stat(filepath.Join(b.downloadPath, ".git")); err == nil {
			slog.Info("Reusable directory with repository exists: ", "path", b.downloadPath)
			cmd := exec.Command("git", "pull")
			if err := cmd.Run(); err != nil {
				slog.Warn("Failed to pull latest changes from repository", "error", err)
			}
			return nil
		}
	}

	if err := os.MkdirAll(b.downloadPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create download directory: %w", err)
	}

	args := make([]string, 0)
	args = append(args, "clone")
	if b.branch != "" {
		args = append(args, "--branch", b.branch)
	}
	args = append(args, "--depth", "1", b.repoUrl, b.downloadPath)

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

// Build runs the make command in the downloadPath and returns the path to the thor binary.
func (b *Builder) Build() (string, error) {
	if _, err := os.Stat(b.downloadPath); os.IsNotExist(err) {
		return "", fmt.Errorf("download directory does not exist: %s", b.downloadPath)
	}

	makeCmd := exec.Command("make")
	makeCmd.Dir = b.downloadPath
	// Capture output
	var stdout, stderr bytes.Buffer
	makeCmd.Stdout = &stdout
	makeCmd.Stderr = &stderr

	if err := makeCmd.Run(); err != nil {
		slog.Error("Make command failed",
			"stdout", stdout.String(),
			"stderr", stderr.String(),
			"error", err,
		)
		slog.Error("extra deets:", "str", makeCmd.String(), "path", makeCmd.Path, "dir", makeCmd.Dir)
		return "", fmt.Errorf("failed to build project: %w", err)
	}

	thorBinaryPath := filepath.Join(b.downloadPath, "bin", "thor")
	if _, err := os.Stat(thorBinaryPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("thor binary not found at expected path: %s", thorBinaryPath)
		}
		return "", fmt.Errorf("error checking thor binary path: %w", err)
	}

	return thorBinaryPath, nil
}

func (b *Builder) BuildDockerImage() (string, error) {
	if err := b.Download(); err != nil {
		return "", fmt.Errorf("failed to download repository: %w", err)
	}

	tag := fmt.Sprintf("test_%s_%s", b.branch, generateRandomSuffix(4))

	// Build the Docker image
	cmd := exec.Command("docker", "build", "-t", tag, b.downloadPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to build Docker image: %w", err)
	}

	slog.Info("Successfully built Docker image", "tag", tag)
	return tag, nil
}

func FetchCustomGenesisFile(genesisUrl string) (*string, error) {
	resp, err := http.Get(genesisUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch genesis file: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	genesis := string(body)
	return &genesis, err
}

// generateRandomSuffix returns a random hexadecimal string.
func generateRandomSuffix(n int) string {
	bytes := make([]byte, n)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
