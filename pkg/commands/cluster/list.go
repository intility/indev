package cluster

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
	"github.com/intility/indev/pkg/outputformat"
)

func NewListCommand(set clientset.ClientSet) *cobra.Command {
	output := outputformat.Format("")
	// clusterListCmd represents the list command.
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all clusters",
		Long:    `List all clusters that are currently running in kind on the local machine.`,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "cluster.list")
			defer span.End()

			cmd.SilenceUsage = true

			clusters, err := set.PlatformClient.ListClusters(ctx)
			if err != nil {
				return redact.Errorf("could not list clusters: %w", redact.Safe(err))
			}

			if len(clusters) == 0 {
				ux.Fprint(cmd.OutOrStdout(), "No clusters found\n")
				return nil
			}

			if err = printClusterList(cmd.OutOrStdout(), output, clusters); err != nil {
				return redact.Errorf("could not print cluster list: %w", redact.Safe(err))
			}

			return nil
		},
	}

	cmd.Flags().VarP(&output, "output", "o", "Output format (wide, json, yaml)")

	return cmd
}

func printClusterList(writer io.Writer, format outputformat.Format, clusters client.ClusterList) error {
	var err error

	switch format {
	case "wide":
		table := ux.TableFromObjects(clusters, func(cluster client.Cluster) []ux.Row {
			return []ux.Row{
				ux.NewRow("Name", cluster.Name),
				ux.NewRow("Version", cluster.Version),
				ux.NewRow("Console URL", cluster.ConsoleURL),
				ux.NewRow("Node Pools", nodePoolSummary(cluster)),
				ux.NewRow("Status", statusString(cluster)),
				ux.NewRow("Status Details", statusMessage(cluster)),
				ux.NewRow("Roles", rolesString(cluster)),
			}
		})

		ux.Fprint(writer, "%s", table.String())

		return nil
	case "json":
		enc := json.NewEncoder(writer)
		enc.SetIndent("", "  ")
		err = enc.Encode(clusters)
	case "yaml":
		indent := 2
		enc := yaml.NewEncoder(writer)
		enc.SetIndent(indent)
		err = enc.Encode(clusters)
	default:
		table := ux.TableFromObjects(clusters, func(cluster client.Cluster) []ux.Row {
			return []ux.Row{
				ux.NewRow("Name", cluster.Name),
				ux.NewRow("Version", cluster.Version),
				ux.NewRow("Status", statusString(cluster)),
				ux.NewRow("Node Pools", nodePoolSummary(cluster)),
			}
		})

		ux.Fprint(writer, "%s", table.String())

		return nil
	}

	if err != nil {
		return redact.Errorf("output encoder failed: %w", redact.Safe(err))
	}

	return nil
}

func statusString(cluster client.Cluster) string {
	if cluster.Status.Ready.Status {
		return "Ready"
	}

	if cluster.Status.Deployment.Active {
		return "In Deployment"
	}

	if cluster.Status.Deployment.Failed {
		return "Deployment Failed"
	}

	return "Not Ready"
}

func statusMessage(cluster client.Cluster) string {
	return cluster.Status.Ready.Message
}

func nodePoolSummary(cluster client.Cluster) string {
	if len(cluster.NodePools) == 0 {
		return "0"
	}

	totalNodes := 0

	for _, pool := range cluster.NodePools {
		if pool.Replicas != nil {
			totalNodes += *pool.Replicas
		}
	}

	return fmt.Sprintf("%d pool(s), %d node(s)", len(cluster.NodePools), totalNodes)
}

func rolesString(cluster client.Cluster) string {
	if len(cluster.Roles) == 0 {
		return ""
	}

	return fmt.Sprintf("%v", cluster.Roles)
}
