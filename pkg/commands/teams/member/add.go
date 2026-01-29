package member

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

var (
	errTeamRequired = redact.Errorf("team must be specified")
	errUserRequired = redact.Errorf("user must be specified")
	errRoleRequired = redact.Errorf("role must be specified")
	errInvalidRole  = redact.Errorf(
		"invalid role supplied, valid roles are: \"%s\"",
		strings.Join(client.GetMemberRoleValues(), ", "),
	)
)

type AddMemberOptions struct {
	Team   string
	TeamID string
	User   string
	UserID string
	Role   client.MemberRole
}

func NewAddCommand(set clientset.ClientSet) *cobra.Command {
	var options AddMemberOptions

	cmd := &cobra.Command{
		Use:     "add",
		Short:   "Add a new team member",
		Long:    `Add a new team member with the specified configuration.`,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "team.addMember")
			defer span.End()

			return runAddMemberCommand(ctx, cmd, set, options)
		},
	}

	cmd.Flags().StringVarP(&options.Team,
		"team", "t", "", "Name of the team to add the member to")

	cmd.Flags().StringVar(&options.TeamID,
		"team-id", "", "ID of the team to add the member to")

	cmd.Flags().StringVarP(&options.User,
		"user", "u", "", "UPN of the user to add to the team")

	cmd.Flags().StringVar(&options.UserID,
		"user-id", "", "ID of the user to add to the team")

	roleFlagDescription := "Role to assign to the new team member. Valid roles are: " +
		strings.Join(client.GetMemberRoleValues(), ", ")
	cmd.Flags().StringVarP((*string)(&options.Role),
		"role", "r", "", roleFlagDescription)

	return cmd
}

func runAddMemberCommand(
	ctx context.Context,
	cmd *cobra.Command,
	set clientset.ClientSet,
	options AddMemberOptions,
) error {
	err := validateAddOptions(options)
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

	err = set.PlatformClient.AddTeamMember(ctx, options.TeamID, []client.AddTeamMemberRequest{
		{
			Roles: []client.MemberRole{options.Role},
			Subject: client.AddMemberSubject{
				ID:   options.UserID,
				Type: "user",
			},
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "409 Conflict") {
			return redact.Errorf("user %s is already a member of team %s", options.User, options.Team)
		}

		return redact.Errorf("could not add team member: %w", redact.Safe(err))
	}

	ux.Fsuccessf(
		cmd.OutOrStdout(),
		"added user: %s (%s) to team: %s (%s)\n",
		options.User, options.UserID, options.Team, options.TeamID,
	)

	return nil
}

func validateAddOptions(options AddMemberOptions) error {
	if options.TeamID == "" && options.Team == "" {
		return errTeamRequired
	}

	if options.UserID == "" && options.User == "" {
		return errUserRequired
	}

	if options.Role == "" {
		return errRoleRequired
	}

	if !options.Role.IsValid() {
		return errInvalidRole
	}

	return nil
}
