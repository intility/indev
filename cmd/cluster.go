package cmd

import (
	"github.com/spf13/cobra"

	"github.com/intility/minctl/internal/cli"
	"github.com/intility/minctl/internal/redact"
)

var errEmptyName = redact.Errorf("%s", redact.Safe("cluster name cannot be empty"))

// clusterCmd represents the cluster command.
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage cluster resources",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cmd.Help()
		if err != nil {
			return redact.Errorf("could not run help command: %w", redact.Safe(err))
		}

		return nil
	},
}

func init() {
	clusterCmd.PersistentPreRunE = cli.CreateAuthGate(
		redact.Safe("please log in with 'minctl login' before managing cluster resources"))

	rootCmd.AddCommand(clusterCmd)
}
