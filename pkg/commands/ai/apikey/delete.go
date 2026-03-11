package apikey

import (
	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/clientset"
)

func NewDeleteCommand(set clientset.ClientSet) *cobra.Command {
	var deployment string

	var name string

	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete an API key",
		Long:    "Delete an API key from an AI deployment",
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "apikey.delete")
			defer span.End()

			cmd.SilenceUsage = true

			if deployment == "" {
				return redact.Errorf("deployment name must be specified")
			}

			if name == "" {
				return redact.Errorf("API key name must be specified")
			}

			deploy, err := set.PlatformClient.GetAIDeployment(ctx, deployment)
			if err != nil {
				return redact.Errorf("could not find AI deployment: %w", redact.Safe(err))
			}

			key, err := set.PlatformClient.GetAIAPIKey(ctx, deploy.ID, name)
			if err != nil {
				return redact.Errorf("could not find API key: %w", redact.Safe(err))
			}

			if err = set.PlatformClient.DeleteAIAPIKey(ctx, deploy.ID, key.ID); err != nil {
				return redact.Errorf("could not delete API key: %w", redact.Safe(err))
			}

			ux.Fsuccessf(cmd.OutOrStdout(), "deleted API key: %s", name)

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the API key to delete")
	cmd.Flags().StringVarP(&deployment, "deployment", "d", "", "Name of the AI deployment")

	return cmd
}
