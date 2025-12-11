package member

import (
	"context"
	"fmt"
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
	errInvalidRole  = redact.Errorf("invalid role supplied, valid roles are: \"%s\"", strings.Join(client.GetMemberRoleValues(), ", "))
)

type AddMemberOptions struct {
	Team   string
	TeamId string
	User   string
	UserId string
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

			var err error
			err = validateOptions(options)
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

			err = set.PlatformClient.AddTeamMember(ctx, options.TeamId, []client.AddTeamMemberRequest{
				{
					Roles: []client.MemberRole{options.Role},
					Subject: client.AddMemberSubject{
						ID:   options.UserId,
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

			ux.Fsuccess(cmd.OutOrStdout(), "added user: %s (%s) to team: %s (%s) \n", options.User, options.UserId, options.Team, options.TeamId)

			return nil
		},
	}

	cmd.Flags().StringVarP(&options.Team,
		"team", "t", "", "Name of the team to add the member to")

	cmd.Flags().StringVar(&options.TeamId,
		"team-id", "", "ID of the team to add the member to")

	cmd.Flags().StringVarP(&options.User,
		"user", "u", "", "UPN of the user to add to the team")

	cmd.Flags().StringVar(&options.UserId,
		"user-id", "", "ID of the user to add to the team")

	roleFlagDescription := fmt.Sprintf("Role to assign to the new team member. Valid roles are: %s", strings.Join(client.GetMemberRoleValues(), ", "))
	cmd.Flags().StringVarP((*string)(&options.Role),
		"role", "r", "", roleFlagDescription)

	return cmd
}

func validateOptions(options AddMemberOptions) error {
	if options.TeamId == "" && options.Team == "" {
		return errTeamRequired
	}

	if options.UserId == "" && options.User == "" {
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

func getTeamIdByName(ctx context.Context, set clientset.ClientSet, teamName string) (string, error) {
	teams, err := set.PlatformClient.ListTeams(ctx)
	if err != nil {
		return "", redact.Errorf("could not list teams: %w", redact.Safe(err))
	}

	var team *client.Team
	for _, t := range teams {
		if strings.EqualFold(t.Name, teamName) {
			team = &t
			break
		}
	}

	if team == nil {
		return "", redact.Errorf("team not found: %s", teamName)
	}

	return team.ID, nil
}

func getUserIdByUpn(ctx context.Context, set clientset.ClientSet, upn string) (string, error) {
	users, err := set.PlatformClient.ListUsers(ctx)
	if err != nil {
		return "", redact.Errorf("could not list users: %w", redact.Safe(err))
	}

	var user *client.User
	for _, u := range users {
		if strings.EqualFold(u.UPN, upn) {
			user = &u
			break
		}
	}

	if user == nil {
		return "", redact.Errorf("user not found: %s", upn)
	}

	return user.ID, nil
}
