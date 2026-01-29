package cluster

import (
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

func NewGetCommand(set clientset.ClientSet) *cobra.Command {
	var (
		clusterName  string
		errEmptyName = redact.Errorf("cluster name cannot be empty")
	)

	cmd := &cobra.Command{
		Use:     "get [name]",
		Short:   "Get detailed information about a cluster",
		Long:    `Display comprehensive cluster information.`,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "cluster.get")
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

			printClusterDetails(cmd.OutOrStdout(), cluster)

			return nil
		},
	}

	cmd.Flags().StringVarP(&clusterName, "name", "n", "", "Name of the cluster")

	return cmd
}

func printClusterDetails(writer io.Writer, cluster *client.Cluster) {
	// Basic cluster information
	ux.Fprintf(writer, "Cluster Information:\n")
	ux.Fprintf(writer, "  Name:        %s\n", cluster.Name)
	ux.Fprintf(writer, "  ID:          %s\n", cluster.ID)
	ux.Fprintf(writer, "  Version:     %s\n", cluster.Version)
	ux.Fprintf(writer, "  Console URL: %s\n", cluster.ConsoleURL)

	if len(cluster.Roles) > 0 {
		ux.Fprintf(writer, "  Roles:       %s\n", strings.Join(cluster.Roles, ", "))
	}

	printStatusInfo(writer, cluster)

	// Node pools
	if len(cluster.NodePools) > 0 {
		ux.Fprintf(writer, "\nNode Pools:\n")

		for i, pool := range cluster.NodePools {
			if i > 0 {
				ux.Fprintf(writer, "\n")
			}

			printNodePool(writer, pool, i+1)
		}
	}
}

// printStatusInfo prints the cluster status information.
func printStatusInfo(writer io.Writer, cluster *client.Cluster) {
	ux.Fprintf(writer, "  Status:      ")

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

	if cluster.Status.Ready.Message != "" {
		ux.Fprintf(writer, "  Message:     %s\n", cluster.Status.Ready.Message)
	}

	if cluster.Status.Ready.Reason != "" {
		ux.Fprintf(writer, "  Reason:      %s\n", cluster.Status.Ready.Reason)
	}
}

// printNodePool prints detailed information about a single node pool.
func printNodePool(writer io.Writer, pool client.NodePool, index int) {
	ux.Fprintf(writer, "  Pool %d:\n", index)
	ux.Fprintf(writer, "    Name:               %s\n", pool.Name)

	if pool.ID != "" {
		ux.Fprintf(writer, "    ID:                 %s\n", pool.ID)
	}

	ux.Fprintf(writer, "    Preset:             %s\n", pool.Preset)

	if pool.Replicas != nil {
		ux.Fprintf(writer, "    Replicas:           %d\n", *pool.Replicas)
	}

	if pool.Compute != nil {
		ux.Fprintf(writer, "    Compute:\n")
		ux.Fprintf(writer, "      Cores:            %d\n", pool.Compute.Cores)
		ux.Fprintf(writer, "      Memory:           %s\n", pool.Compute.Memory)
	}

	ux.Fprintf(writer, "    Autoscaling:        %v\n", pool.AutoscalingEnabled)

	if pool.AutoscalingEnabled {
		if pool.MinCount != nil {
			ux.Fprintf(writer, "      Min Count:        %d\n", *pool.MinCount)
		}

		if pool.MaxCount != nil {
			ux.Fprintf(writer, "      Max Count:        %d\n", *pool.MaxCount)
		}
	}
}
