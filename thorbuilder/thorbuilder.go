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

type Config struct {
	DownloadConfig *DownloadConfig
	BuildConfig    *BuildConfig
}
type DownloadConfig struct {
	RepoUrl    string
	Branch     string
	IsReusable bool
}

type BuildConfig struct {
	ExistingPath string
	DebugBuild   bool
}

type Builder struct {
	config       *Config
	DownloadPath string
}

func DefaultConfig() *Config {
	branch := "master"
	env := os.Getenv("THOR_BRANCH")
	if env != "" {
		branch = env
	}

	return &Config{
		DownloadConfig: &DownloadConfig{
			RepoUrl:    "https://github.com/vechain/thor",
			Branch:     branch,
			IsReusable: true,
		},
		BuildConfig: nil,
	}
}

// New creates a new Builder instance for the specified branch.
// If reusable is true, it skips cloning if the directory exists and checks for the binary.
func New(cfg *Config) *Builder {
	downloadPath := ""
	if cfg.DownloadConfig != nil {
		suffix := generateRandomSuffix(4)

		downloadPath = filepath.Join(os.TempDir(), fmt.Sprintf("thor_%s_%d_%s", cfg.DownloadConfig.Branch, os.Getpid(), suffix))
		if cfg.DownloadConfig.IsReusable {
			downloadPath = filepath.Join(os.TempDir(), fmt.Sprintf("thor_%s_reusable", cfg.DownloadConfig.Branch))
		}
	}

	if cfg.BuildConfig != nil {
		downloadPath = cfg.BuildConfig.ExistingPath
	}

	return &Builder{
		config:       cfg,
		DownloadPath: downloadPath,
	}
}

// Download clones the specified branch of the Thor repository into the downloadPath.
func (b *Builder) Download() error {
	if b.config.DownloadConfig == nil {
		slog.Info("Skipping Download... No download config was provided")
		return nil
	}
	if b.config.DownloadConfig.IsReusable {
		// Check if the folder exists and ensure it contains a cloned repository
		if _, err := os.Stat(filepath.Join(b.DownloadPath, ".git")); err == nil {
			slog.Info("Reusable directory with repository exists: ", "path", b.DownloadPath)
			cmd := exec.Command("git", "pull")
			cmd.Dir = b.DownloadPath
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				slog.Warn("Failed to pull latest changes from repository", "error", err)
			}
			return nil
		}
	}

	if err := os.MkdirAll(b.DownloadPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create download directory: %w", err)
	}

	args := make([]string, 0)
	args = append(args, "clone")
	if b.config.DownloadConfig.Branch != "" {
		args = append(args, "--branch", b.config.DownloadConfig.Branch)
	}
	args = append(args, "--depth", "1", b.config.DownloadConfig.RepoUrl, b.DownloadPath)

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
	if _, err := os.Stat(b.DownloadPath); os.IsNotExist(err) {
		return "", fmt.Errorf("download directory does not exist: %s", b.DownloadPath)
	}

	var cmd *exec.Cmd

	if b.config.BuildConfig != nil && b.config.BuildConfig.DebugBuild {
		cmd = exec.Command(
			"go", "build",
			"-gcflags=all=-N -l", // Disable optimizations. Useful for debugging.
			"-v",
			"-o", "./bin/thor",
			"-ldflags", "-X main.version=0.0.0 -X main.gitCommit=sha -X main.gitTag=v0.0.0 -X main.copyrightYear=2025",
			"./cmd/thor",
		)
	} else {
		cmd = exec.Command("make")
	}
	cmd.Dir = b.DownloadPath
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		slog.Error("Make command failed",
			"stdout", stdout.String(),
			"stderr", stderr.String(),
			"error", err,
		)
		slog.Error("extra deets:", "str", cmd.String(), "path", cmd.Path, "dir", cmd.Dir)
		return "", fmt.Errorf("failed to build project: %w", err)
	}

	thorBinaryPath := filepath.Join(b.DownloadPath, "bin", "thor")
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

	tag := fmt.Sprintf("test_%s_%s", b.config.DownloadConfig.Branch, generateRandomSuffix(4))

	// Build the Docker image
	cmd := exec.Command("docker", "build", "-t", tag, b.DownloadPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to build Docker image: %w", err)
	}

	slog.Info("Successfully built Docker image", "tag", tag)
	return tag, nil
}

// generateRandomSuffix returns a random hexadecimal string.
func generateRandomSuffix(n int) string {
	bytes := make([]byte, n)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
