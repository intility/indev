package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

const (
	spinnerDelay = 100 * time.Millisecond
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
	rootCmd.AddCommand(clusterCmd)
}
