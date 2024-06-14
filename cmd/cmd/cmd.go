package cmd

import (
	"fmt"
	"github.com/vechain/networkhub/environments/docker"
	"io/ioutil"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/vechain/networkhub/environments/local"
	"github.com/vechain/networkhub/hub"
	"github.com/vechain/networkhub/preset"

	"github.com/spf13/cobra"

	cmdentrypoint "github.com/vechain/networkhub/entrypoint/cmd"
)

func setup() *cmdentrypoint.Cmd {
	envManager := hub.NewNetworkHub()
	envManager.RegisterEnvironment("local", local.NewLocalEnv)
	envManager.RegisterEnvironment("docker", docker.NewDockerEnv)

	presets := preset.NewPresetNetworks()
	presets.Register("threeMasterNodesNetwork", preset.LocalThreeMasterNodesNetwork)
	presets.Register("sixNodesNetwork", preset.LocalSixNodesNetwork)

	execDir, err := os.Getwd() // TODO might want to make this configurable in the future ?
	if err != nil {
		panic(fmt.Errorf("unable to use current directory: %w", err))
	}

	cmdEntrypoint := cmdentrypoint.New(envManager, presets, filepath.Join(execDir, "networks_db.json"))
	if err = cmdEntrypoint.LoadExistingNetworks(); err != nil {
		panic(fmt.Errorf("unable to load existing networks: %w", err))
	}

	return cmdEntrypoint
}

var cmdCmd = &cobra.Command{
	Use:   "cmd",
	Short: "Directly uses NetworkHub",
}

var startCmd = &cobra.Command{
	Use:   "start [network-id]",
	Short: "Start a specific network",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cmdManager := setup()
		networkID := args[0]
		slog.Info("Starting network...", "ID", networkID)

		// Channel to listen for interrupt signals
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			err := cmdManager.Start(networkID)
			if err != nil {
				slog.Error("unable to start network", "err", err)
				return
			}
			slog.Info("network started successfully...")
		}()

		// Wait for interrupt signal
		<-sigChan
		slog.Info("Interrupt signal received. Stopping the network...")

		err := cmdManager.Stop(networkID)
		if err != nil {
			slog.Error("unable to stop network", "err", err)
		} else {
			slog.Info("network stopped successfully.")
		}
	},
}

var configureCmd = &cobra.Command{
	Use:   "config [network-json-config]",
	Short: "Configures a specific network",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cmdManager := setup()

		// Read from the specified file
		data, err := ioutil.ReadFile(args[0])
		if err != nil {
			fmt.Printf("Error reading config file: %v\n", err)
			os.Exit(1)
		}

		slog.Info("Configuring network...")

		networkID, err := cmdManager.Config(string(data))
		if err != nil {
			slog.Error("unable to config network", "err", err)
			return
		}
		slog.Info("network config was successful...", "networkId", networkID)
	},
}

// TODO add a preset list
var presetCmd = &cobra.Command{
	Use:   "preset [environment] [preset-name] [preset-thor-path]",
	Short: "Configures a preset network",
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		cmdManager := setup()

		presetEnv := args[0]
		presetNetwork := args[1]
		presetArtifactPath := args[2]

		slog.Info("Configuring network...")
		networkID, err := cmdManager.Preset(presetNetwork, presetEnv, presetArtifactPath)
		if err != nil {
			slog.Error("unable to config preset network", "err", err)
			return
		}
		slog.Info("preset network config was successful...", "networkId", networkID)
	},
}

func init() {
	cmdCmd.AddCommand(startCmd, configureCmd, presetCmd)
	rootCmd.AddCommand(cmdCmd)
}
