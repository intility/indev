package pullsecret

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
		Short:   "List all image pull secrets",
		Long:    `List all image pull secrets in the Intility Developer Platform`,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "pullsecret.list")
			defer span.End()

			cmd.SilenceUsage = true

			pullSecrets, err := set.PlatformClient.ListPullSecrets(ctx)
			if err != nil {
				return redact.Errorf("could not list pull secrets: %w", redact.Safe(err))
			}

			if len(pullSecrets) == 0 {
				ux.Fprintf(cmd.OutOrStdout(), "No pull secrets found\n")
				return nil
			}

			if err = printPullSecretsList(cmd.OutOrStdout(), output, pullSecrets); err != nil {
				return redact.Errorf("could not print pull secrets list: %w", redact.Safe(err))
			}

			return nil
		},
	}

	cmd.Flags().VarP(&output, "output", "o", "Output format (wide, json, yaml)")

	return cmd
}

func printPullSecretsList(writer io.Writer, format outputformat.Format, pullSecrets []client.PullSecret) error {
	var err error

	switch format {
	case "wide":
		table := ux.TableFromObjects(pullSecrets, func(ps client.PullSecret) []ux.Row {
			return []ux.Row{
				ux.NewRow("Id", ps.ID),
				ux.NewRow("Name", ps.Name),
				ux.NewRow("Registries", strconv.Itoa(len(ps.Registries))),
				ux.NewRow("Created At", ps.CreatedAt),
			}
		})

		ux.Fprintf(writer, "%s", table.String())
	case "json":
		enc := json.NewEncoder(writer)
		enc.SetIndent("", "  ")
		err = enc.Encode(pullSecrets)
	case "yaml":
		indent := 2
		enc := yaml.NewEncoder(writer)
		enc.SetIndent(indent)
		err = enc.Encode(pullSecrets)
	default:
		table := ux.TableFromObjects(pullSecrets, func(ps client.PullSecret) []ux.Row {
			return []ux.Row{
				ux.NewRow("Name", ps.Name),
				ux.NewRow("Registries", strconv.Itoa(len(ps.Registries))),
				ux.NewRow("Created At", ps.CreatedAt),
			}
		})

		ux.Fprintf(writer, "%s", table.String())
	}

	if err != nil {
		return redact.Errorf("output encoder failed: %w", redact.Safe(err))
	}

	return nil
}
