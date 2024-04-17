package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/intility/minctl/pkg/cli"
)

// clusterCmd represents the cluster command.
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage cluster resources",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cmd.Help()
		if err != nil {
			return fmt.Errorf("could not run help command: %w", err)
		}

		return nil
	},
}

func init() {
	clusterCmd.PersistentPreRunE = cli.CreateAuthGate(
		"please log in with 'minctl login' before managing cluster resources")

	rootCmd.AddCommand(clusterCmd)
}
