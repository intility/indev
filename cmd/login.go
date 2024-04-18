package cmd

import (
	"context"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"time"

	"github.com/intility/minctl/pkg/config"

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

		style := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		cmd.Println(style.Render("success: ") + "authenticated as " + result.Account.PreferredUsername)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
