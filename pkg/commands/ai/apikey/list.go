package apikey

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
	var deployment string

	output := outputformat.Format("")

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List API keys",
		Long:    "List all API keys for an AI deployment",
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "apikey.list")
			defer span.End()

			cmd.SilenceUsage = true

			if deployment == "" {
				return redact.Errorf("deployment name must be specified")
			}

			deploy, err := set.PlatformClient.GetAIDeployment(ctx, deployment)
			if err != nil {
				return redact.Errorf("could not find AI deployment: %w", redact.Safe(err))
			}

			keys, err := set.PlatformClient.ListAIAPIKeys(ctx, deploy.ID)
			if err != nil {
				return redact.Errorf("could not list API keys: %w", redact.Safe(err))
			}

			if len(keys) == 0 {
				ux.Fprintf(cmd.OutOrStdout(), "No API keys found\n")
				return nil
			}

			if err = printAPIKeyList(cmd.OutOrStdout(), output, keys); err != nil {
				return redact.Errorf("could not print API keys: %w", redact.Safe(err))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&deployment, "deployment", "d", "", "Name of the AI deployment")
	cmd.Flags().VarP(&output, "output", "o", "Output format (wide, json, yaml)")

	return cmd
}

func printAPIKeyList(writer io.Writer, format outputformat.Format, keys []client.AIAPIKey) error {
	var err error

	switch format {
	case "json":
		enc := json.NewEncoder(writer)
		enc.SetIndent("", "  ")
		err = enc.Encode(keys)
	case "yaml":
		indent := 2
		enc := yaml.NewEncoder(writer)
		enc.SetIndent(indent)
		err = enc.Encode(keys)
	default:
		table := ux.TableFromObjects(keys, func(k client.AIAPIKey) []ux.Row {
			return []ux.Row{
				ux.NewRow("Name", k.Name),
				ux.NewRow("Prefix", k.Prefix),
				ux.NewRow("Created At", k.CreatedAt),
				ux.NewRow("Expires At", k.ExpiresAt),
			}
		})

		ux.Fprintf(writer, "%s", table.String())
	}

	if err != nil {
		return redact.Errorf("output encoder failed: %w", redact.Safe(err))
	}

	return nil
}
