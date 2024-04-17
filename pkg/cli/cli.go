package cli

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/intility/minctl/pkg/authenticator"
)

const (
	timeout = 3 * time.Second
)

var errNotAuthenticated = errors.New("not authenticated")

func CreateAuthGate(message string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		auth := authenticator.NewAuthenticator(authenticator.AuthConfig{
			ClientID:  "b65cf9b0-290c-4b44-a4b1-0b02b7752b3c",
			Authority: "https://login.microsoftonline.com/intility.no",
			Scopes:    []string{"api://containerplatform.intility.com/user_impersonation"},
		})

		ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
		defer cancel()

		authenticated, err := auth.IsAuthenticated(ctx)
		if err != nil {
			return fmt.Errorf("failed to determine authentication status: %w", err)
		}

		if !authenticated {
			cmd.SilenceUsage = true
			return fmt.Errorf("%w: %s", errNotAuthenticated, message)
		}

		return nil
	}
}
