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
		Use:     "status [name]",
		Short:   "Show detailed status of a cluster",
		Long:    `Display comprehensive status information for a cluster.`,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "cluster.status")
			defer span.End()

			cmd.SilenceUsage = true

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

			printClusterStatus(cmd.OutOrStdout(), cluster)

			return nil
		},
	}

	cmd.Flags().StringVarP(&clusterName, "name", "n", "", "Name of the cluster")

	return cmd
}

func printClusterStatus(writer io.Writer, cluster *client.Cluster) {
	// Determine and display ONLY the status
	if cluster.Status.Ready.Status {
		ux.Fprintf(writer, "Ready\n")
	} else if cluster.Status.Deployment.Active {
		ux.Fprintf(writer, "In Deployment\n")
	} else if cluster.Status.Deployment.Failed {
		ux.Fprintf(writer, "Deployment Failed\n")
	} else {
		ux.Fprintf(writer, "Not Ready\n")
	}
}
