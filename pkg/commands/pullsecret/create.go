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

//nolint:cyclop // wizard loop is inherently sequential
func createFromWizard() (string, map[string]client.PullSecretCredential, error) {
	registries := make(map[string]client.PullSecretCredential)

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

	for {
		regWz := wizard.NewWizard([]wizard.Input{
			{
				ID:          "address",
				Placeholder: "Registry Address (e.g. ghcr.io)",
				Type:        wizard.InputTypeText,
				Limit:       0,
				Validator:   nil,
				Options:     nil,
				DependsOn:   "",
				ShowWhen:    nil,
			},
			{
				ID:          "username",
				Placeholder: "Username",
				Type:        wizard.InputTypeText,
				Limit:       0,
				Validator:   nil,
				Options:     nil,
				DependsOn:   "",
				ShowWhen:    nil,
			},
			{
				ID:          "password",
				Placeholder: "Password",
				Type:        wizard.InputTypePassword,
				Limit:       0,
				Validator:   nil,
				Options:     nil,
				DependsOn:   "",
				ShowWhen:    nil,
			},
		})

		regResult, err := regWz.Run()
		if err != nil {
			return "", nil, redact.Errorf("could not gather registry information: %w", redact.Safe(err))
		}

		if regResult.Cancelled() {
			return "", nil, errCancelledByUser
		}

		address := regResult.MustGetValue("address")
		username := regResult.MustGetValue("username")
		password := regResult.MustGetValue("password")

		registries[address] = client.PullSecretCredential{
			Username: username,
			Password: password,
		}

		addMoreWz := wizard.NewWizard([]wizard.Input{
			{
				ID:          "addMore",
				Placeholder: "Add another registry?",
				Type:        wizard.InputTypeToggle,
				Limit:       0,
				Validator:   nil,
				Options:     []string{"no", answerYes},
				DependsOn:   "",
				ShowWhen:    nil,
			},
		})

		addMoreResult, err := addMoreWz.Run()
		if err != nil {
			return "", nil, redact.Errorf("could not gather information: %w", redact.Safe(err))
		}

		if addMoreResult.Cancelled() {
			return "", nil, errCancelledByUser
		}

		if addMoreResult.MustGetValue("addMore") != answerYes {
			break
		}
	}

	return name, registries, nil
}
