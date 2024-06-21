package cluster

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/intility/icpctl/internal/redact"
	"github.com/intility/icpctl/internal/telemetry"
	"github.com/intility/icpctl/internal/ux"
	"github.com/intility/icpctl/internal/wizard"
	"github.com/intility/icpctl/pkg/client"
	"github.com/intility/icpctl/pkg/clientset"
	"github.com/spf13/cobra"
)

const (
	maxCount = 20
	minCount = 2
)

func NewCreateCommand(set clientset.ClientSet) *cobra.Command {
	var (
		clusterName string
		nodePreset  string
		nodeCount   int

		errEmptyName        = redact.Errorf("cluster name cannot be empty")
		errInvalidPreset    = redact.Errorf("invalid node preset: preset must be one of minimal, balanced, performance")
		errInvalidNodeCount = redact.Errorf("invalid node count: count must be between 2 and 20")
	)

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a new cluster",
		Long:    `Create a new cluster with the specified configuration.`,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "cluster.create")
			defer span.End()

			if clusterName == "" {
				wz := wizard.NewWizard([]wizard.Input{
					{
						ID:          "name",
						Placeholder: "Cluster Name",
						Type:        wizard.InputTypeText,
						Limit:       0,
						Validator:   nil,
					},
					{
						ID:          "preset",
						Placeholder: "Node Preset (minimal, balanced, performance)",
						Type:        wizard.InputTypeText,
						Limit:       0,
						Validator:   nil,
					},
					{
						ID:          "nodes",
						Placeholder: "Node Count (" + strconv.Itoa(minCount) + "-" + strconv.Itoa(maxCount) + ")",
						Type:        wizard.InputTypeText,
						Limit:       0,
						Validator:   nil,
					},
				})

				result, err := wz.Run()
				if err != nil {
					return redact.Errorf("could not gather information: %w", redact.Safe(err))
				}

				if result.Cancelled() {
					return nil
				}

				clusterName = result.MustGetValue("name")
				nodePreset = result.MustGetValue("preset")
				nodeCountStr := result.MustGetValue("nodes")

				nodeCount, err = strconv.Atoi(nodeCountStr)
				if err != nil {
					return redact.Errorf("invalid node count: %w", redact.Safe(err))
				}
			}

			if clusterName == "" {
				return errEmptyName
			}

			if !slices.Contains([]string{"minimal", "balanced", "performance"}, nodePreset) {
				return errInvalidPreset
			}

			if nodeCount < minCount || nodeCount > maxCount {
				return errInvalidNodeCount
			}

			req := client.NewClusterRequest{
				Name: clusterName,
				NodePools: []client.NodePool{
					{
						Preset:    nodePreset,
						NodeCount: nodeCount,
					},
				},
			}
			cmd.SilenceUsage = true

			_, err := set.PlatformClient.CreateCluster(ctx, req)
			if err != nil {
				return redact.Errorf("could not create cluster: %w", redact.Safe(err))
			}

			ux.Fsuccess(cmd.OutOrStdout(), "cluster created\n")

			return nil
		},
	}

	cmd.Flags().StringVarP(&clusterName, "name", "n", "", "Name of the cluster to create")
	cmd.Flags().StringVar(&nodePreset, "preset", "minimal", "Node preset to use (minimal, balanced, performance)")
	cmd.Flags().IntVar(&nodeCount, "nodes", minCount, fmt.Sprintf("Number of nodes to create (%d-%d)", minCount, maxCount))

	return cmd
}
