package cluster

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/internal/wizard"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

const (
	maxCount     = 8
	minCount     = 2
	suffixLength = 6
)

var (
	errCancelledByUser   = redact.Errorf("cancelled by user")
	errEmptyName         = redact.Errorf("cluster name cannot be empty")
	errInvalidPreset     = redact.Errorf("invalid node preset: preset must be one of minimal, balanced, performance")
	errInvalidNodeCount  = redact.Errorf("invalid node count: count must be between %d and %d", minCount, maxCount)
	errInvalidMinNodes   = redact.Errorf("invalid minimum node count: count must be between %d and %d", minCount, maxCount)
	errInvalidMaxNodes   = redact.Errorf("invalid maximum node count: count must be between %d and %d", minCount, maxCount)
	errMinGreaterThanMax = redact.Errorf("minimum node count cannot be greater than maximum node count")
)

type CreateOptions struct {
	Name              string
	Preset            string
	NodeCount         int // Used when autoscaling is disabled
	EnableAutoscaling bool
	MinNodes          int // Used when autoscaling is enabled
	MaxNodes          int // Used when autoscaling is enabled
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

			// Fetch SSO provisioner
			ssoProvisioner, err := selectSSOProvisioner(ctx, set.PlatformClient, cmd.OutOrStdout())
			if err != nil {
				return redact.Errorf("could not select SSO provisioner: %w", redact.Safe(err))
			}

			clusterName := options.Name + "-" + generateSuffix()

			// Create node pool based on autoscaling configuration
			nodePool := client.NodePool{
				Preset: options.Preset,
			}

			if options.EnableAutoscaling {
				nodePool.AutoscalingEnabled = true
				nodePool.MinCount = &options.MinNodes
				nodePool.MaxCount = &options.MaxNodes
			} else {
				replicas := options.NodeCount
				nodePool.Replicas = &replicas
			}

			_, err = set.PlatformClient.CreateCluster(ctx, client.NewClusterRequest{
				Name:           clusterName,
				SSOProvisioner: ssoProvisioner,
				NodePools:      []client.NodePool{nodePool},
			})
			if err != nil {
				return redact.Errorf("could not create cluster: %w", redact.Safe(err))
			}

			ux.Fsuccess(cmd.OutOrStdout(), "created cluster: %s\n", clusterName)

			return nil
		},
	}

	cmd.Flags().StringVarP(&options.Name,
		"name", "n", "", "Name of the cluster to create")

	cmd.Flags().StringVar(&options.Preset,
		"preset", "minimal", "Node preset to use (minimal, balanced, performance)")

	cmd.Flags().IntVar(&options.NodeCount,
		"nodes", minCount, fmt.Sprintf("Number of nodes to create (%d-%d)", minCount, maxCount))

	cmd.Flags().BoolVar(&options.EnableAutoscaling,
		"enable-autoscaling", false, "Enable autoscaling for the node pool")

	cmd.Flags().IntVar(&options.MinNodes,
		"min-nodes", minCount, fmt.Sprintf("Minimum number of nodes when autoscaling is enabled (%d-%d)", minCount, maxCount))

	cmd.Flags().IntVar(&options.MaxNodes,
		"max-nodes", maxCount, fmt.Sprintf("Maximum number of nodes when autoscaling is enabled (%d-%d)", minCount, maxCount))

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
			Placeholder: "Node Preset",
			Type:        wizard.InputTypeSelect,
			Options:     []string{"minimal", "balanced", "performance"},
		},
		{
			ID:          "autoscaling",
			Placeholder: "Enable autoscaling",
			Type:        wizard.InputTypeToggle,
			Options:     []string{"no", "yes"},
		},
		{
			ID:          "nodes",
			Placeholder: "Node Count (" + strconv.Itoa(minCount) + "-" + strconv.Itoa(maxCount) + ")",
			Type:        wizard.InputTypeText,
			ShowWhen: func(answers map[string]wizard.Answer) bool {
				// Show this field only when autoscaling is disabled
				return answers["autoscaling"].Value == "no"
			},
		},
		{
			ID:          "minNodes",
			Placeholder: "Minimum Nodes (" + strconv.Itoa(minCount) + "-" + strconv.Itoa(maxCount) + ")",
			Type:        wizard.InputTypeText,
			ShowWhen: func(answers map[string]wizard.Answer) bool {
				// Show this field only when autoscaling is enabled
				return answers["autoscaling"].Value == "yes"
			},
		},
		{
			ID:          "maxNodes",
			Placeholder: "Maximum Nodes (" + strconv.Itoa(minCount) + "-" + strconv.Itoa(maxCount) + ")",
			Type:        wizard.InputTypeText,
			ShowWhen: func(answers map[string]wizard.Answer) bool {
				// Show this field only when autoscaling is enabled
				return answers["autoscaling"].Value == "yes"
			},
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
	options.EnableAutoscaling = result.MustGetValue("autoscaling") == "yes"

	if options.EnableAutoscaling {
		minNodesStr := result.MustGetValue("minNodes")
		options.MinNodes, err = strconv.Atoi(minNodesStr)
		if err != nil {
			return options, redact.Errorf("invalid minimum node count: %w", redact.Safe(err))
		}

		maxNodesStr := result.MustGetValue("maxNodes")
		options.MaxNodes, err = strconv.Atoi(maxNodesStr)
		if err != nil {
			return options, redact.Errorf("invalid maximum node count: %w", redact.Safe(err))
		}
	} else {
		nodeCountStr := result.MustGetValue("nodes")
		options.NodeCount, err = strconv.Atoi(nodeCountStr)
		if err != nil {
			return options, redact.Errorf("invalid node count: %w", redact.Safe(err))
		}
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

	if options.EnableAutoscaling {
		// Validate autoscaling parameters
		if options.MinNodes < minCount || options.MinNodes > maxCount {
			return errInvalidMinNodes
		}

		if options.MaxNodes < minCount || options.MaxNodes > maxCount {
			return errInvalidMaxNodes
		}

		if options.MinNodes > options.MaxNodes {
			return errMinGreaterThanMax
		}
	} else {
		// Validate fixed node count
		if options.NodeCount < minCount || options.NodeCount > maxCount {
			return errInvalidNodeCount
		}
	}

	return nil
}

func generateSuffix() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	suffix := make([]rune, suffixLength)
	for i := range suffix {
		suffix[i] = letters[rand.IntN(len(letters))]
	}

	return string(suffix)
}

func selectSSOProvisioner(ctx context.Context, platformClient client.Client, out interface{ Write([]byte) (int, error) }) (string, error) {
	instances, err := platformClient.ListIntegrationInstances(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list integration instances: %w", err)
	}

	// Filter for EntraID type
	var provisioners []client.IntegrationInstance
	for _, instance := range instances {
		if instance.Type == "EntraID" {
			provisioners = append(provisioners, instance)
		}
	}

	switch len(provisioners) {
	case 0:
		return "", redact.Errorf("no SSO provisioner configured for your organization")
	case 1:
		ux.Fprint(out, "Using SSO provisioner: %s\n", provisioners[0].Name)
		return provisioners[0].ID, nil
	default:
		// For now, auto-select the first one. Future: add wizard step
		ux.Fprint(out, "Using SSO provisioner: %s\n", provisioners[0].Name)
		return provisioners[0].ID, nil
	}
}
