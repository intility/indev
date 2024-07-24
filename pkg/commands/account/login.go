package account

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"

	"github.com/intility/icpctl/internal/build"
	"github.com/intility/icpctl/internal/cli"
	"github.com/intility/icpctl/internal/redact"
	"github.com/intility/icpctl/internal/telemetry"
	"github.com/intility/icpctl/internal/ux"
	"github.com/intility/icpctl/pkg/authenticator"
	"github.com/intility/icpctl/pkg/clientset"
)

const (
	authTimeout = 5 * time.Minute
)

func NewLoginCommand(set clientset.ClientSet) *cobra.Command {
	var useDeviceCodeFlow bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Sign in to Intility Container Platform",
		Long:  `Sign in to Intility Container Platform using your Intility credentials.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "account.login")
			defer span.End()

			span.SetAttributes(attribute.Bool("device_code_flow", useDeviceCodeFlow))
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

			ctx, cancel := context.WithTimeout(ctx, authTimeout)
			defer cancel()

			result, err := auth.Authenticate(ctx)
			if err != nil {
				return redact.Errorf("could not authenticate: %w", err)
			}

			ux.Fsuccess(cmd.OutOrStdout(), "authenticated as "+result.Account.PreferredUsername)

			return nil
		},
	}

	cmd.Flags().BoolVar(&useDeviceCodeFlow, "device", false, "Use device code flow for authentication")

	return cmd
}
