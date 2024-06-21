package cluster

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/intility/icpctl/internal/redact"
	"github.com/intility/icpctl/internal/telemetry"
	"github.com/intility/icpctl/internal/ux"
	"github.com/intility/icpctl/pkg/client"
	"github.com/intility/icpctl/pkg/clientset"
)

var errInvalidOutputFormat = errors.New(`must be one of "wide", "json", "yaml"`)

func NewListCommand(set clientset.ClientSet) *cobra.Command {
	output := outputFormat("")
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

func printClusterList(writer io.Writer, format outputFormat, clusters client.ClusterList) error {
	var err error

	switch format {
	case "wide":
		table := ux.TableFromObjects(clusters, func(cluster client.Cluster) []ux.Row {
			return []ux.Row{
				ux.NewRow("Name", cluster.Name),
				ux.NewRow("Console URL", cluster.ConsoleURL),
				ux.NewRow("Status", statusString(cluster)),
				ux.NewRow("Status Details", statusMessage(cluster)),
			}
		})

		ux.Fprint(writer, table.String())

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
				ux.NewRow("Console URL", cluster.ConsoleURL),
			}
		})

		ux.Fprint(writer, table.String())

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

type OutputFormat interface {
	String() string
	Set(val string) error
	Type() string
}

type outputFormat string

func (o *outputFormat) String() string {
	return string(*o)
}

func (o *outputFormat) Set(value string) error {
	switch value {
	case "wide", "json", "yaml":
		*o = outputFormat(value)
		return nil
	default:
		return errInvalidOutputFormat
	}
}

func (o *outputFormat) Type() string {
	return "outputFormat"
}
