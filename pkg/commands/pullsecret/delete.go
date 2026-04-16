package pullsecret

import (
	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/clientset"
)

func NewDeleteCommand(set clientset.ClientSet) *cobra.Command {
	var (
		name         string
		errEmptyName = redact.Errorf("pull secret name cannot be empty")
	)

	cmd := &cobra.Command{
		Use:     "delete [name]",
		Short:   "Delete an existing image pull secret",
		Long:    `Delete an existing image pull secret by name.`,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "pullsecret.delete")
			defer span.End()

			if len(args) > 0 {
				name = args[0]
			}

			if name == "" {
				return errEmptyName
			}

			cmd.SilenceUsage = true

			ps, err := FindPullSecretByName(ctx, set.PlatformClient, name)
			if err != nil {
				return err
			}

			err = set.PlatformClient.DeletePullSecret(ctx, ps.ID)
			if err != nil {
				return redact.Errorf("could not delete pull secret: %w", redact.Safe(err))
			}

			ux.Fsuccessf(cmd.OutOrStdout(), "deleted pull secret: %s\n", name)

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the pull secret to delete")

	return cmd
}
