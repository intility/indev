package member

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/clientset"
)

type RemoveMemberOptions struct {
	Team   string
	TeamId string
	User   string
	UserId string
}

func NewRemoveCommand(set clientset.ClientSet) *cobra.Command {
	var options RemoveMemberOptions

	cmd := &cobra.Command{
		Use:     "remove",
		Short:   "Remove a team member",
		Long:    `Remove a member from a team, revoking all their roles.`,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "team.removeMember")
			defer span.End()

			err := validateRemoveOptions(options)
			if err != nil {
				return err
			}

			if options.TeamId == "" {
				teamId, err := getTeamIdByName(ctx, set, options.Team)
				if err != nil {
					return err
				}
				options.TeamId = teamId
			}

			if options.UserId == "" {
				userId, err := getUserIdByUpn(ctx, set, options.User)
				if err != nil {
					return err
				}
				options.UserId = userId
			}

			memberId := "user:" + options.UserId

			err = set.PlatformClient.RemoveTeamMember(ctx, options.TeamId, memberId)
			if err != nil {
				if strings.Contains(err.Error(), "404 Not Found") {
					return redact.Errorf("user %s is not a member of team %s", options.User, options.Team)
				}
				return redact.Errorf("could not remove team member: %w", redact.Safe(err))
			}

			ux.Fsuccess(cmd.OutOrStdout(), "removed user: %s (%s) from team: %s (%s)\n", options.User, options.UserId, options.Team, options.TeamId)

			return nil
		},
	}

	cmd.Flags().StringVarP(&options.Team,
		"team", "t", "", "Name of the team to remove the member from")

	cmd.Flags().StringVar(&options.TeamId,
		"team-id", "", "ID of the team to remove the member from")

	cmd.Flags().StringVarP(&options.User,
		"user", "u", "", "UPN of the user to remove from the team")

	cmd.Flags().StringVar(&options.UserId,
		"user-id", "", "ID of the user to remove from the team")

	return cmd
}

func validateRemoveOptions(options RemoveMemberOptions) error {
	if options.TeamId == "" && options.Team == "" {
		return errTeamRequired
	}

	if options.UserId == "" && options.User == "" {
		return errUserRequired
	}

	return nil
}
