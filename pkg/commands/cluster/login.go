package cluster

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/clientset"
)

const mustafarTenantID = "93e01775-815e-4327-83d4-5f9ad73b5aa1"

func NewLoginCommand(set clientset.ClientSet) *cobra.Command {
	var (
		clusterName  string
		errEmptyName = redact.Errorf("cluster name cannot be empty")
	)

	cmd := &cobra.Command{
		Use:     "login [name]",
		Short:   "Login to a cluster using oc",
		Long:    `Login to a cluster using the OpenShift CLI (oc). Opens a browser for OAuth authentication.`,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "cluster.login")
			defer span.End()

			cmd.SilenceUsage = true

			// If positional argument is provided, use it (takes precedence)
			if len(args) > 0 {
				clusterName = args[0]
			}

			if clusterName == "" {
				return errEmptyName
			}

			// Get cluster by name
			cluster, err := set.PlatformClient.GetCluster(ctx, clusterName)
			if err != nil {
				return redact.Errorf("could not get cluster: %w", redact.Safe(err))
			}

			if cluster == nil {
				return redact.Errorf("cluster not found: %s", clusterName)
			}

			// Get tenant ID to determine API URL
			tenantID, err := set.GetTenantID(ctx)
			if err != nil {
				return redact.Errorf("could not get tenant ID: %w", redact.Safe(err))
			}

			apiURL := getAPIURL(clusterName, tenantID)

			// Check if oc is installed
			if _, err := exec.LookPath("oc"); err != nil {
				return redact.Errorf("oc command not found. Please install the OpenShift CLI: https://developers.intility.com/docs/getting-started/first-steps/deploy-first-application/?h=oc#install-openshift-cli")
			}

			ux.Fprint(cmd.OutOrStdout(), "Logging in to cluster %s...\n\n", clusterName)

			// Execute oc login with web authentication
			ocCmd := exec.CommandContext(ctx, "oc", "login", "-w", apiURL)
			ocCmd.Stdin = os.Stdin
			ocCmd.Stdout = cmd.OutOrStdout()
			ocCmd.Stderr = cmd.ErrOrStderr()

			if err := ocCmd.Run(); err != nil {
				return redact.Errorf("oc login failed: %w", redact.Safe(err))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&clusterName, "name", "n", "", "Name of the cluster")

	return cmd
}

// getAPIURL returns the API URL for the cluster based on the tenant ID.
func getAPIURL(clusterName, tenantID string) string {
	if tenantID == mustafarTenantID {
		return fmt.Sprintf("https://api-%s.clusters.zone-1.xcv.net", clusterName)
	}
	return fmt.Sprintf("https://api-%s.apps.intilitycloud.com", clusterName)
}
