package cmd

import (
	"fmt"
	"github.com/vechain/networkhub/environments/docker"

	"github.com/vechain/networkhub/hub"
	"github.com/vechain/networkhub/preset"

	"github.com/vechain/networkhub/entrypoint/api"
	"github.com/vechain/networkhub/environments/local"

	"github.com/spf13/cobra"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Starts NetworkHub as an HTTP API server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("api called")

		envManager := hub.NewNetworkHub()
		envManager.RegisterEnvironment("local", local.NewLocalEnv)
		envManager.RegisterEnvironment("docker", docker.NewDockerEnv)

		presets := preset.NewPresetNetworks()
		presets.Register("threeMasterNodesNetwork", preset.LocalThreeMasterNodesNetwork)
		presets.Register("sixNodesNetwork", preset.LocalSixNodesNetwork)

		httpAPI := api.New(envManager, presets)

		if err := httpAPI.Start(); err != nil {
			fmt.Println("Shutting down.. Unexpected error in api - %w", err)
			return
		}
		fmt.Println("Shutting down..")
	},
}

func init() {
	rootCmd.AddCommand(apiCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// localCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// localCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
