package thorbuilder

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

type Builder struct {
	branch       string
	downloadPath string
	reusable     bool
}

// New creates a new Builder instance for the specified branch.
// If reusable is true, it skips cloning if the directory exists and checks for the binary.
func New(branch string, reusable bool) *Builder {
	suffix := generateRandomSuffix(4)

	downloadPath := filepath.Join(os.TempDir(), fmt.Sprintf("thor_%s_%d_%s", branch, os.Getpid(), suffix))
	if reusable {
		downloadPath = filepath.Join(os.TempDir(), fmt.Sprintf("thor_%s_reusable", branch))
	}
	return &Builder{
		branch:       branch,
		reusable:     reusable,
		downloadPath: downloadPath,
	}
}

// Download clones the specified branch of the Thor repository into the downloadPath.
func (b *Builder) Download() error {
	if b.reusable {
		// Check if the folder exists and ensure it contains a cloned repository
		if _, err := os.Stat(filepath.Join(b.downloadPath, ".git")); err == nil {
			slog.Info("Reusable directory with repository exists: ", "path", b.downloadPath)
			return nil
		}
	}

	if err := os.MkdirAll(b.downloadPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create download directory: %w", err)
	}

	repoURL := "https://github.com/vechain/thor"
	cmd := exec.Command("git", "clone", "--branch", b.branch, "--depth", "1", repoURL, b.downloadPath)
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

	if b.reusable {
		// Check if the binary exists and if it does return the path
		thorBinaryPath := filepath.Join(b.downloadPath, "bin", "thor")
		if _, err := os.Stat(thorBinaryPath); err == nil {
			slog.Info("Reusable binary exists: ", "path", thorBinaryPath)
			return thorBinaryPath, nil
		}
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

// generateRandomSuffix returns a random hexadecimal string.
func generateRandomSuffix(n int) string {
	bytes := make([]byte, n)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
