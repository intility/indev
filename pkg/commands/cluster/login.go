package cluster

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
		Use:   "login [name]",
		Short: "Login to a cluster using oc",
		Long:  `Login to a cluster using the OpenShift CLI (oc). Opens a browser for OAuth authentication.`,
		Args:  cobra.MaximumNArgs(1),
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

			// List clusters to verify the cluster exists
			clusters, err := set.PlatformClient.ListClusters(ctx)
			if err != nil {
				return redact.Errorf("could not list clusters: %w", redact.Safe(err))
			}

			// Find the cluster with the matching name (case-insensitive)
			var found bool
			for _, c := range clusters {
				if strings.EqualFold(c.Name, clusterName) {
					clusterName = c.Name // Use the exact name from the API
					found = true
					break
				}
			}

			if !found {
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
				return redact.Errorf("oc command not found. Please install the OpenShift CLI: https://docs.openshift.com/container-platform/latest/cli_reference/openshift_cli/getting-started-cli.html")
			}

			ux.Fprint(cmd.OutOrStdout(), "Logging in to cluster %s...\n", clusterName)
			ux.Fprint(cmd.OutOrStdout(), "API URL: %s\n\n", apiURL)

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
