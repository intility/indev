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

// RemoveOptions holds the options for the registry remove command.
type RemoveOptions struct {
	PullSecretName string
	Address        string
}

func NewRemoveCommand(set clientset.ClientSet) *cobra.Command {
	var options RemoveOptions

	cmd := &cobra.Command{
		Use:     "remove",
		Short:   "Remove a registry from a pull secret",
		Long:    `Remove a registry from an existing image pull secret.`,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "pullsecret.registry.remove")
			defer span.End()

			if err := validateRemoveOptions(options); err != nil {
				return err
			}

			cmd.SilenceUsage = true

			ps, err := pullsecretcmd.FindPullSecretByName(ctx, set.PlatformClient, options.PullSecretName)
			if err != nil {
				return redact.Errorf("could not find pull secret: %w", redact.Safe(err))
			}

			_, err = set.PlatformClient.EditPullSecret(ctx, ps.ID, client.EditPullSecretRequest{
				Registries: map[string]*client.PullSecretCredential{
					options.Address: nil,
				},
			})
			if err != nil {
				return redact.Errorf("could not remove registry: %w", redact.Safe(err))
			}

			ux.Fsuccessf(cmd.OutOrStdout(), "removed registry %s from pull secret: %s\n",
				options.Address, options.PullSecretName)

			return nil
		},
	}

	cmd.Flags().StringVarP(&options.PullSecretName, "pull-secret", "p", "", "Name of the pull secret")
	cmd.Flags().StringVarP(&options.Address, "address", "a", "", "Registry address to remove")

	return cmd
}

func validateRemoveOptions(options RemoveOptions) error {
	if options.PullSecretName == "" {
		return errEmptyPullSecretName
	}

	if options.Address == "" {
		return errEmptyAddress
	}

	return nil
}
