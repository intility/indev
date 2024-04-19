package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/intility/minctl/pkg/authenticator"
	"github.com/intility/minctl/pkg/config"
)

const (
	authTimeout = 5 * time.Minute
)

// loginCmd represents the login command.
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Intility Container Platform",
	Long:  `Login to Intility Container Platform using your Intility credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		auth := authenticator.NewAuthenticator(authenticator.Config{
			ClientID:  config.ClientID,
			Authority: config.Authority,
			Scopes: []string{
				config.ScopePlatform,
			},
		})

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
	rootCmd.AddCommand(loginCmd)
}
