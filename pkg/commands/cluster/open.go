package cluster

import (
	"strings"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

func NewOpenCommand(set clientset.ClientSet) *cobra.Command {
	var (
		clusterName  string
		errEmptyName = redact.Errorf("cluster name cannot be empty")
	)

	cmd := &cobra.Command{
		Use:   "open [name]",
		Short: "Open the cluster console in a browser",
		Long:  `Open the OpenShift web console for the specified cluster in your default browser.`,
		Args:  cobra.MaximumNArgs(1),
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "cluster.open")
			defer span.End()

			cmd.SilenceUsage = true

			// If positional argument is provided, use it (takes precedence)
			if len(args) > 0 {
				clusterName = args[0]
			}

			if clusterName == "" {
				return errEmptyName
			}

			// List clusters to find the one by name
			clusters, err := set.PlatformClient.ListClusters(ctx)
			if err != nil {
				return redact.Errorf("could not list clusters: %w", redact.Safe(err))
			}

			// Find the cluster with the matching name (case-insensitive)
			var cluster *client.Cluster
			for _, c := range clusters {
				if strings.EqualFold(c.Name, clusterName) {
					cluster = &c
					break
				}
			}

			if cluster == nil {
				return redact.Errorf("cluster not found: %s", clusterName)
			}

			if cluster.ConsoleURL == "" {
				return redact.Errorf("console URL not available for cluster: %s", clusterName)
			}

			ux.Fprint(cmd.OutOrStdout(), "Opening console for cluster %s...\n", cluster.Name)
			ux.Fprint(cmd.OutOrStdout(), "URL: %s\n", cluster.ConsoleURL)

			if err := browser.OpenURL(cluster.ConsoleURL); err != nil {
				return redact.Errorf("could not open browser: %w", redact.Safe(err))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&clusterName, "name", "n", "", "Name of the cluster")

	return cmd
}
