package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/intility/minctl/pkg/client"
)

// clusterDeleteCmd represents the clusterDelete command.
var clusterDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a cluster",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(client.WithDevConfig())

		err := c.DeleteCluster(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("could not delete cluster: %w", err)
		}

		cmd.Println(args[0])

		return nil
	},
}

func init() {
	clusterCmd.AddCommand(clusterDeleteCmd)
}
