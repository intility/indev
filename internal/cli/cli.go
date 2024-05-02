package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/term"

	"github.com/intility/minctl/internal/build"
	"github.com/intility/minctl/internal/redact"
	"github.com/intility/minctl/internal/telemetry"
	"github.com/intility/minctl/internal/ux"
	"github.com/intility/minctl/pkg/authenticator"
)

const (
	timeout = 3 * time.Second
)

var errNotAuthenticated = errors.New("not authenticated")

func CreateAuthGate(message any) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var (
			span trace.Span
			ctx  = cmd.Context()
		)

		ctx, span = telemetry.StartSpan(ctx, "authenticator.AuthGate")
		defer span.End()

		auth := authenticator.NewAuthenticator(authenticator.Config{
			ClientID:    build.ClientID(),
			Authority:   build.Authority(),
			Scopes:      build.Scopes(),
			RedirectURI: build.SuccessRedirect(),
		})

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		authenticated, err := auth.IsAuthenticated(ctx)
		if err != nil {
			reportErr := redact.Errorf("auth gate failed to determine authentication status: %w", redact.Safe(err))

			if span != nil {
				span.SetStatus(codes.Error, "authentication gate failed")
				span.RecordError(reportErr, trace.WithStackTrace(true))
			}

			cmd.SilenceUsage = true

			return redact.Errorf("%w: %v", redact.Safe(errNotAuthenticated), message)
		}

		if !authenticated {
			cmd.SilenceUsage = true
			return redact.Errorf("%w: %v", redact.Safe(errNotAuthenticated), message)
		}

		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
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
