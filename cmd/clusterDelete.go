package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/intility/minctl/pkg/client"
)

// clusterDeleteCmd represents the clusterDelete command.
var clusterDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a cluster",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(client.WithDevConfig())

		if clusterName == "" {
			return errEmptyName
		}

		err := c.DeleteCluster(cmd.Context(), clusterName)
		if err != nil {
			return fmt.Errorf("could not delete cluster: %w", err)
		}

		cmd.Println(clusterName)

		return nil
	},
}

func init() {
	clusterDeleteCmd.Flags().StringVarP(&clusterName, "name", "n", "", "Name of the cluster to delete")
	clusterCmd.AddCommand(clusterDeleteCmd)
}
