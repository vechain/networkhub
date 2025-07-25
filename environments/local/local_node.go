package local

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vechain/networkhub/network/node"
	nodegenesis "github.com/vechain/networkhub/network/node/genesis"
)

type Node struct {
	nodeCfg *node.Config
	cmdExec *exec.Cmd
	enodes  []string
}

func NewLocalNode(nodeCfg *node.Config, enodes []string) *Node {
	return &Node{
		nodeCfg: nodeCfg,
		enodes:  enodes,
	}
}

// cleanup deletes any previous process that may be running
func (n *Node) cleanup() error {
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to execute ps command: %w", err)
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, n.nodeCfg.GetDataDir()) && strings.Contains(line, "thor --network") {
			// kill the process
			parts := strings.Fields(line)
			if len(parts) > 1 {
				pid := parts[1]
				slog.Info("killing previous process", "pid", pid, "id", n.nodeCfg.GetID())
				cmd := exec.Command("kill", pid)
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("failed to kill process %s: %w", pid, err)
				}
			}
		}
	}

	return nil
}

func (n *Node) Start() error {
	// cleanup any previous process
	if err := n.cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup previous process - %w", err)
	}
	// ensure directories exist
	if err := os.MkdirAll(n.nodeCfg.GetConfigDir(), 0777); err != nil {
		return fmt.Errorf("unable to create configDir - %w", err)
	}
	if err := os.MkdirAll(n.nodeCfg.GetDataDir(), 0777); err != nil {
		return fmt.Errorf("unable to create configDir - %w", err)
	}

	// write keys to disk
	if n.nodeCfg.GetKey() != "" {
		err := os.WriteFile(filepath.Join(n.nodeCfg.GetConfigDir(), "master.key"), []byte(n.nodeCfg.GetKey()), 0644)
		if err != nil {
			return fmt.Errorf("failed to write master key file - %w", err)
		}
		err = os.WriteFile(filepath.Join(n.nodeCfg.GetConfigDir(), "p2p.key"), []byte(n.nodeCfg.GetKey()), 0644)
		if err != nil {
			return fmt.Errorf("failed to p2p master key file - %w", err)
		}
	}

	// write genesis to disk
	genesisPath := filepath.Join(n.nodeCfg.GetConfigDir(), "genesis.json")
	genesisBytes, err := nodegenesis.Marshal(n.nodeCfg.GetGenesis())
	if err != nil {
		return fmt.Errorf("unable to marshal genesis - %w", err)
	}
	err = os.WriteFile(genesisPath, genesisBytes, 0777)
	if err != nil {
		return fmt.Errorf("failed to write genesis file - %w", err)
	}

	cleanEnode := []string{} // todo theres a clever way of doing this
	for _, enode := range n.enodes {
		nodeEnode, err := n.nodeCfg.Enode("127.0.0.1")
		if err != nil {
			return err
		}
		if nodeEnode != enode {
			cleanEnode = append(cleanEnode, enode)
		}
	}
	enodeString := strings.Join(cleanEnode, ",")

	if err := os.RemoveAll(n.nodeCfg.GetDataDir()); err != nil {
		return fmt.Errorf("failed to remove data dir - %w", err)
	}

	args := []string{
		"thor",
		"--network", genesisPath,
		"--data-dir", n.nodeCfg.GetDataDir(),
		"--config-dir", n.nodeCfg.GetConfigDir(),
		"--api-addr", n.nodeCfg.GetAPIAddr(),
		"--api-cors", n.nodeCfg.GetAPICORS(),
		"--verbosity", strconv.Itoa(n.nodeCfg.GetVerbosity()),
		"--nat", "none",
		"--p2p-port", fmt.Sprintf("%d", n.nodeCfg.GetP2PListenPort()),
		"--bootnode", enodeString,
	}

	for key, value := range n.nodeCfg.GetAdditionalArgs() {
		args = append(args, fmt.Sprintf("--%s", key))
		args = append(args, value)
	}

	cmd := &exec.Cmd{
		Path: n.nodeCfg.GetExecArtifact(),
		Args: args,
		Stdout: &nodeWriter{
			id: n.nodeCfg.GetID(),
			w:  os.Stdout,
		},
		Stderr: &nodeWriter{
			id: n.nodeCfg.GetID(),
			w:  os.Stderr,
		},
	}

	if n.nodeCfg.GetVerbosity() != 0 {
		cmd.Args = append(cmd.Args, "--verbosity", strconv.Itoa(n.nodeCfg.GetVerbosity()))
	}

	slog.Info(cmd.String())
	if n.nodeCfg.GetFakeExecution() {
		slog.Info("FakeExecution enabled - Not starting node: ", "id", n.nodeCfg.GetID())
		return nil
	}
	// Start the command and check for errors
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start thor command: %w", err)
	}

	n.cmdExec = cmd

	slog.Info("started node", "id", n.nodeCfg.GetID(), "pid", n.cmdExec.Process.Pid)
	return nil
}

func (n *Node) Stop() error {
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
			slog.Warn("failed to kill node", "id", n.nodeCfg.GetID(), "pid", n.cmdExec.Process.Pid)
		} else {
			slog.Warn("process killed as timeout reached", "id", n.nodeCfg.GetID(), "pid", n.cmdExec.Process.Pid)
		}
	case err := <-done:
		if err != nil {
			return fmt.Errorf("process exited with error - %w", err)
		}
		slog.Info("node stopped gracefully", "id", n.nodeCfg.GetID(), "pid", n.cmdExec.Process.Pid)
	}
	return nil
}

type nodeWriter struct {
	id string
	w  io.Writer
}

func (nw *nodeWriter) Write(p []byte) (int, error) {
	lines := strings.Split(string(p), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		line = fmt.Sprintf("[%s] %s\n", nw.id, line)
		if _, err := nw.w.Write([]byte(line)); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}
