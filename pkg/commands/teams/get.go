package teams

import (
	"io"
	"slices"

	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

func NewGetCommand(set clientset.ClientSet) *cobra.Command {
	var (
		teamName     string
		errEmptyName = redact.Errorf("team name cannot be empty")
	)

	cmd := &cobra.Command{
		Use:     "get [name]",
		Short:   "Get detailed information about a team",
		Long:    `Display comprehensive information about a specific team`,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "team.get")
			defer span.End()

			cmd.SilenceUsage = true

			// If positional argument is provided, use it (takes precedence)
			if len(args) > 0 {
				teamName = args[0]
			}

			if teamName == "" {
				return errEmptyName
			}

			// List teams to find the one by name
			teams, err := set.PlatformClient.ListTeams(ctx)
			if err != nil {
				return redact.Errorf("could not get team: %w", redact.Safe(err))
			}

			// Find the team with the matching name
			var team *client.Team
			for _, c := range teams {
				if c.Name == teamName {
					team = &c
					break
				}
			}

			if team == nil {
				return redact.Errorf("team not found: %s", teamName)
			}

			members, err := set.PlatformClient.GetTeamMembers(ctx, team.ID)
			if err != nil {
				return redact.Errorf("could not get members from team: %w", redact.Safe(err))
			}

			if err = printTeamDetails(cmd.OutOrStdout(), team, members); err != nil {
				return redact.Errorf("could not print team details: %w", redact.Safe(err))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&teamName, "name", "n", "", "Name of the team")

	return cmd
}

func printTeamDetails(writer io.Writer, team *client.Team, members []client.TeamMember) error {
	ux.Fprint(writer, "Team Information:\n")
	ux.Fprint(writer, "  ID:          	%s\n", team.ID)
	ux.Fprint(writer, "  Name:        	%s\n", team.Name)
	ux.Fprint(writer, "  Description:  %s\n", team.Description)
	ux.Fprint(writer, "Members:\n")

	table := ux.TableFromObjects(members, func(member client.TeamMember) []ux.Row {
		return []ux.Row{
			ux.NewRow("  Name", "  "+member.Subject.Name),
			ux.NewRow("  Role", "  "+getTeamRole(member.Roles)),
		}
	})

	ux.Fprint(writer, "%s", table.String())

	return nil
}

func getTeamRole(roles []client.MemberRole) string {
	if slices.Contains(roles, client.MemberRoleOwner) {
		return "Owner"
	}
	if slices.Contains(roles, client.MemberRoleMember) {
		return "Member"
	}
	if len(roles) == 0 {
		return "None"
	}
	return "Unknown"
}
