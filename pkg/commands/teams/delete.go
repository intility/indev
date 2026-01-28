package teams

import (
	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

func NewDeleteCommand(set clientset.ClientSet) *cobra.Command {
	var (
		teamName     string
		errEmptyName = redact.Errorf("team name cannot be empty")
	)

	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete an existing team",
		Long:    `Delete an existing team with the specified name.`,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "team.delete")
			defer span.End()

			var err error

			if teamName == "" {
				return errEmptyName
			}

			team, err := set.PlatformClient.GetTeam(ctx, teamName)
			if err != nil {
				return redact.Errorf("could not get team: %w", redact.Safe(err))
			}

			if team == nil {
				return redact.Errorf("team not found: %s", teamName)
			}

			err = set.PlatformClient.DeleteTeam(ctx, client.DeleteTeamRequest{
				TeamID: team.ID,
			})
			if err != nil {
				return redact.Errorf("could not delete team: %w", redact.Safe(err))
			}

			ux.Fsuccess(cmd.OutOrStdout(), "deleted team: %s\n", teamName)

			return nil
		},
	}

	cmd.Flags().StringVarP(&teamName,
		"name", "n", "", "Name of the team to delete")

	return cmd
}
