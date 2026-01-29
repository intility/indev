package cluster

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

func NewStatusCommand(set clientset.ClientSet) *cobra.Command {
	var clusterName string

	cmd := &cobra.Command{
		Use:     "status [name]",
		Short:   "Show detailed status of a cluster",
		Long:    `Display comprehensive status information for a cluster.`,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "cluster.status")
			defer span.End()

			return runClusterLookupCommand(ctx, lookupParams{
				cmd:         cmd,
				set:         set,
				args:        args,
				clusterName: clusterName,
				printer:     printClusterStatus,
			})
		},
	}

	cmd.Flags().StringVarP(&clusterName, "name", "n", "", "Name of the cluster")

	return cmd
}

func printClusterStatus(writer io.Writer, cluster *client.Cluster) {
	// Determine and display ONLY the status
	switch {
	case cluster.Status.Ready.Status:
		ux.Fprintf(writer, "Ready\n")
	case cluster.Status.Deployment.Active:
		ux.Fprintf(writer, "In Deployment\n")
	case cluster.Status.Deployment.Failed:
		ux.Fprintf(writer, "Deployment Failed\n")
	default:
		ux.Fprintf(writer, "Not Ready\n")
	}
}
