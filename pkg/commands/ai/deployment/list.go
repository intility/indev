package deployment

import (
	"encoding/json"
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
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List AI deployments",
		Long:    "List all AI deployments in the Intility Developer Platform",
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "aideployment.list")
			defer span.End()

			cmd.SilenceUsage = true

			deployments, err := set.PlatformClient.ListAIDeployments(ctx)
			if err != nil {
				return redact.Errorf("could not list AI deployments: %w", redact.Safe(err))
			}

			if len(deployments) == 0 {
				ux.Fprintf(cmd.OutOrStdout(), "No AI deployments found\n")
				return nil
			}

			if err = printDeploymentList(cmd.OutOrStdout(), output, deployments); err != nil {
				return redact.Errorf("could not print AI deployments: %w", redact.Safe(err))
			}

			return nil
		},
	}

	cmd.Flags().VarP(&output, "output", "o", "Output format (wide, json, yaml)")

	return cmd
}

func printDeploymentList(writer io.Writer, format outputformat.Format, deployments []client.AIDeployment) error {
	var err error

	switch format {
	case "json":
		enc := json.NewEncoder(writer)
		enc.SetIndent("", "  ")
		err = enc.Encode(deployments)
	case "yaml":
		indent := 2
		enc := yaml.NewEncoder(writer)
		enc.SetIndent(indent)
		err = enc.Encode(deployments)
	default:
		table := ux.TableFromObjects(deployments, func(d client.AIDeployment) []ux.Row {
			return []ux.Row{
				ux.NewRow("Name", d.Name),
				ux.NewRow("Model", d.Model),
				ux.NewRow("Endpoint", d.Endpoint),
				ux.NewRow("Created By", d.CreatedBy.Name),
			}
		})

		ux.Fprintf(writer, "%s", table.String())
	}

	if err != nil {
		return redact.Errorf("output encoder failed: %w", redact.Safe(err))
	}

	return nil
}
