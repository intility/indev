package cluster

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

// Printer is a function type for printing cluster information.
type Printer func(writer io.Writer, cluster *client.Cluster)

// lookupParams contains the parameters for a cluster lookup command.
type lookupParams struct {
	cmd         *cobra.Command
	set         clientset.ClientSet
	args        []string
	clusterName string
	printer     Printer
}

// runClusterLookupCommand executes the common logic for commands that
// look up a cluster by name and print information about it.
func runClusterLookupCommand(ctx context.Context, params lookupParams) error {
	cmd := params.cmd
	set := params.set
	args := params.args
	clusterName := params.clusterName
	printer := params.printer

	cmd.SilenceUsage = true

	if len(args) > 0 {
		clusterName = args[0]
	}

	if clusterName == "" {
		return redact.Errorf("cluster name cannot be empty")
	}

	cluster, err := set.PlatformClient.GetCluster(ctx, clusterName)
	if err != nil {
		return redact.Errorf("could not get cluster: %w", redact.Safe(err))
	}

	if cluster == nil {
		return redact.Errorf("cluster not found: %s", clusterName)
	}

	printer(cmd.OutOrStdout(), cluster)

	return nil
}
