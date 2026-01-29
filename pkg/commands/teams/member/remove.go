package member

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/clientset"
)

type RemoveMemberOptions struct {
	Team   string
	TeamID string
	User   string
	UserID string
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

			return runRemoveMemberCommand(ctx, cmd, set, options)
		},
	}

	cmd.Flags().StringVarP(&options.Team,
		"team", "t", "", "Name of the team to remove the member from")

	cmd.Flags().StringVar(&options.TeamID,
		"team-id", "", "ID of the team to remove the member from")

	cmd.Flags().StringVarP(&options.User,
		"user", "u", "", "UPN of the user to remove from the team")

	cmd.Flags().StringVar(&options.UserID,
		"user-id", "", "ID of the user to remove from the team")

	return cmd
}

func runRemoveMemberCommand(
	ctx context.Context,
	cmd *cobra.Command,
	set clientset.ClientSet,
	options RemoveMemberOptions,
) error {
	err := validateRemoveOptions(options)
	if err != nil {
		return err
	}

	if options.TeamID == "" {
		teamID, err := getTeamIDByName(ctx, set, options.Team)
		if err != nil {
			return err
		}

		options.TeamID = teamID
	}

	if options.UserID == "" {
		userID, err := getUserIDByUpn(ctx, set, options.User)
		if err != nil {
			return err
		}

		options.UserID = userID
	}

	memberID := "user:" + options.UserID

	err = set.PlatformClient.RemoveTeamMember(ctx, options.TeamID, memberID)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return redact.Errorf("user %s is not a member of team %s", options.User, options.Team)
		}

		return redact.Errorf("could not remove team member: %w", redact.Safe(err))
	}

	ux.Fsuccessf(
		cmd.OutOrStdout(),
		"removed user: %s (%s) from team: %s (%s)\n",
		options.User, options.UserID, options.Team, options.TeamID,
	)

	return nil
}

func validateRemoveOptions(options RemoveMemberOptions) error {
	if options.TeamID == "" && options.Team == "" {
		return errTeamRequired
	}

	if options.UserID == "" && options.User == "" {
		return errUserRequired
	}

	return nil
}
