package registry

import (
	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
	pullsecretcmd "github.com/intility/indev/pkg/commands/pullsecret"
)

var (
	errEmptyPullSecretName = redact.Errorf("pull secret name cannot be empty")
	errEmptyAddress        = redact.Errorf("registry address cannot be empty")
	errEmptyUsername       = redact.Errorf("registry username cannot be empty")
	errEmptyPassword       = redact.Errorf("registry password cannot be empty")
)

// AddOptions holds the options for the registry add command.
type AddOptions struct {
	PullSecretName string
	Address        string
	Username       string
	Password       string //nolint:gosec // G117: this is a credential payload, not a hardcoded secret
}

func NewAddCommand(set clientset.ClientSet) *cobra.Command {
	var options AddOptions

	cmd := &cobra.Command{
		Use:     "add",
		Short:   "Add a registry to a pull secret",
		Long:    `Add a registry with credentials to an existing image pull secret.`,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "pullsecret.registry.add")
			defer span.End()

			if err := validateAddOptions(options); err != nil {
				return err
			}

			cmd.SilenceUsage = true

			ps, err := pullsecretcmd.FindPullSecretByName(ctx, set.PlatformClient, options.PullSecretName)
			if err != nil {
				return redact.Errorf("could not find pull secret: %w", redact.Safe(err))
			}

			_, err = set.PlatformClient.EditPullSecret(ctx, ps.ID, client.EditPullSecretRequest{
				Registries: map[string]*client.PullSecretCredential{
					options.Address: {
						Username: options.Username,
						Password: options.Password,
					},
				},
			})
			if err != nil {
				return redact.Errorf("could not add registry: %w", redact.Safe(err))
			}

			ux.Fsuccessf(cmd.OutOrStdout(), "added registry %s to pull secret: %s\n", options.Address, options.PullSecretName)

			return nil
		},
	}

	cmd.Flags().StringVarP(&options.PullSecretName, "pull-secret", "p", "", "Name of the pull secret")
	cmd.Flags().StringVarP(&options.Address, "address", "a", "", "Registry address (e.g. ghcr.io)")
	cmd.Flags().StringVarP(&options.Username, "username", "u", "", "Registry username")
	cmd.Flags().StringVar(&options.Password, "password", "", "Registry password")

	return cmd
}

func validateAddOptions(options AddOptions) error {
	if options.PullSecretName == "" {
		return errEmptyPullSecretName
	}

	if options.Address == "" {
		return errEmptyAddress
	}

	if options.Username == "" {
		return errEmptyUsername
	}

	if options.Password == "" {
		return errEmptyPassword
	}

	return nil
}
