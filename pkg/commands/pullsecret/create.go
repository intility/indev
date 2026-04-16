package pullsecret

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

var (
	errEmptyPullSecretName = redact.Errorf("pull secret name cannot be empty")
	errNoRegistries        = redact.Errorf("at least one --registry must be provided")
	errInvalidRegistry     = redact.Errorf("invalid --registry format, expected address:username:password")
	errEmptyAddress        = redact.Errorf("registry address cannot be empty")
	errEmptyUsername       = redact.Errorf("registry username cannot be empty")
	errEmptyPassword       = redact.Errorf("registry password cannot be empty")
	errDuplicateAddress    = redact.Errorf("duplicate registry address")
)

// RegistryEntry holds credentials for a single registry.
type RegistryEntry struct {
	Address  string
	Username string
	Password string //nolint:gosec // G117: this is a credential payload, not a hardcoded secret
}

// CreateOptions holds the options for the pull secret create command.
type CreateOptions struct {
	Name       string
	Registries []RegistryEntry
}

func NewCreateCommand(set clientset.ClientSet) *cobra.Command {
	var (
		name          string
		registryFlags []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new image pull secret",
		Long: `Create a new image pull secret with one or more registries.

Use --registry once per registry in the format address:username:password.
Passwords may contain colons.

To add registries with explicit ports (e.g. myregistry.io:5000), create the
pull secret first and then use 'indev pull-secret registry add'.`,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "pullsecret.create")
			defer span.End()

			registries := make([]RegistryEntry, 0, len(registryFlags))
			for _, flag := range registryFlags {
				entry, err := parseRegistryFlag(flag)
				if err != nil {
					return err
				}

				registries = append(registries, entry)
			}

			options := CreateOptions{Name: name, Registries: registries}

			if err := validateCreateOptions(options); err != nil {
				return err
			}

			cmd.SilenceUsage = true

			regMap := make(map[string]client.PullSecretCredential, len(options.Registries))
			for _, r := range options.Registries {
				regMap[r.Address] = client.PullSecretCredential{
					Username: r.Username,
					Password: r.Password,
				}
			}

			ps, err := set.PlatformClient.CreatePullSecret(ctx, client.NewPullSecretRequest{
				Name:       options.Name,
				Registries: regMap,
			})
			if err != nil {
				return redact.Errorf("could not create pull secret: %w", redact.Safe(err))
			}

			ux.Fsuccessf(cmd.OutOrStdout(), "created pull secret: %s\n", ps.Name)

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the pull secret to create")
	cmd.Flags().StringArrayVar(&registryFlags, "registry", nil,
		"Registry credentials in the format address:username:password (repeat for multiple registries)")

	return cmd
}

// parseRegistryFlag parses a registry flag value in the format address:username:password.
// Passwords may contain colons — only the first two colons are used as separators.
func parseRegistryFlag(value string) (RegistryEntry, error) {
	parts := strings.SplitN(value, ":", 3)
	if len(parts) != 3 {
		return RegistryEntry{}, errInvalidRegistry
	}

	return RegistryEntry{Address: parts[0], Username: parts[1], Password: parts[2]}, nil
}

func validateCreateOptions(options CreateOptions) error {
	if options.Name == "" {
		return errEmptyPullSecretName
	}

	if len(options.Registries) == 0 {
		return errNoRegistries
	}

	seen := make(map[string]struct{}, len(options.Registries))

	for _, r := range options.Registries {
		if r.Address == "" {
			return errEmptyAddress
		}

		if r.Username == "" {
			return errEmptyUsername
		}

		if r.Password == "" {
			return errEmptyPassword
		}

		if _, dup := seen[r.Address]; dup {
			return errDuplicateAddress
		}

		seen[r.Address] = struct{}{}
	}

	return nil
}
