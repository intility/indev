package cluster

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/client"
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

			// List clusters to find the one by name
			clusters, err := set.PlatformClient.ListClusters(ctx)
			if err != nil {
				return redact.Errorf("could not list clusters: %w", redact.Safe(err))
			}

			// Find the cluster with the matching name
			var cluster *client.Cluster
			for _, c := range clusters {
				if strings.EqualFold(c.Name, clusterName) {
					cluster = &c
					break
				}
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
