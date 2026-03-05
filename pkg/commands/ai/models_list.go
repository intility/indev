package ai

import (
	"encoding/json"
	"io"
	"strconv"

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
		Short:   "List all AI models",
		Long:    `List all available AI models in the Intility Developer Platform`,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "aimodels.list")
			defer span.End()

			cmd.SilenceUsage = true

			models, err := set.PlatformClient.ListAIModels(ctx)
			if err != nil {
				return redact.Errorf("could not list aimodels: %w", redact.Safe(err))
			}

			if len(models) == 0 {
				ux.Fprintf(cmd.OutOrStdout(), "No ai models found\n")
				return nil
			}

			if err = printAIModelsList(cmd.OutOrStdout(), output, models); err != nil {
				return redact.Errorf("could not print ai models list: %w", redact.Safe(err))
			}

			return nil
		},
	}

	cmd.Flags().VarP(&output, "output", "o", "Output format (wide, json, yaml)")

	return cmd
}

func printAIModelsList(writer io.Writer, format outputformat.Format, aimodels []client.AIModel) error {
	var err error

	switch format {
	case "json":
		enc := json.NewEncoder(writer)
		enc.SetIndent("", "  ")
		err = enc.Encode(aimodels)
	case "yaml":
		indent := 2
		enc := yaml.NewEncoder(writer)
		enc.SetIndent(indent)
		err = enc.Encode(aimodels)
	default:
		table := ux.TableFromObjects(aimodels, func(aimodel client.AIModel) []ux.Row {
			return []ux.Row{
				ux.NewRow("Name", aimodel.DisplayName),
				ux.NewRow("ID", aimodel.Slug),
				ux.NewRow("Description", aimodel.Description),
				ux.NewRow("Context Length", strconv.Itoa(aimodel.ContextLength)),
			}
		})

		ux.Fprintf(writer, "%s", table.String())
	}

	if err != nil {
		return redact.Errorf("output encoder failed: %w", redact.Safe(err))
	}

	return nil
}
