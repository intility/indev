package teams

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

var errInvalidOutputFormat = errors.New(`must be one of "wide", "json", "yaml"`)

func NewListCommand(set clientset.ClientSet) *cobra.Command {
	output := outputFormat("")
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all teams",
		Long:    `List all teams in the Intility Developer Platform`,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "teams.list")
			defer span.End()

			cmd.SilenceUsage = true

			teams, err := set.PlatformClient.ListTeams(ctx)
			if err != nil {
				return redact.Errorf("could not list teams: %w", redact.Safe(err))
			}

			if len(teams) == 0 {
				ux.Fprint(cmd.OutOrStdout(), "No teams found\n")
				return nil
			}

			if err = printTeamsList(cmd.OutOrStdout(), output, teams); err != nil {
				return redact.Errorf("could not print teams list: %w", redact.Safe(err))
			}

			return nil
		},
	}

	cmd.Flags().VarP(&output, "output", "o", "Output format (wide, json, yaml)")

	return cmd
}

func printTeamsList(writer io.Writer, format outputFormat, teams []client.Team) error {
	var err error

	switch format {
	case "wide":
		table := ux.TableFromObjects(teams, func(team client.Team) []ux.Row {
			return []ux.Row{
				ux.NewRow("Id", team.ID),
				ux.NewRow("Name", team.Name),
				ux.NewRow("Description", team.Description),
				ux.NewRow("Role", strings.Join(team.Role, ",")),
			}
		})

		ux.Fprint(writer, "%s", table.String())

		return nil
	case "json":
		enc := json.NewEncoder(writer)
		enc.SetIndent("", "  ")
		err = enc.Encode(teams)
	case "yaml":
		indent := 2
		enc := yaml.NewEncoder(writer)
		enc.SetIndent(indent)
		err = enc.Encode(teams)
	default:
		table := ux.TableFromObjects(teams, func(team client.Team) []ux.Row {
			return []ux.Row{
				ux.NewRow("Name", team.Name),
				ux.NewRow("Description", team.Description),
				ux.NewRow("Role", strings.Join(team.Role, ",")),
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
