package user

import (
	"encoding/json"
	"io"
	"slices"
	"sort"
	"strings"

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
		Short:   "List all users",
		Long:    `List all users in the Intility Developer Platform`,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "users.list")
			defer span.End()

			cmd.SilenceUsage = true

			users, err := set.PlatformClient.ListUsers(ctx)
			if err != nil {
				return redact.Errorf("could not list users: %w", redact.Safe(err))
			}

			if len(users) == 0 {
				ux.Fprint(cmd.OutOrStdout(), "No users found\n")
				return nil
			}

			if err = printUsersList(cmd.OutOrStdout(), output, users); err != nil {
				return redact.Errorf("could not print users list: %w", redact.Safe(err))
			}

			return nil
		},
	}

	cmd.Flags().VarP(&output, "output", "o", "Output format (wide, json, yaml)")

	return cmd
}

func printUsersList(writer io.Writer, format outputformat.Format, users []client.User) error {
	var err error

	sortUsersByOwnerThenName(users)

	switch format {
	case "wide":
		table := ux.TableFromObjects(users, func(user client.User) []ux.Row {
			return []ux.Row{
				ux.NewRow("Id", user.ID),
				ux.NewRow("Name", user.Name),
				ux.NewRow("UPN", user.UPN),
				ux.NewRow("Role", strings.Join(user.Roles, ",")),
			}
		})

		ux.Fprint(writer, "%s", table.String())
	case "json":
		enc := json.NewEncoder(writer)
		enc.SetIndent("", "  ")
		err = enc.Encode(users)
	case "yaml":
		indent := 2
		enc := yaml.NewEncoder(writer)
		enc.SetIndent(indent)
		err = enc.Encode(users)
	default:
		table := ux.TableFromObjects(users, func(user client.User) []ux.Row {
			return []ux.Row{
				ux.NewRow("Name", user.Name),
				ux.NewRow("UPN", user.UPN),
				ux.NewRow("Role", strings.Join(user.Roles, ",")),
			}
		})

		ux.Fprint(writer, "%s", table.String())
	}

	if err != nil {
		return redact.Errorf("output encoder failed: %w", redact.Safe(err))
	}

	return nil
}

func sortUsersByOwnerThenName(users []client.User) {
	sort.Slice(users, func(i, j int) bool {
		hasOwnerI := slices.Contains(users[i].Roles, "owner")
		hasOwnerJ := slices.Contains(users[j].Roles, "owner")

		if hasOwnerI != hasOwnerJ {
			return hasOwnerI // Owners first
		}

		return users[i].Name < users[j].Name // Then by name
	})
}
