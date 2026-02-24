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
			_, span := telemetry.StartSpan(cmd.Context(), "pullsecret.edit")
			defer span.End()

			if len(args) > 0 {
				name = args[0]
			}

			if name == "" {
				return redact.Errorf("pull secret name cannot be empty")
			}

			cmd.SilenceUsage = true

			return runEdit(cmd, set, editOptions{
				name:             name,
				addRegistries:    addRegistries,
				removeRegistries: removeRegistries,
			})
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the pull secret")
	cmd.Flags().StringArrayVar(&addRegistries, "add-registry", nil, "Add registry (format: address,username,password)")
	cmd.Flags().StringArrayVar(&removeRegistries, "remove-registry", nil, "Remove registry by address")

	return cmd
}

type editOptions struct {
	name             string
	addRegistries    []string
	removeRegistries []string
}

func runEdit(cmd *cobra.Command, set clientset.ClientSet, opts editOptions) error {
	ctx := cmd.Context()

	ps, err := FindPullSecretByName(ctx, set.PlatformClient, opts.name)
	if err != nil {
		return err
	}

	registries := make(map[string]*client.PullSecretCredential)

	hasFlags := len(opts.addRegistries) > 0 || len(opts.removeRegistries) > 0

	if hasFlags {
		if err = applyFlagEdits(registries, opts.addRegistries, opts.removeRegistries); err != nil {
			return err
		}
	} else {
		registries, err = editFromWizard(ps)
		if err != nil {
			return err
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

func editFromWizard(ps *client.PullSecret) (map[string]*client.PullSecretCredential, error) {
	registries := make(map[string]*client.PullSecretCredential)

	if err := promptRemoveRegistries(registries, ps.Registries); err != nil {
		return nil, err
	}

	if err := promptAddRegistries(registries); err != nil {
		return nil, err
	}

	return registries, nil
}

func promptRemoveRegistries(registries map[string]*client.PullSecretCredential, current []string) error {
	for _, registry := range current {
		wz := wizard.NewWizard([]wizard.Input{
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

		result, err := wz.Run()
		if err != nil {
			return redact.Errorf("could not gather information: %w", redact.Safe(err))
		}

		if result.Cancelled() {
			return errCancelledByUser
		}

		if result.MustGetValue("remove") == answerYes {
			registries[registry] = nil
		}
	}

	return nil
}

func promptAddRegistries(registries map[string]*client.PullSecretCredential) error {
	for {
		more, err := promptAddMore("Add a new registry?")
		if err != nil {
			return err
		}

		if !more {
			break
		}

		cred, err := promptRegistryCredential()
		if err != nil {
			return err
		}

		registries[cred.address] = &client.PullSecretCredential{
			Username: cred.username,
			Password: cred.password,
		}
	}

	return nil
}
