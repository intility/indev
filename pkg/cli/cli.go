package cli

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/intility/minctl/pkg/authenticator"
	"github.com/intility/minctl/pkg/config"
)

const (
	timeout = 3 * time.Second
)

var errNotAuthenticated = errors.New("not authenticated")

func CreateAuthGate(message string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		auth := authenticator.NewAuthenticator(authenticator.AuthConfig{
			ClientID:  config.ClientID,
			Authority: config.Authority,
			Scopes:    []string{config.ScopePlatform},
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
