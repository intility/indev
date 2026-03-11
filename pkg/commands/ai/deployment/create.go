package deployment

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
	maxNameLength  = 50
	minNameLength  = 3
	validNameRegex = "^[a-z0-9]([a-z0-9-]*[a-z0-9])?$"
)

var (
	errEmptyName         = redact.Errorf("deployment name cannot be empty")
	errEmptyModel        = redact.Errorf("model must be specified")
	errInvalidNameLength = redact.Errorf(
		"deployment name must be between %d and %d characters long",
		minNameLength, maxNameLength,
	)
	errInvalidNameFormat = redact.Errorf("deployment name must match the pattern %s", validNameRegex)
)

type CreateOptions struct {
	Name  string
	Model string
}

func NewCreateCommand(set clientset.ClientSet) *cobra.Command {
	var options CreateOptions

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a new AI deployment",
		Long:    "Create a new AI deployment with specified model",
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "aideployment.create")
			defer span.End()

			cmd.SilenceUsage = true

			err := validateCreateOptions(options)
			if err != nil {
				return err
			}

			aideployment, err := set.PlatformClient.CreateAIDeployment(ctx, client.NewAIDeploymentRequest{
				Name:  options.Name,
				Model: options.Model,
			})
			if err != nil {
				return redact.Errorf("could not create AI deployment: %w", redact.Safe(err))
			}

			ux.Fsuccessf(cmd.OutOrStdout(), "created AI deployment: %s", aideployment.Name)

			return nil
		},
	}

	cmd.Flags().StringVarP(&options.Name,
		"name", "n", "", "Name of the deployment to create")

	cmd.Flags().StringVarP(&options.Model,
		"model", "m", "", "ID of the AI model to use. Available models can be found using 'indev ai model list'")

	return cmd
}

func validateCreateOptions(options CreateOptions) error {
	if options.Name == "" {
		return errEmptyName
	}

	if options.Model == "" {
		return errEmptyModel
	}

	if len(options.Name) < minNameLength || len(options.Name) > maxNameLength {
		return errInvalidNameLength
	}

	if matched, err := regexp.MatchString(validNameRegex, options.Name); err != nil || !matched {
		return errInvalidNameFormat
	}

	return nil
}
