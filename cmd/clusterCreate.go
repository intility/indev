package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/intility/minctl/pkg/client"
)

var (
	clusterName string
)

// clusterCreateCmd represents the create command.
var clusterCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cluster",
	Long:  `Create a new cluster with the specified configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(client.WithDevConfig())

		if clusterName == "" {
			return fmt.Errorf("cluster name is required")
		}

		req := client.NewClusterRequest{Name: clusterName}
		cluster, err := c.CreateCluster(cmd.Context(), req)

		// we are done with validating the input
		cmd.SilenceUsage = true

		if err != nil {
			return fmt.Errorf("could not create cluster: %w", err)
		}

		cmd.Println(cluster.Name)

		return nil
	},
}

func init() {
	clusterCreateCmd.Flags().StringVarP(&clusterName, "name", "n", "", "Name of the cluster to create")

	clusterCmd.AddCommand(clusterCreateCmd)
}
