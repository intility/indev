package teams

import (
	"regexp"

	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

const (
	maxNameLength        = 50
	minNameLength        = 3
	validNameRegex       = "^[a-zA-Z0-9]+([-_ ]{0,1}[a-zA-Z0-9])+$"
	maxDescriptionLength = 100
	minDescriptionLength = 1
)

var (
	errEmptyName         = redact.Errorf("team name cannot be empty")
	errEmptyDescription  = redact.Errorf("team description cannot be empty")
	errInvalidNameLength = redact.Errorf("team name must be between %d and %d characters long", minNameLength, maxNameLength)
	errInvalidDescLength = redact.Errorf("team description must be between %d and %d characters long", minDescriptionLength, maxDescriptionLength)
	errInvalidNameFormat = redact.Errorf("team name must match the pattern %s", validNameRegex)
)

type CreateOptions struct {
	Name        string
	Description string
}

func NewCreateCommand(set clientset.ClientSet) *cobra.Command {
	var options CreateOptions

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a new team",
		Long:    `Create a new team with the specified configuration.`,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "team.create")
			defer span.End()

			var err error

			err = validateCreateOptions(options)
			if err != nil {
				return err
			}

			team, err := set.PlatformClient.CreateTeam(ctx, client.NewTeamRequest{
				Name:        options.Name,
				Description: options.Description,
			})
			if err != nil {
				return redact.Errorf("could not create team: %w", redact.Safe(err))
			}

			ux.Fsuccess(cmd.OutOrStdout(), "created team: %s (ID: %s)\n", team.Name, team.ID)

			return nil
		},
	}

	cmd.Flags().StringVarP(&options.Name,
		"name", "n", "", "Name of the team to create")

	cmd.Flags().StringVarP(&options.Description,
		"description", "d", "", "Description of the team to create")

	return cmd
}

func validateCreateOptions(options CreateOptions) error {
	if options.Name == "" {
		return errEmptyName
	}

	if matched, err := regexp.MatchString(validNameRegex, options.Name); err != nil || !matched {
		return errInvalidNameFormat
	}

	if options.Description == "" {
		return errEmptyDescription
	}

	if len(options.Name) < minNameLength || len(options.Name) > maxNameLength {
		return errInvalidNameLength
	}

	if len(options.Description) < minDescriptionLength || len(options.Description) > maxDescriptionLength {
		return errInvalidDescLength
	}

	return nil
}
