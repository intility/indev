package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/intility/minctl/pkg/authenticator"
	"github.com/intility/minctl/pkg/cli"
	"github.com/intility/minctl/pkg/config"
)

const (
	authTimeout = 5 * time.Minute
)

var useDeviceCodeFlow bool

// loginCmd represents the login command.
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Intility Container Platform",
	Long:  `Login to Intility Container Platform using your Intility credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := authenticator.Config{
			ClientID:  config.ClientID,
			Authority: config.Authority,
			Scopes: []string{
				config.ScopePlatform,
			},
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
			return fmt.Errorf("could not authenticate: %w", err)
		}

		cmd.Println(styleSuccess.Render("success: ") + "authenticated as " + result.Account.PreferredUsername)

		return nil
	},
}

func init() {
	loginCmd.Flags().BoolVar(&useDeviceCodeFlow, "device", false, "Use device code flow for authentication")
	rootCmd.AddCommand(loginCmd)
}
