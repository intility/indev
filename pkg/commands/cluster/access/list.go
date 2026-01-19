package access

import (
	"encoding/json"
	"io"
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
	var (
		clusterName string
		clusterID   string
		output      = outputformat.Format("")
	)

	cmd := &cobra.Command{
		Use:     "list [cluster-name]",
		Short:   "List cluster members",
		Long:    `List all members who have access to a cluster and their roles.`,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "cluster.access.list")
			defer span.End()

			cmd.SilenceUsage = true

			// If positional argument is provided, use it (takes precedence)
			if len(args) > 0 {
				clusterName = args[0]
			}

			// Resolve cluster ID
			resolvedClusterID, err := resolveClusterID(ctx, set, clusterName, clusterID)
			if err != nil {
				return err
			}

			members, err := set.PlatformClient.GetClusterMembers(ctx, resolvedClusterID)
			if err != nil {
				return redact.Errorf("could not get cluster members: %w", redact.Safe(err))
			}

			if len(members) == 0 {
				ux.Fprint(cmd.OutOrStdout(), "No members found\n")
				return nil
			}

			if err = printMemberList(cmd.OutOrStdout(), output, members); err != nil {
				return redact.Errorf("could not print member list: %w", redact.Safe(err))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "Name of the cluster")
	cmd.Flags().StringVar(&clusterID, "cluster-id", "", "ID of the cluster")
	cmd.Flags().VarP(&output, "output", "o", "Output format (wide, json, yaml)")

	return cmd
}

func printMemberList(writer io.Writer, format outputformat.Format, members []client.ClusterMember) error {
	var err error

	switch format {
	case "wide":
		table := ux.TableFromObjects(members, func(member client.ClusterMember) []ux.Row {
			return []ux.Row{
				ux.NewRow("Name", member.Subject.Name),
				ux.NewRow("Type", member.Subject.Type),
				ux.NewRow("Roles", formatRoles(member.Roles)),
				ux.NewRow("Details", member.Subject.Details),
			}
		})

		ux.Fprint(writer, "%s", table.String())

		return nil
	case "json":
		enc := json.NewEncoder(writer)
		enc.SetIndent("", "  ")
		err = enc.Encode(members)
	case "yaml":
		indent := 2
		enc := yaml.NewEncoder(writer)
		enc.SetIndent(indent)
		err = enc.Encode(members)
	default:
		table := ux.TableFromObjects(members, func(member client.ClusterMember) []ux.Row {
			return []ux.Row{
				ux.NewRow("Name", member.Subject.Name),
				ux.NewRow("Type", member.Subject.Type),
				ux.NewRow("Roles", formatRoles(member.Roles)),
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

func formatRoles(roles []client.ClusterMemberRole) string {
	if len(roles) == 0 {
		return ""
	}

	roleStrings := make([]string, len(roles))
	for i, role := range roles {
		roleStrings[i] = string(role)
	}

	return strings.Join(roleStrings, ", ")
}
