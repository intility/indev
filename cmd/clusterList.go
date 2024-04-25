package cmd

import (
	"github.com/spf13/cobra"

	"github.com/intility/minctl/internal/redact"
	"github.com/intility/minctl/internal/ux"
	"github.com/intility/minctl/pkg/client"
)

// clusterListCmd represents the list command.
var clusterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all clusters",
	Long:  `List all clusters that are currently running in kind on the local machine.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		c := client.New()

		clusters, err := c.ListClusters(cmd.Context())
		if err != nil {
			return redact.Errorf("could not list clusters: %w", redact.Safe(err))
		}

		for _, cluster := range clusters {
			ux.Fprint(cmd.OutOrStdout(), cluster.Name+"\n")
		}

		return nil
	},
}

func init() {
	clusterCmd.AddCommand(clusterListCmd)
}
