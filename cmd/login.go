package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/intility/icpctl/internal/build"
	"github.com/intility/icpctl/internal/cli"
	"github.com/intility/icpctl/internal/redact"
	"github.com/intility/icpctl/internal/ux"
	"github.com/intility/icpctl/pkg/authenticator"
)

const (
	authTimeout = 5 * time.Minute
)

var useDeviceCodeFlow bool

// loginCmd represents the login command.
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Sign in to Intility Container Platform",
	Long:  `Sign in to Intility Container Platform using your Intility credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := authenticator.Config{
			ClientID:    build.ClientID(),
			Authority:   build.Authority(),
			Scopes:      build.Scopes(),
			RedirectURI: build.SuccessRedirect(),
		}

		var options []authenticator.Option
		if useDeviceCodeFlow {
			options = append(options, authenticator.WithDeviceCodeFlow(cli.CreatePrinter(cmd)))
		}

		auth := authenticator.NewAuthenticator(cfg, options...)

		ctx, cancel := context.WithTimeout(cmd.Context(), authTimeout)
		defer cancel()

		result, err := auth.Authenticate(ctx)
		if err != nil {
			return redact.Errorf("could not authenticate: %w", err)
		}

		ux.Fsuccess(cmd.OutOrStdout(), "authenticated as "+result.Account.PreferredUsername)

		return nil
	},
}

func init() {
	loginCmd.Flags().BoolVar(&useDeviceCodeFlow, "device", false, "Use device code flow for authentication")
	rootCmd.AddCommand(loginCmd)
}
