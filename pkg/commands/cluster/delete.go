package cluster

import (
	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/clientset"
)

func NewDeleteCommand(set clientset.ClientSet) *cobra.Command {
	var (
		clusterName  string
		errEmptyName = redact.Errorf("cluster name cannot be empty")
	)

	cmd := &cobra.Command{
		Use:     "delete [name]",
		Short:   "Delete a cluster",
		Long:    ``,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "cluster.delete")
			defer span.End()

			// If positional argument is provided, use it (takes precedence)
			if len(args) > 0 {
				clusterName = args[0]
			}

			if clusterName == "" {
				return errEmptyName
			}

			// Get cluster by name
			cluster, err := set.PlatformClient.GetCluster(ctx, clusterName)
			if err != nil {
				return redact.Errorf("could not get cluster: %w", redact.Safe(err))
			}

			if cluster == nil {
				return redact.Errorf("cluster not found: %s", clusterName)
			}

			err = set.PlatformClient.DeleteCluster(ctx, cluster.ID)
			if err != nil {
				return redact.Errorf("could not delete cluster: %w", redact.Safe(err))
			}

			ux.Fprint(cmd.OutOrStdout(), "%s\n", clusterName)

			return nil
		},
	}

	cmd.Flags().StringVarP(&clusterName, "name", "n", "", "Name of the cluster to delete")

	return cmd
}
