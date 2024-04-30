package local

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/vechain/networkhub/network/node"
)

type LocalNode struct {
	nodeCfg *node.Node
	cmdExec *exec.Cmd
	enodes  string
}

func NewLocalNode(nodeCfg *node.Node, enodes string) *LocalNode {
	return &LocalNode{
		nodeCfg: nodeCfg,
		enodes:  enodes,
	}
}

func (n *LocalNode) Start() error {
	if n.nodeCfg.Type == "masterNode" && n.nodeCfg.Key != "" {
		err := os.WriteFile(filepath.Join(n.nodeCfg.ConfigDir, "master.key"), []byte(n.nodeCfg.Key), 0644)
		if err != nil {
			return fmt.Errorf("failed to write to file %s: %w", filepath.Join(n.nodeCfg.ConfigDir, "master.key"), err)
		}
	}

	cmd := &exec.Cmd{
		Path: "/Users/pedro/go/src/github.com/vechain/thor/bin/thor",
		Args: []string{
			"thor",
			"--network", n.nodeCfg.Genesis,
			"--data-dir", n.nodeCfg.DataDir,
			"--config-dir", n.nodeCfg.ConfigDir,
			"--api-addr", n.nodeCfg.APIAddr,
			"--api-cors", n.nodeCfg.APICORS,
			"--p2p-port", fmt.Sprintf("%d", n.nodeCfg.P2PListenPort),
			"--bootnode", n.enodes,
		},
		Stdout: os.Stdout, // Directing stdout to the same stdout of the Go program
		Stderr: os.Stderr, // Directing stderr to the same stderr of the Go program
	}

	// Start the command and check for errors
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start thor command: %w", err)
	}

	n.cmdExec = cmd

	fmt.Println("Thor command executed successfully")
	return nil
}

func (n *LocalNode) Stop() error {
	// Send an interrupt signal
	if err := n.cmdExec.Process.Signal(os.Interrupt); err != nil {
		return fmt.Errorf("failed to send interrupt signal - %w", err)
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Wait for the command to finish with a timeout
	done := make(chan error, 1)
	go func() {
		done <- n.cmdExec.Wait()
	}()

	select {
	case <-ctx.Done():
		if err := n.cmdExec.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process - %w", err)
		}
		fmt.Println("Process killed as timeout reached")
	case err := <-done:
		if err != nil {
			return fmt.Errorf("process exited with error - %w", err)
		}
		fmt.Println("Process stopped gracefully")
	}
	return nil
}
