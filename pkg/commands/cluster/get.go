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
		Long:    `Display comprehensive information about a specific cluster, including configuration, node pools, and status.`,
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

			if err = printClusterDetails(cmd.OutOrStdout(), cluster); err != nil {
				return redact.Errorf("could not print cluster details: %w", redact.Safe(err))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&clusterName, "name", "n", "", "Name of the cluster")

	return cmd
}

func printClusterDetails(writer io.Writer, cluster *client.Cluster) error {
	// Basic cluster information
	ux.Fprint(writer, "Cluster Information:\n")
	ux.Fprint(writer, "  Name:        %s\n", cluster.Name)
	ux.Fprint(writer, "  ID:          %s\n", cluster.ID)
	ux.Fprint(writer, "  Version:     %s\n", cluster.Version)
	ux.Fprint(writer, "  Console URL: %s\n", cluster.ConsoleURL)
	if len(cluster.Roles) > 0 {
		ux.Fprint(writer, "  Roles:       %s\n", strings.Join(cluster.Roles, ", "))
	}

	// Status - use same logic as status command
	ux.Fprint(writer, "  Status:      ")
	if cluster.Status.Ready.Status {
		ux.Fprint(writer, "Ready\n")
	} else if cluster.Status.Deployment.Active {
		ux.Fprint(writer, "In Deployment\n")
	} else if cluster.Status.Deployment.Failed {
		ux.Fprint(writer, "Deployment Failed\n")
	} else {
		ux.Fprint(writer, "Not Ready\n")
	}

	// Show message and reason if available
	if cluster.Status.Ready.Message != "" {
		ux.Fprint(writer, "  Message:     %s\n", cluster.Status.Ready.Message)
	}
	if cluster.Status.Ready.Reason != "" {
		ux.Fprint(writer, "  Reason:      %s\n", cluster.Status.Ready.Reason)
	}

	// Node pools
	if len(cluster.NodePools) > 0 {
		ux.Fprint(writer, "\nNode Pools:\n")
		for i, pool := range cluster.NodePools {
			if i > 0 {
				ux.Fprint(writer, "\n")
			}
			ux.Fprint(writer, "  Pool %d:\n", i+1)
			ux.Fprint(writer, "    Name:               %s\n", pool.Name)
			if pool.ID != "" {
				ux.Fprint(writer, "    ID:                 %s\n", pool.ID)
			}
			ux.Fprint(writer, "    Preset:             %s\n", pool.Preset)

			if pool.Replicas != nil {
				ux.Fprint(writer, "    Replicas:           %d\n", *pool.Replicas)
			}

			if pool.Compute != nil {
				ux.Fprint(writer, "    Compute:\n")
				ux.Fprint(writer, "      Cores:            %d\n", pool.Compute.Cores)
				ux.Fprint(writer, "      Memory:           %s\n", pool.Compute.Memory)
			}

			ux.Fprint(writer, "    Autoscaling:        %v\n", pool.AutoscalingEnabled)
			if pool.AutoscalingEnabled {
				if pool.MinCount != nil {
					ux.Fprint(writer, "      Min Count:        %d\n", *pool.MinCount)
				}
				if pool.MaxCount != nil {
					ux.Fprint(writer, "      Max Count:        %d\n", *pool.MaxCount)
				}
			}
		}
	}

	return nil
}
