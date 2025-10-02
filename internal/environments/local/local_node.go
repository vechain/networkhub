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

	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	nodegenesis "github.com/vechain/networkhub/network/node/genesis"
)

type Node struct {
	nodeCfg    node.Config
	networkCfg *network.Network
	cmdExec    *exec.Cmd
	enodes     []string
}

func NewLocalNode(nodeCfg node.Config, networkCfg *network.Network, enodes []string) *Node {
	return &Node{
		nodeCfg:    nodeCfg,
		networkCfg: networkCfg,
		enodes:     enodes,
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
	if err := n.prepareNode(); err != nil {
		return fmt.Errorf("failed to prepare node: %w", err)
	}

	args, err := n.buildCommandArgs()
	if err != nil {
		return fmt.Errorf("failed to build command args: %w", err)
	}

	cmd, err := n.createCommand(args)
	if err != nil {
		return fmt.Errorf("failed to create command: %w", err)
	}

	return n.executeCommand(cmd)
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
	lines := strings.SplitSeq(string(p), "\n")
	for line := range lines {
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

// isPublicNetwork checks if this node is configured to connect to a public network (testnet/mainnet)
func (n *Node) isPublicNetwork() bool {
	return n.networkCfg.IsPublicNetwork()
}

// getPublicNetworkName returns the public network name from network configuration
func (n *Node) getPublicNetworkName() string {
	return n.networkCfg.GetThorNetworkArg()
}

// prepareNode prepares the node for startup by cleaning up, creating directories, and writing config files
func (n *Node) prepareNode() error {
	// Cleanup any previous process
	if err := n.cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup previous process: %w", err)
	}

	// Ensure directories exist
	if err := n.createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Write configuration files
	if err := n.writeConfigFiles(); err != nil {
		return fmt.Errorf("failed to write config files: %w", err)
	}

	return nil
}

// createDirectories creates the necessary directories for the node
func (n *Node) createDirectories() error {
	if err := os.MkdirAll(n.nodeCfg.GetConfigDir(), 0777); err != nil {
		return fmt.Errorf("unable to create configDir: %w", err)
	}
	if err := os.MkdirAll(n.nodeCfg.GetDataDir(), 0777); err != nil {
		return fmt.Errorf("unable to create dataDir: %w", err)
	}
	return nil
}

// writeConfigFiles writes keys and genesis files as needed
func (n *Node) writeConfigFiles() error {
	isPublicNetwork := n.isPublicNetwork()

	// Write keys to disk (only for local networks or master nodes)
	if err := n.writeKeys(isPublicNetwork); err != nil {
		return fmt.Errorf("failed to write keys: %w", err)
	}

	// Write genesis to disk (only for local networks)
	if err := n.writeGenesis(isPublicNetwork); err != nil {
		return fmt.Errorf("failed to write genesis: %w", err)
	}

	// Clean data directory (only for local networks)
	if err := n.cleanDataDirectory(isPublicNetwork); err != nil {
		return fmt.Errorf("failed to clean data directory: %w", err)
	}

	return nil
}

// writeKeys writes master and p2p keys to disk if needed
func (n *Node) writeKeys(isPublicNetwork bool) error {
	if n.nodeCfg.GetKey() == "" || isPublicNetwork {
		return nil // No keys needed for public networks or nodes without keys
	}

	keyData := []byte(n.nodeCfg.GetKey())

	// Write master key
	masterKeyPath := filepath.Join(n.nodeCfg.GetConfigDir(), "master.key")
	if err := os.WriteFile(masterKeyPath, keyData, 0644); err != nil {
		return fmt.Errorf("failed to write master key file: %w", err)
	}

	// Write p2p key
	p2pKeyPath := filepath.Join(n.nodeCfg.GetConfigDir(), "p2p.key")
	if err := os.WriteFile(p2pKeyPath, keyData, 0644); err != nil {
		return fmt.Errorf("failed to write p2p key file: %w", err)
	}

	return nil
}

// writeGenesis writes the genesis file to disk if needed
func (n *Node) writeGenesis(isPublicNetwork bool) error {
	if isPublicNetwork {
		return nil // Public networks don't need genesis files
	}

	genesisPath := filepath.Join(n.nodeCfg.GetConfigDir(), "genesis.json")
	genesisBytes, err := nodegenesis.Marshal(n.nodeCfg.GetGenesis())
	if err != nil {
		return fmt.Errorf("unable to marshal genesis: %w", err)
	}

	if err := os.WriteFile(genesisPath, genesisBytes, 0777); err != nil {
		return fmt.Errorf("failed to write genesis file: %w", err)
	}

	return nil
}

// cleanDataDirectory removes the data directory for local networks
func (n *Node) cleanDataDirectory(isPublicNetwork bool) error {
	if isPublicNetwork {
		return nil // Public networks should sync from scratch, don't clean
	}

	if err := os.RemoveAll(n.nodeCfg.GetDataDir()); err != nil {
		return fmt.Errorf("failed to remove data dir: %w", err)
	}

	return nil
}

// buildCommandArgs builds the command line arguments for the thor process
func (n *Node) buildCommandArgs() ([]string, error) {
	args := []string{"thor"}

	// Add network parameter
	args = n.addNetworkArg(args)

	// Add common arguments
	args = n.addCommonArgs(args)

	// Add bootnodes for local networks
	args, err := n.addBootnodes(args)
	if err != nil {
		return nil, fmt.Errorf("failed to add bootnodes: %w", err)
	}

	// Add additional arguments
	args = n.addAdditionalArgs(args)

	return args, nil
}

// addNetworkArg adds the network parameter to the command args
func (n *Node) addNetworkArg(args []string) []string {
	isPublicNetwork := n.isPublicNetwork()

	if isPublicNetwork {
		// For public networks, use the network name directly
		networkName := n.getPublicNetworkName()
		args = append(args, "--network", networkName)
	} else {
		// For local networks, use genesis file
		genesisPath := filepath.Join(n.nodeCfg.GetConfigDir(), "genesis.json")
		args = append(args, "--network", genesisPath)
	}

	return args
}

// addCommonArgs adds common command line arguments
func (n *Node) addCommonArgs(args []string) []string {
	args = append(args,
		"--data-dir", n.nodeCfg.GetDataDir(),
		"--config-dir", n.nodeCfg.GetConfigDir(),
		"--api-addr", n.nodeCfg.GetAPIAddr(),
		"--api-cors", n.nodeCfg.GetAPICORS(),
		"--verbosity", strconv.Itoa(n.nodeCfg.GetVerbosity()),
		"--nat", "none",
		"--p2p-port", fmt.Sprintf("%d", n.nodeCfg.GetP2PListenPort()),
	)
	return args
}

// addBootnodes adds bootnode arguments for local networks
func (n *Node) addBootnodes(args []string) ([]string, error) {
	isPublicNetwork := n.isPublicNetwork()
	if isPublicNetwork {
		return args, nil
	}

	cleanEnodes := n.cleanEnodes()
	if len(cleanEnodes) > 0 {
		enodeString := strings.Join(cleanEnodes, ",")
		args = append(args, "--bootnode", enodeString)
	}

	return args, nil
}

// cleanEnodes filters out the current node's enode from the list
func (n *Node) cleanEnodes() []string {
	var cleanEnodes []string
	for _, enode := range n.enodes {
		nodeEnode, err := n.nodeCfg.Enode("127.0.0.1")
		if err != nil {
			continue // Skip invalid enodes
		}
		if nodeEnode != enode {
			cleanEnodes = append(cleanEnodes, enode)
		}
	}
	return cleanEnodes
}

// addAdditionalArgs adds additional command line arguments
func (n *Node) addAdditionalArgs(args []string) []string {
	for key, value := range n.nodeCfg.GetAdditionalArgs() {
		args = append(args, fmt.Sprintf("--%s", key))
		args = append(args, value)
	}
	return args
}

// createCommand creates the exec.Cmd with the given arguments
func (n *Node) createCommand(args []string) (*exec.Cmd, error) {
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

	return cmd, nil
}

// executeCommand executes the command and handles fake execution
func (n *Node) executeCommand(cmd *exec.Cmd) error {
	slog.Info(cmd.String())

	if n.nodeCfg.GetFakeExecution() {
		slog.Info("FakeExecution enabled - Not starting node: ", "id", n.nodeCfg.GetID())
		slog.Info("Waiting 10 seconds for node to start...")
		time.Sleep(10 * time.Second)
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
