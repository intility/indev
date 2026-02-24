package pullsecret

import (
	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/clientset"
	pullsecretcmd "github.com/intility/indev/pkg/commands/pullsecret"
)

var (
	errClusterNameRequired    = redact.Errorf("cluster name is required")
	errPullSecretNameRequired = redact.Errorf("pull secret name is required")
)

func NewSetCommand(set clientset.ClientSet) *cobra.Command {
	var (
		clusterName    string
		pullSecretName string
	)

	cmd := &cobra.Command{
		Use:     "set [cluster-name] [pull-secret-name]",
		Short:   "Set the image pull secret for a cluster",
		Long:    `Assign an image pull secret to a cluster.`,
		Args:    cobra.MaximumNArgs(2), //nolint:mnd
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "cluster.pullsecret.set")
			defer span.End()

			if len(args) > 0 {
				clusterName = args[0]
			}

			if len(args) > 1 {
				pullSecretName = args[1]
			}

			if clusterName == "" {
				return errClusterNameRequired
			}

			if pullSecretName == "" {
				return errPullSecretNameRequired
			}

			cmd.SilenceUsage = true

			cluster, err := set.PlatformClient.GetCluster(ctx, clusterName)
			if err != nil {
				return redact.Errorf("could not get cluster: %w", redact.Safe(err))
			}

			ps, err := pullsecretcmd.FindPullSecretByName(ctx, set.PlatformClient, pullSecretName)
			if err != nil {
				return redact.Errorf("could not find pull secret: %w", redact.Safe(err))
			}

			err = set.PlatformClient.SetClusterPullSecret(ctx, cluster.ID, ps.ID)
			if err != nil {
				return redact.Errorf("could not set cluster pull secret: %w", redact.Safe(err))
			}

			ux.Fsuccessf(cmd.OutOrStdout(), "set pull secret %s on cluster %s\n", pullSecretName, clusterName)

			return nil
		},
	}

	cmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "Name of the cluster")
	cmd.Flags().StringVarP(&pullSecretName, "pull-secret", "p", "", "Name of the pull secret")

	return cmd
}
