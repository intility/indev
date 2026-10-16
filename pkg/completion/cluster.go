package completion

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/intility/indev/pkg/clientset"
)

// ResourceClusters identifies the cluster name cache resource.
const ResourceClusters = "clusters"

// ClusterNames returns a completion function that suggests cluster names.
// It is intended for flags that take a cluster name.
func ClusterNames(set clientset.ClientSet) cobra.CompletionFunc {
	return completeNames(set, ResourceClusters, listClusterNames)
}

// ClusterNameArg returns a completion function that suggests cluster names
// for the first positional argument only.
func ClusterNameArg(set clientset.ClientSet) cobra.CompletionFunc {
	complete := ClusterNames(set)

	return func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return complete(cmd, args, toComplete)
	}
}

// RegisterClusterNameFlag enables cluster name completion for the given flag.
func RegisterClusterNameFlag(cmd *cobra.Command, flagName string, set clientset.ClientSet) {
	// Registration only fails if the flag does not exist, which is a programming error.
	cobra.CheckErr(cmd.RegisterFlagCompletionFunc(flagName, ClusterNames(set)))
}

func listClusterNames(ctx context.Context, set clientset.ClientSet) ([]string, error) {
	clusters, err := set.PlatformClient.ListClusters(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list clusters: %w", err)
	}

	return clusters.Names(), nil
}
