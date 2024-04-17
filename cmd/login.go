package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/intility/minctl/pkg/authenticator"
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
		auth := authenticator.NewAuthenticator(authenticator.AuthConfig{
			ClientID:  "b65cf9b0-290c-4b44-a4b1-0b02b7752b3c",
			Authority: "https://login.microsoftonline.com/intility.no",
			Scopes:    []string{"api://containerplatform.intility.com/user_impersonation"},
		})

		ctx, cancel := context.WithTimeout(cmd.Context(), authTimeout)
		defer cancel()

		result, err := auth.Authenticate(ctx)
		if err != nil {
			return fmt.Errorf("could not authenticate: %w", err)
		}

		cmd.Println("Access token: ", result.AccessToken)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
