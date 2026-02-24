package pullsecret

import (
	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/internal/wizard"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

var (
	errEmptyPullSecretName = redact.Errorf("pull secret name cannot be empty")
	errNoRegistries        = redact.Errorf("at least one registry is required")
	errCancelledByUser     = redact.Errorf("cancelled by user")
)

func NewCreateCommand(set clientset.ClientSet) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a new image pull secret",
		Long:    `Create a new image pull secret with registry credentials.`,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "pullsecret.create")
			defer span.End()

			registries := make(map[string]client.PullSecretCredential)

			if name == "" {
				var err error

				name, registries, err = createFromWizard()
				if err != nil {
					return err
				}
			}

			if name == "" {
				return errEmptyPullSecretName
			}

			if len(registries) == 0 {
				return errNoRegistries
			}

			cmd.SilenceUsage = true

			ps, err := set.PlatformClient.CreatePullSecret(ctx, client.NewPullSecretRequest{
				Name:       name,
				Registries: registries,
			})
			if err != nil {
				return redact.Errorf("could not create pull secret: %w", redact.Safe(err))
			}

			ux.Fsuccessf(cmd.OutOrStdout(), "created pull secret: %s\n", ps.Name)

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the pull secret to create")

	return cmd
}

const answerYes = "yes"

func createFromWizard() (string, map[string]client.PullSecretCredential, error) {
	nameWz := wizard.NewWizard([]wizard.Input{
		{
			ID:          "name",
			Placeholder: "Pull Secret Name",
			Type:        wizard.InputTypeText,
			Limit:       0,
			Validator:   nil,
			Options:     nil,
			DependsOn:   "",
			ShowWhen:    nil,
		},
	})

	nameResult, err := nameWz.Run()
	if err != nil {
		return "", nil, redact.Errorf("could not gather information: %w", redact.Safe(err))
	}

	if nameResult.Cancelled() {
		return "", nil, errCancelledByUser
	}

	name := nameResult.MustGetValue("name")

	registries, err := collectRegistries()
	if err != nil {
		return "", nil, err
	}

	return name, registries, nil
}

func collectRegistries() (map[string]client.PullSecretCredential, error) {
	registries := make(map[string]client.PullSecretCredential)

	for {
		cred, err := promptRegistryCredential()
		if err != nil {
			return nil, err
		}

		registries[cred.address] = client.PullSecretCredential{
			Username: cred.username,
			Password: cred.password,
		}

		more, err := promptAddMore("Add another registry?")
		if err != nil {
			return nil, err
		}

		if !more {
			break
		}
	}

	return registries, nil
}
