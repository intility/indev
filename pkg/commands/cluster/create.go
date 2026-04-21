package cluster

import (
	"context"
	"errors"
	"fmt"
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
	answerYes    = "yes"
	answerNo     = "no"
	noPullSecret = "(none)"
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
	PullSecret        string
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

			return runCreateCommand(ctx, cmd, set, options)
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

	cmd.Flags().StringVar(&options.PullSecret,
		"pull-secret", "", "Name of the image pull secret to use")

	return cmd
}

func runCreateCommand(ctx context.Context, cmd *cobra.Command, set clientset.ClientSet, options CreateOptions) error {
	var err error

	// Fetch pull secrets for wizard
	pullSecrets, psErr := set.PlatformClient.ListPullSecrets(ctx)
	if psErr != nil {
		pullSecrets = nil
	}

	if options.Name == "" {
		options, err = optionsFromWizard(pullSecrets)
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

	nodePool := buildNodePool(options)

	pullSecretRef, err := resolvePullSecretRef(options.PullSecret, pullSecrets)
	if err != nil {
		return err
	}

	cluster, err := set.PlatformClient.CreateCluster(ctx, client.NewClusterRequest{
		Name:           options.Name,
		SSOProvisioner: ssoProvisioner,
		NodePools:      []client.NodePool{nodePool},
		Version:        "",
		Environment:    "",
		PullSecretRef:  pullSecretRef,
	})
	if err != nil {
		return redact.Errorf("could not create cluster: %w", redact.Safe(err))
	}

	ux.Fsuccessf(cmd.OutOrStdout(), "created cluster: %s\n", cluster.Name)

	return nil
}

func buildNodePool(options CreateOptions) client.NodePool {
	if options.EnableAutoscaling {
		return client.NodePool{
			ID:                 "",
			Name:               "",
			Preset:             options.Preset,
			Replicas:           nil,
			Compute:            nil,
			AutoscalingEnabled: true,
			MinCount:           &options.MinNodes,
			MaxCount:           &options.MaxNodes,
		}
	}

	replicas := options.NodeCount

	return client.NodePool{
		ID:                 "",
		Name:               "",
		Preset:             options.Preset,
		Replicas:           &replicas,
		Compute:            nil,
		AutoscalingEnabled: false,
		MinCount:           nil,
		MaxCount:           nil,
	}
}

var errPullSecretNotFound = redact.Errorf("pull secret not found")

func resolvePullSecretRef(name string, pullSecrets []client.PullSecret) (*string, error) {
	if name == "" {
		return nil, nil //nolint:nilnil // nil,nil is intentional: no name means no pull secret ref
	}

	for i := range pullSecrets {
		if pullSecrets[i].Name == name {
			return &pullSecrets[i].ID, nil
		}
	}

	return nil, errPullSecretNotFound
}

func optionsFromWizard(pullSecrets []client.PullSecret) (CreateOptions, error) {
	var options CreateOptions

	wz := wizard.NewWizard(getClusterWizardInputs(pullSecrets))

	result, err := wz.Run()
	if err != nil {
		return options, redact.Errorf("could not gather information: %w", redact.Safe(err))
	}

	if result.Cancelled() {
		return options, errCancelledByUser
	}

	options.Name = result.MustGetValue("name")
	options.Preset = result.MustGetValue("preset")
	options.EnableAutoscaling = result.MustGetValue("autoscaling") == answerYes

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

	if len(pullSecrets) > 0 {
		selected := result.MustGetValue("pullSecret")
		if selected != noPullSecret {
			options.PullSecret = selected
		}
	}

	return options, nil
}

//nolint:funlen // wizard input definitions are declarative
func getClusterWizardInputs(pullSecrets []client.PullSecret) []wizard.Input {
	inputs := []wizard.Input{
		{
			ID:          "name",
			Placeholder: "Cluster Name",
			Type:        wizard.InputTypeText,
			Limit:       0,
			Validator:   nil,
			Options:     nil,
			DependsOn:   "",
			ShowWhen:    nil,
		},
		{
			ID:          "preset",
			Placeholder: "Node Preset",
			Type:        wizard.InputTypeSelect,
			Limit:       0,
			Validator:   nil,
			Options:     []string{"minimal", "balanced", "performance"},
			DependsOn:   "",
			ShowWhen:    nil,
		},
		{
			ID:          "autoscaling",
			Placeholder: "Enable autoscaling",
			Type:        wizard.InputTypeToggle,
			Limit:       0,
			Validator:   nil,
			Options:     []string{answerNo, answerYes},
			DependsOn:   "",
			ShowWhen:    nil,
		},
		{
			ID:          "nodes",
			Placeholder: "Node Count (" + strconv.Itoa(minCount) + "-" + strconv.Itoa(maxCount) + ")",
			Type:        wizard.InputTypeText,
			Limit:       0,
			Validator:   nil,
			Options:     nil,
			DependsOn:   "",
			ShowWhen: func(answers map[string]wizard.Answer) bool {
				// Show this field only when autoscaling is disabled
				return answers["autoscaling"].Value == answerNo
			},
		},
		{
			ID:          "minNodes",
			Placeholder: "Minimum Nodes (" + strconv.Itoa(minCount) + "-" + strconv.Itoa(maxCount) + ")",
			Type:        wizard.InputTypeText,
			Limit:       0,
			Validator:   nil,
			Options:     nil,
			DependsOn:   "",
			ShowWhen: func(answers map[string]wizard.Answer) bool {
				// Show this field only when autoscaling is enabled
				return answers["autoscaling"].Value == answerYes
			},
		},
		{
			ID:          "maxNodes",
			Placeholder: "Maximum Nodes (" + strconv.Itoa(minCount) + "-" + strconv.Itoa(maxCount) + ")",
			Type:        wizard.InputTypeText,
			Limit:       0,
			Validator:   nil,
			Options:     nil,
			DependsOn:   "",
			ShowWhen: func(answers map[string]wizard.Answer) bool {
				// Show this field only when autoscaling is enabled
				return answers["autoscaling"].Value == answerYes
			},
		},
	}

	if len(pullSecrets) > 0 {
		psOptions := make([]string, 0, 1+len(pullSecrets))
		psOptions = append(psOptions, noPullSecret)

		for _, ps := range pullSecrets {
			psOptions = append(psOptions, ps.Name)
		}

		inputs = append(inputs, wizard.Input{
			ID:          "pullSecret",
			Placeholder: "Image Pull Secret",
			Type:        wizard.InputTypeSelect,
			Limit:       0,
			Validator:   nil,
			Options:     psOptions,
			DependsOn:   "",
			ShowWhen:    nil,
		})
	}

	return inputs
}

//nolint:cyclop // validation logic is inherently sequential
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
	} else if options.NodeCount < minCount || options.NodeCount > maxCount {
		// Validate fixed node count
		return errInvalidNodeCount
	}

	return nil
}

func selectSSOProvisioner(
	ctx context.Context,
	platformClient client.Client,
	out interface {
		Write(p []byte) (n int, err error)
	},
) (string, error) {
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
		ux.Fprintf(out, "Using SSO provisioner: %s\n", provisioners[0].Name)
		return provisioners[0].ID, nil
	default:
		// For now, auto-select the first one. Future: add wizard step
		ux.Fprintf(out, "Using SSO provisioner: %s\n", provisioners[0].Name)
		return provisioners[0].ID, nil
	}
}
