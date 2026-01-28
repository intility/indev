package account

import (
	"errors"
	"slices"

	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/authenticator"
	"github.com/intility/indev/pkg/clientset"
)

func NewShowCommand(set clientset.ClientSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show account information",
		Long:  `Show account information`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, span := telemetry.StartSpan(cmd.Context(), "account.show")
			defer span.End()

			account, err := set.Authenticator.GetCurrentAccount(cmd.Context())
			if err != nil {
				if errors.Is(err, authenticator.ErrNoAccounts) {
					ux.Fprint(cmd.OutOrStdout(), "You are not signed in to any accounts\n")
					ux.Fprint(cmd.OutOrStdout(), "Use `%s %s` to sign in\n", cmd.Root().Name(), "login")

					return nil
				}

				return redact.Errorf("could not get account information: %w", err)
			}

			me, err := set.PlatformClient.GetMe(cmd.Context())
			if err != nil {
				return redact.Errorf("could not get user information from Me endpoint: %w", redact.Safe(err))
			}

			role := getOrganizationalRole(me.OrganizationRoles)

			ux.Fprint(cmd.OutOrStdout(), "Account information\n")
			ux.Fprint(cmd.OutOrStdout(), "Username: %s\n", account.PreferredUsername)
			ux.Fprint(cmd.OutOrStdout(), "Realm: %s\n", account.Realm)
			ux.Fprint(cmd.OutOrStdout(), "Organization: %s\n", me.OrganizationName)
			ux.Fprint(cmd.OutOrStdout(), "Organization Role: %s\n", role)

			return nil
		},
	}

	return cmd
}

func getOrganizationalRole(roles []string) string {
	if slices.Contains(roles, "owner") {
		return "Admin"
	}

	if slices.Contains(roles, "member") {
		return "Member"
	}

	return "Unknown"
}
