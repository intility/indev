package cli

import (
	"context"
	"errors"
	"fmt"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/intility/minctl/pkg/authenticator"
	"github.com/intility/minctl/pkg/config"
)

const (
	timeout = 3 * time.Second
)

var errNotAuthenticated = errors.New("not authenticated")

func CreateAuthGate(message string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		auth := authenticator.NewAuthenticator(authenticator.Config{
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

func CreatePasswordPrompter(cmd *cobra.Command) func(string) (string, error) {
	return func(prompt string) (string, error) {
		cmd.Print(prompt + ": ")

		bytePassword, err := term.ReadPassword(syscall.Stdin)

		cmd.Println()

		if err != nil {
			return "", fmt.Errorf("could not read password: %w", err)
		}

		return string(bytePassword), nil
	}
}
