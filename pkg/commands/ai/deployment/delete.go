package deployment

import (
	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/clientset"
)

func NewDeleteCommand(set clientset.ClientSet) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete an AI deployment",
		Long:    "Delete an AI deployment from the Intility Developer Platform",
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "aideployment.delete")
			defer span.End()

			cmd.SilenceUsage = true

			if len(args) > 0 {
				name = args[0]
			}

			if name == "" {
				return redact.Errorf("deployment name must be specified")
			}

			deploy, err := set.PlatformClient.GetAIDeployment(ctx, name)
			if err != nil {
				return redact.Errorf("could not find AI deployment: %w", redact.Safe(err))
			}

			if err = set.PlatformClient.DeleteAIDeployment(ctx, deploy.ID); err != nil {
				return redact.Errorf("could not delete AI deployment: %w", redact.Safe(err))
			}

			ux.Fsuccessf(cmd.OutOrStdout(), "deleted AI deployment: %s", name)

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the deployment to delete")

	return cmd
}
