package apikey

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
	defaultTTLDays = 30
	validNameRegex = "^[a-z0-9]([a-z0-9-]*[a-z0-9])?$"
)

var (
	errEmptyName         = redact.Errorf("API key name cannot be empty")
	errEmptyDeployment   = redact.Errorf("deployment name must be specified")
	errInvalidTTL        = redact.Errorf("TTL is required and must be a positive integer in the range 1 - 365")
	errInvalidNameLength = redact.Errorf(
		"API key name must be between %d and %d characters long",
		minNameLength, maxNameLength,
	)
	errInvalidNameFormat = redact.Errorf("API key name must match the pattern %s", validNameRegex)
)

type CreateOptions struct {
	Name       string
	Deployment string
	TTLDays    int
}

func NewCreateCommand(set clientset.ClientSet) *cobra.Command {
	var options CreateOptions

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a new API key",
		Long:    "Create a new API key for an AI deployment",
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "apikey.create")
			defer span.End()

			cmd.SilenceUsage = true

			err := validateCreateOptions(options)
			if err != nil {
				return err
			}

			deploy, err := set.PlatformClient.GetAIDeployment(ctx, options.Deployment)
			if err != nil {
				return redact.Errorf("could not find AI deployment: %w", redact.Safe(err))
			}

			key, err := set.PlatformClient.CreateAIAPIKey(ctx, deploy.ID, client.NewAIAPIKeyRequest{
				Name:    options.Name,
				TTLDays: options.TTLDays,
			})
			if err != nil {
				return redact.Errorf("could not create API key: %w", redact.Safe(err))
			}

			ux.Fsuccessf(cmd.OutOrStdout(), "API key created successfully\n")
			ux.Fprintf(cmd.OutOrStdout(), "\n  %s\n\n", key.Key)
			ux.Fwarningf(cmd.OutOrStdout(), "Store this key securely — it will not be shown again.\n")

			return nil
		},
	}

	cmd.Flags().StringVarP(&options.Name,
		"name", "n", "", "Name of the API key to create")

	cmd.Flags().StringVarP(&options.Deployment,
		"deployment", "d", "", "Name of the AI deployment")

	cmd.Flags().IntVarP(&options.TTLDays,
		"ttl", "t", defaultTTLDays, "Number of days the API key will be valid")

	return cmd
}

func validateCreateOptions(options CreateOptions) error {
	if options.Deployment == "" {
		return errEmptyDeployment
	}

	if options.Name == "" {
		return errEmptyName
	}

	if options.TTLDays <= 0 || options.TTLDays > 365 {
		return errInvalidTTL
	}

	if len(options.Name) < minNameLength || len(options.Name) > maxNameLength {
		return errInvalidNameLength
	}

	if matched, err := regexp.MatchString(validNameRegex, options.Name); err != nil || !matched {
		return errInvalidNameFormat
	}

	return nil
}
