package cluster

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

func NewStatusCommand(set clientset.ClientSet) *cobra.Command {
	var (
		clusterName  string
		errEmptyName = redact.Errorf("cluster name cannot be empty")
	)

	cmd := &cobra.Command{
		Use:     "status",
		Short:   "Show detailed status of a cluster",
		Long:    `Display comprehensive status information for a specific cluster, including node pools, deployment status, and resources.`,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "cluster.status")
			defer span.End()

			cmd.SilenceUsage = true

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
				if c.Name == clusterName {
					cluster = &c
					break
				}
			}

			if cluster == nil {
				return redact.Errorf("cluster not found: %s", clusterName)
			}

			if err = printClusterStatus(cmd.OutOrStdout(), cluster); err != nil {
				return redact.Errorf("could not print cluster status: %w", redact.Safe(err))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&clusterName, "name", "n", "", "Name of the cluster")

	return cmd
}

func printClusterStatus(writer io.Writer, cluster *client.Cluster) error {
	// Determine and display ONLY the status
	if cluster.Status.Ready.Status {
		ux.Fprint(writer, "Ready\n")
	} else if cluster.Status.Deployment.Active {
		ux.Fprint(writer, "In Deployment\n")
	} else if cluster.Status.Deployment.Failed {
		ux.Fprint(writer, "Deployment Failed\n")
	} else {
		ux.Fprint(writer, "Not Ready\n")
	}

	return nil
}
