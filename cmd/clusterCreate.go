package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/intility/minctl/pkg/client"
)

// clusterCreateCmd represents the create command.
var clusterCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new cluster",
	Long:  `Create a new cluster with the specified configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(client.WithDevConfig())

		req := client.NewClusterRequest{Name: args[0]}
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
	clusterCmd.AddCommand(clusterCreateCmd)
}
