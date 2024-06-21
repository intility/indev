package cluster

import (
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/intility/icpctl/internal/redact"
	"github.com/intility/icpctl/internal/telemetry"
	"github.com/intility/icpctl/internal/ux"
	"github.com/intility/icpctl/internal/wizard"
	"github.com/intility/icpctl/pkg/client"
	"github.com/intility/icpctl/pkg/clientset"
)

const (
	maxCount = 20
	minCount = 2
)

var (
	errCancelledByUser  = redact.Errorf("cancelled by user")
	errEmptyName        = redact.Errorf("cluster name cannot be empty")
	errInvalidPreset    = redact.Errorf("invalid node preset: preset must be one of minimal, balanced, performance")
	errInvalidNodeCount = redact.Errorf("invalid node count: count must be between 2 and 20")
)

type CreateOptions struct {
	Name      string
	Preset    string
	NodeCount int
}

func NewCreateCommand(set clientset.ClientSet) *cobra.Command {
	var options CreateOptions

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a new cluster",
		Long:    `Create a new cluster with the specified configuration.`,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "cluster.create")
			defer span.End()

			var err error
			if options.Name == "" {
				options, err = optionsFromWizard()
				if err != nil {
					if errors.Is(err, errCancelledByUser) {
						return nil
					}

					return redact.Errorf("could not get options from wizard: %w", redact.Safe(err))
				}
			}

			err = validateOptions(options)
			if err != nil {
				return err
			}

			// inputs validated, assume correct usage
			cmd.SilenceUsage = true

			_, err = set.PlatformClient.CreateCluster(ctx, client.NewClusterRequest{
				Name: options.Name,
				NodePools: []client.NodePool{
					{
						Preset:    options.Preset,
						NodeCount: options.NodeCount,
					},
				},
			})
			if err != nil {
				return redact.Errorf("could not create cluster: %w", redact.Safe(err))
			}

			ux.Fsuccess(cmd.OutOrStdout(), "cluster created\n")

			return nil
		},
	}

	cmd.Flags().StringVarP(&options.Name,
		"name", "n", "", "Name of the cluster to create")

	cmd.Flags().StringVar(&options.Preset,
		"preset", "minimal", "Node preset to use (minimal, balanced, performance)")

	cmd.Flags().IntVar(&options.NodeCount,
		"nodes", minCount, fmt.Sprintf("Number of nodes to create (%d-%d)", minCount, maxCount))

	return cmd
}

func optionsFromWizard() (CreateOptions, error) {
	var options CreateOptions

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
		return options, redact.Errorf("could not gather information: %w", redact.Safe(err))
	}

	if result.Cancelled() {
		return options, errCancelledByUser
	}

	options.Name = result.MustGetValue("name")
	options.Preset = result.MustGetValue("preset")
	nodeCountStr := result.MustGetValue("nodes")

	options.NodeCount, err = strconv.Atoi(nodeCountStr)
	if err != nil {
		return options, redact.Errorf("invalid node count: %w", redact.Safe(err))
	}

	return options, nil
}

func validateOptions(options CreateOptions) error {
	if options.Name == "" {
		return errEmptyName
	}

	if !slices.Contains([]string{"minimal", "balanced", "performance"}, options.Preset) {
		return errInvalidPreset
	}

	if options.NodeCount < minCount || options.NodeCount > maxCount {
		return errInvalidNodeCount
	}

	return nil
}
