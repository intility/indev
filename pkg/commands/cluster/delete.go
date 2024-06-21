package cluster

import (
	"github.com/intility/icpctl/internal/redact"
	"github.com/intility/icpctl/internal/telemetry"
	"github.com/intility/icpctl/internal/ux"
	"github.com/intility/icpctl/pkg/clientset"
	"github.com/spf13/cobra"
)

func NewDeleteCommand(set clientset.ClientSet) *cobra.Command {

	var (
		clusterName  string
		errEmptyName = redact.Errorf("cluster name cannot be empty")
	)

	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete a cluster",
		Long:    ``,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "cluster.delete")
			defer span.End()

			if clusterName == "" {
				return errEmptyName
			}

			err := set.PlatformClient.DeleteCluster(ctx, clusterName)
			if err != nil {
				return redact.Errorf("could not delete cluster: %w", redact.Safe(err))
			}

			ux.Fprint(cmd.OutOrStdout(), clusterName+"\n")

			return nil
		},
	}

	cmd.Flags().StringVarP(&clusterName, "name", "n", "", "Name of the cluster to delete")

	return cmd
}
