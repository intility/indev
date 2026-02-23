package pullsecret

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/internal/wizard"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

var (
	errInvalidAddRegistryFormat = redact.Errorf("--add-registry format must be address,username,password")
	errNoChanges                = redact.Errorf("no changes specified")
)

const addRegistryParts = 3

func NewEditCommand(set clientset.ClientSet) *cobra.Command {
	var (
		name             string
		addRegistries    []string
		removeRegistries []string
	)

	cmd := &cobra.Command{
		Use:     "edit [name]",
		Short:   "Edit an existing image pull secret",
		Long:    `Edit an existing image pull secret by adding or removing registries.`,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "pullsecret.edit")
			defer span.End()

			if len(args) > 0 {
				name = args[0]
			}

			if name == "" {
				return redact.Errorf("pull secret name cannot be empty")
			}

			cmd.SilenceUsage = true

			ps, err := FindPullSecretByName(ctx, set.PlatformClient, name)
			if err != nil {
				return err
			}

			registries := make(map[string]*client.PullSecretCredential)

			hasFlags := len(addRegistries) > 0 || len(removeRegistries) > 0

			if hasFlags {
				if err = applyFlagEdits(registries, addRegistries, removeRegistries); err != nil {
					return err
				}
			} else {
				var wizErr error

				registries, wizErr = editFromWizard(ps)
				if wizErr != nil {
					return wizErr
				}
			}

			if len(registries) == 0 {
				return errNoChanges
			}

			result, err := set.PlatformClient.EditPullSecret(ctx, ps.ID, client.EditPullSecretRequest{
				Registries: registries,
			})
			if err != nil {
				return redact.Errorf("could not edit pull secret: %w", redact.Safe(err))
			}

			ux.Fsuccessf(cmd.OutOrStdout(), "updated pull secret: %s\n", result.Name)

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the pull secret")
	cmd.Flags().StringArrayVar(&addRegistries, "add-registry", nil, "Add registry (format: address,username,password)")
	cmd.Flags().StringArrayVar(&removeRegistries, "remove-registry", nil, "Remove registry by address")

	return cmd
}

func applyFlagEdits(
	registries map[string]*client.PullSecretCredential,
	addRegistries []string,
	removeRegistries []string,
) error {
	for _, add := range addRegistries {
		parts := strings.SplitN(add, ",", addRegistryParts)
		if len(parts) != addRegistryParts {
			return errInvalidAddRegistryFormat
		}

		registries[parts[0]] = &client.PullSecretCredential{
			Username: parts[1],
			Password: parts[2],
		}
	}

	for _, remove := range removeRegistries {
		registries[remove] = nil
	}

	return nil
}

//nolint:cyclop // wizard loop is inherently sequential
func editFromWizard(ps *client.PullSecret) (map[string]*client.PullSecretCredential, error) {
	registries := make(map[string]*client.PullSecretCredential)

	if len(ps.Registries) > 0 {
		for _, registry := range ps.Registries {
			removeWz := wizard.NewWizard([]wizard.Input{
				{
					ID:          "remove",
					Placeholder: "Remove " + registry + "?",
					Type:        wizard.InputTypeToggle,
					Limit:       0,
					Validator:   nil,
					Options:     []string{"no", answerYes},
					DependsOn:   "",
					ShowWhen:    nil,
				},
			})

			removeResult, err := removeWz.Run()
			if err != nil {
				return nil, redact.Errorf("could not gather information: %w", redact.Safe(err))
			}

			if removeResult.Cancelled() {
				return nil, errCancelledByUser
			}

			if removeResult.MustGetValue("remove") == answerYes {
				registries[registry] = nil
			}
		}
	}

	for {
		addWz := wizard.NewWizard([]wizard.Input{
			{
				ID:          "addMore",
				Placeholder: "Add a new registry?",
				Type:        wizard.InputTypeToggle,
				Limit:       0,
				Validator:   nil,
				Options:     []string{"no", answerYes},
				DependsOn:   "",
				ShowWhen:    nil,
			},
		})

		addResult, err := addWz.Run()
		if err != nil {
			return nil, redact.Errorf("could not gather information: %w", redact.Safe(err))
		}

		if addResult.Cancelled() {
			return nil, errCancelledByUser
		}

		if addResult.MustGetValue("addMore") != answerYes {
			break
		}

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
			return nil, redact.Errorf("could not gather registry information: %w", redact.Safe(err))
		}

		if regResult.Cancelled() {
			return nil, errCancelledByUser
		}

		address := regResult.MustGetValue("address")
		username := regResult.MustGetValue("username")
		password := regResult.MustGetValue("password")

		registries[address] = &client.PullSecretCredential{
			Username: username,
			Password: password,
		}
	}

	return registries, nil
}
