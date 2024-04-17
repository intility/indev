package authenticator

import (
	"context"
	"fmt"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"

	"github.com/intility/minctl/pkg/tokencache"
)

type AuthConfig struct {
	ClientID  string
	Authority string
	Scopes    []string
}

func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		ClientID:  "b65cf9b0-290c-4b44-a4b1-0b02b7752b3c",
		Authority: "https://login.microsoftonline.com/intility.no",
		Scopes:    []string{"api://containerplatform.intility.com/user_impersonation"},
	}
}

type Authenticator struct {
	clientID  string
	authority string
	scopes    []string
	cache     cache.ExportReplace
}

type Option func(authenticator *Authenticator)

func NewAuthenticator(config AuthConfig, options ...Option) *Authenticator {
	authenticator := &Authenticator{
		clientID:  config.ClientID,
		authority: config.Authority,
		scopes:    config.Scopes,
		cache:     tokencache.New(),
	}

	for _, opt := range options {
		opt(authenticator)
	}

	return authenticator
}

func WithTokenCache(cache cache.ExportReplace) Option {
	return func(auth *Authenticator) {
		auth.cache = cache
	}
}

func (a *Authenticator) Authenticate(ctx context.Context) (public.AuthResult, error) {
	// confidential clients have a credential, such as a secret or a certificate
	var result public.AuthResult

	publicClient, err := public.New(
		a.clientID,
		public.WithAuthority(a.authority),
		public.WithCache(a.cache),
	)
	if err != nil {
		return result, fmt.Errorf("could not create public client: %w", err)
	}

	accounts, err := publicClient.Accounts(ctx)
	if len(accounts) > 0 {
		result, err = publicClient.AcquireTokenSilent(
			ctx,
			a.scopes,
			public.WithSilentAccount(accounts[0]),
		)
	}

	if err != nil || len(accounts) == 0 {
		result, err = publicClient.AcquireTokenInteractive(
			ctx,
			a.scopes,
			public.WithRedirectURI("http://localhost:42069"),
		)
		if err != nil {
			return result, fmt.Errorf("could not acquire token: %w", err)
		}
	}

	return result, nil
}

func (a *Authenticator) IsAuthenticated(ctx context.Context) (bool, error) {
	publicClient, err := public.New(
		a.clientID,
		public.WithAuthority(a.authority),
		public.WithCache(a.cache),
	)
	if err != nil {
		return false, fmt.Errorf("could not create public client: %w", err)
	}

	accounts, err := publicClient.Accounts(ctx)
	if err != nil {
		return false, fmt.Errorf("could not get accounts: %w", err)
	}

	return len(accounts) > 0, nil
}
