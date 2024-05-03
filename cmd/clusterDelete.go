package cmd

import (
	"github.com/spf13/cobra"

	"github.com/intility/icpctl/internal/redact"
	"github.com/intility/icpctl/internal/ux"
	"github.com/intility/icpctl/pkg/client"
)

// clusterDeleteCmd represents the clusterDelete command.
var clusterDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a cluster",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New()

		if clusterName == "" {
			return errEmptyName
		}

		err := c.DeleteCluster(cmd.Context(), clusterName)
		if err != nil {
			return redact.Errorf("could not delete cluster: %w", redact.Safe(err))
		}

		ux.Fprint(cmd.OutOrStdout(), clusterName+"\n")

		return nil
	},
}

func init() {
	clusterDeleteCmd.Flags().StringVarP(&clusterName, "name", "n", "", "Name of the cluster to delete")
	clusterCmd.AddCommand(clusterDeleteCmd)
}
