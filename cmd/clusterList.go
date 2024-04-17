package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/intility/minctl/pkg/client"
)

// clusterListCmd represents the list command.
var clusterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all clusters",
	Long:  `List all clusters that are currently running in kind on the local machine.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		c := client.New(client.WithDevConfig())

		clusters, err := c.ListClusters(cmd.Context())
		if err != nil {
			return fmt.Errorf("could not list clusters: %w", err)
		}

		for _, cluster := range clusters {
			cmd.Println(cluster.Name)
		}

		return nil
	},
}

func init() {
	clusterCmd.AddCommand(clusterListCmd)
}
