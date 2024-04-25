package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/intility/minctl/internal/redact"
	"github.com/intility/minctl/internal/ux"
	"github.com/intility/minctl/pkg/authenticator"
	"github.com/intility/minctl/pkg/config"
)

const (
	timeout = 3 * time.Second
)

var errNotAuthenticated = errors.New("not authenticated")

func CreateAuthGate(message any) func(cmd *cobra.Command, args []string) error {
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
			return redact.Errorf("failed to determine authentication status: %w", redact.Safe(err))
		}

		if !authenticated {
			cmd.SilenceUsage = true
			return redact.Errorf("%w: %v", redact.Safe(errNotAuthenticated), message)
		}

		return nil
	}
}

func CreatePasswordPrompter(cmd *cobra.Command) func(string) (string, error) {
	return func(prompt string) (string, error) {
		ux.Fprint(cmd.OutOrStdout(), prompt+": ")

		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))

		ux.Fprint(cmd.OutOrStdout(), "\n")

		if err != nil {
			return "", fmt.Errorf("could not read password: %w", err)
		}

		return string(bytePassword), nil
	}
}

func CreatePrinter(cmd *cobra.Command) func(ctx context.Context, message string) error {
	return func(ctx context.Context, message string) error {
		ux.Fprint(cmd.OutOrStdout(), message+"\n")
		return nil
	}
}
