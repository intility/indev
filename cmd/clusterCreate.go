package cmd

import (
	"slices"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/intility/icpctl/internal/redact"
	"github.com/intility/icpctl/internal/telemetry"
	"github.com/intility/icpctl/internal/ux"
	"github.com/intility/icpctl/internal/wizard"
	"github.com/intility/icpctl/pkg/client"
)

const (
	maxNodeCount = 20
	minNodeCount = 2
)

var (
	clusterName string
	nodePreset  string
	nodeCount   int

	errInvalidPreset    = redact.Errorf("invalid node preset: preset must be one of minimal, balanced, performance")
	errInvalidNodeCount = redact.Errorf("invalid node count: count must be between 2 and 20")
)

// clusterCreateCmd represents the create command.
var clusterCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cluster",
	Long:  `Create a new cluster with the specified configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New()

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
					Placeholder: "Node Count (" + strconv.Itoa(minNodeCount) + "-" + strconv.Itoa(maxNodeCount) + ")",
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

		if nodeCount < minNodeCount || nodeCount > maxNodeCount {
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

		tracer, ok := telemetry.TracerFromContext(cmd.Context())
		if !ok {
			return redact.Errorf("could not get tracer from context")
		}

		ctx, span := tracer.Start(cmd.Context(), "CreateCluster")
		defer span.End()

		_, err := c.CreateCluster(ctx, req)
		if err != nil {
			return redact.Errorf("could not create cluster: %w", redact.Safe(err))
		}

		ux.Fsuccess(cmd.OutOrStdout(), "cluster created\n")

		return nil
	},
}

func init() {
	clusterCreateCmd.Flags().StringVarP(&clusterName, "name", "n", "",
		"Name of the cluster to create")

	clusterCreateCmd.Flags().StringVar(&nodePreset, "preset", "minimal",
		"Node preset to use (minimal, balanced, performance)")

	clusterCreateCmd.Flags().IntVar(&nodeCount, "nodes", minNodeCount,
		"Number of nodes to create (2-20)")

	clusterCmd.AddCommand(clusterCreateCmd)
}
