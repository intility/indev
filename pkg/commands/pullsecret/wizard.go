package pullsecret

import (
	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/wizard"
)

type registryInput struct {
	address  string
	username string
	password string
}

func registryWizardInputs() []wizard.Input {
	return []wizard.Input{
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
	}
}

func promptRegistryCredential() (*registryInput, error) {
	wz := wizard.NewWizard(registryWizardInputs())

	result, err := wz.Run()
	if err != nil {
		return nil, redact.Errorf("could not gather registry information: %w", redact.Safe(err))
	}

	if result.Cancelled() {
		return nil, errCancelledByUser
	}

	return &registryInput{
		address:  result.MustGetValue("address"),
		username: result.MustGetValue("username"),
		password: result.MustGetValue("password"),
	}, nil
}

func promptAddMore(placeholder string) (bool, error) {
	wz := wizard.NewWizard([]wizard.Input{
		{
			ID:          "addMore",
			Placeholder: placeholder,
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
		return false, redact.Errorf("could not gather information: %w", redact.Safe(err))
	}

	if result.Cancelled() {
		return false, errCancelledByUser
	}

	return result.MustGetValue("addMore") == answerYes, nil
}
