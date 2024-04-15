package cmd

import (
	"fmt"
	"os/exec"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
)

// clusterCreateCmd represents the create command.
var clusterCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cluster",
	Long:  `Create a new cluster with the specified configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		command := exec.Command("init-kind.sh")
		spin := spinner.New(spinner.CharSets[11], spinnerDelay)

		spin.Start()

		out, err := command.Output()
		if err != nil {
			return fmt.Errorf("could not run command: %w", err)
		}

		spin.Stop()

		fmt.Println(string(out))

		return nil
	},
}

func init() {
	clusterCmd.AddCommand(clusterCreateCmd)
	clusterCreateCmd.Flags().StringVarP(&Name, "name", "n", "", "Name of the cluster to create")
}
