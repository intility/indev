package pullsecret

import (
	"context"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

var errPullSecretNotFound = redact.Errorf("pull secret not found")

func NewGetCommand(set clientset.ClientSet) *cobra.Command {
	var (
		name         string
		errEmptyName = redact.Errorf("pull secret name cannot be empty")
	)

	cmd := &cobra.Command{
		Use:     "get [name]",
		Short:   "Get detailed information about a pull secret",
		Long:    `Display comprehensive information about a specific image pull secret`,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "pullsecret.get")
			defer span.End()

			cmd.SilenceUsage = true

			if len(args) > 0 {
				name = args[0]
			}

			if name == "" {
				return errEmptyName
			}

			ps, err := FindPullSecretByName(ctx, set.PlatformClient, name)
			if err != nil {
				return err
			}

			printPullSecretDetails(cmd.OutOrStdout(), ps)

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the pull secret")

	return cmd
}

// FindPullSecretByName lists all pull secrets and returns the one matching the given name.
func FindPullSecretByName(ctx context.Context, c client.Client, name string) (*client.PullSecret, error) {
	pullSecrets, err := c.ListPullSecrets(ctx)
	if err != nil {
		return nil, redact.Errorf("could not list pull secrets: %w", redact.Safe(err))
	}

	for i := range pullSecrets {
		if pullSecrets[i].Name == name {
			return &pullSecrets[i], nil
		}
	}

	return nil, errPullSecretNotFound
}

func printPullSecretDetails(writer io.Writer, ps *client.PullSecret) {
	ux.Fprintf(writer, "Pull Secret Information:\n")
	ux.Fprintf(writer, "  Name:       %s\n", ps.Name)
	ux.Fprintf(writer, "  ID:         %s\n", ps.ID)
	ux.Fprintf(writer, "  Created At: %s\n", ps.CreatedAt)

	if len(ps.Registries) > 0 {
		ux.Fprintf(writer, "Registries:\n")

		for _, registry := range ps.Registries {
			ux.Fprintf(writer, "  - %s\n", registry)
		}
	}
}

// FormatRegistries joins registry addresses with commas for display.
func FormatRegistries(registries []string) string {
	return strings.Join(registries, ", ")
}
