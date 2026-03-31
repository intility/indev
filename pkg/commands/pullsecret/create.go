package pullsecret

import (
	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

var errEmptyPullSecretName = redact.Errorf("pull secret name cannot be empty")

// CreateOptions holds the options for the pull secret create command.
type CreateOptions struct {
	Name string
}

func NewCreateCommand(set clientset.ClientSet) *cobra.Command {
	var options CreateOptions

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a new image pull secret",
		Long:    `Create a new image pull secret with the specified name.`,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "pullsecret.create")
			defer span.End()

			if err := validateCreateOptions(options); err != nil {
				return err
			}

			cmd.SilenceUsage = true

			ps, err := set.PlatformClient.CreatePullSecret(ctx, client.NewPullSecretRequest{
				Name:       options.Name,
				Registries: make(map[string]client.PullSecretCredential),
			})
			if err != nil {
				return redact.Errorf("could not create pull secret: %w", redact.Safe(err))
			}

			ux.Fsuccessf(cmd.OutOrStdout(), "created pull secret: %s\n", ps.Name)

			return nil
		},
	}

	cmd.Flags().StringVarP(&options.Name, "name", "n", "", "Name of the pull secret to create")

	return cmd
}

func validateCreateOptions(options CreateOptions) error {
	if options.Name == "" {
		return errEmptyPullSecretName
	}

	return nil
}
