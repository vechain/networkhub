package local

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vechain/networkhub/network/node"
)

type Node struct {
	nodeCfg node.Node
	cmdExec *exec.Cmd
	enodes  []string
}

func NewLocalNode(nodeCfg node.Node, enodes []string) *Node {
	return &Node{
		nodeCfg: nodeCfg,
		enodes:  enodes,
	}
}

func (n *Node) Start() error {
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
	genesisBytes, err := json.Marshal(n.nodeCfg.GetGenesis())
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

	cmd := &exec.Cmd{
		Path: n.nodeCfg.GetExecArtifact(),
		Args: []string{
			"thor",
			"--network", genesisPath,
			"--data-dir", n.nodeCfg.GetDataDir(),
			"--config-dir", n.nodeCfg.GetConfigDir(),
			"--api-addr", n.nodeCfg.GetAPIAddr(),
			"--api-cors", n.nodeCfg.GetAPICORS(),
			"--verbosity", "4",
			"--nat", "none",
			"--p2p-port", fmt.Sprintf("%d", n.nodeCfg.GetP2PListenPort()),
			"--bootnode", enodeString,
		},
		Stdout: os.Stdout, // Directing stdout to the same stdout of the Go program
		Stderr: os.Stderr, // Directing stderr to the same stderr of the Go program
	}

	if n.nodeCfg.GetVerbosity() != 0 {
		cmd.Args = append(cmd.Args, "--verbosity", strconv.Itoa(n.nodeCfg.GetVerbosity()))
	}

	fmt.Println(cmd)
	if n.nodeCfg.GetFakeExecution() {
		fmt.Println("FakeExecution enabled - Not starting node: ", n.nodeCfg.GetID())
		return nil
	}
	// Start the command and check for errors
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start thor command: %w", err)
	}

	n.cmdExec = cmd

	fmt.Println("Thor command executed successfully")
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
