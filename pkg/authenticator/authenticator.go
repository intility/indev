package authenticator

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"

	"github.com/intility/minctl/internal/build"
	"github.com/intility/minctl/internal/redact"
	"github.com/intility/minctl/pkg/tokencache"
)

var (
	errNoPrinter        = errors.New("printer is required for device code flow")
	errFlowNotSupported = errors.New("unsupported authentication flow")
	errNoAccounts       = errors.New("no accounts found")
	errTokenExpired     = errors.New("token has expired")
	errDeclinedScopes   = errors.New("scopes have been declined")
)

type Config struct {
	ClientID    string
	Authority   string
	Scopes      []string
	RedirectURI string
}

func ConfigFromBuildProps() Config {
	return Config{
		ClientID:    build.ClientID(),
		Authority:   build.Authority(),
		Scopes:      build.Scopes(),
		RedirectURI: build.SuccessRedirect(),
	}
}

type Authenticator struct {
	clientID    string
	authority   string
	scopes      []string
	redirectURI string
	cache       cache.ExportReplace
	flow        Flow
	printer     Printer
}

type Flow int

const (
	FlowInteractive Flow = iota
	FlowDeviceCode
)

type Option func(authenticator *Authenticator)

// NewAuthenticator creates a new authenticator with the given configuration.
func NewAuthenticator(config Config, options ...Option) *Authenticator {
	authenticator := &Authenticator{
		clientID:    config.ClientID,
		authority:   config.Authority,
		scopes:      config.Scopes,
		redirectURI: config.RedirectURI,
		cache:       tokencache.New(),
		flow:        FlowInteractive,
		printer:     nil,
	}

	for _, opt := range options {
		opt(authenticator)
	}

	return authenticator
}

// WithTokenCache configures the authenticator to use a custom token cache.
//
//goland:noinspection GoUnusedExportedFunction
func WithTokenCache(cache cache.ExportReplace) Option {
	return func(auth *Authenticator) {
		auth.cache = cache
	}
}

// WithDeviceCodeFlow configures the authenticator to use the device code flow
// for authentication. The printer is used to display the device code message to
// the user.
func WithDeviceCodeFlow(printer Printer) Option {
	return func(auth *Authenticator) {
		auth.flow = FlowDeviceCode
		auth.printer = printer
	}
}

// Authenticate initiates the authentication flow. If the user is already authenticated,
// the cached token is used. If the user is not authenticated, the user is prompted to
// authenticate using the configured flow.
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
		result, err = a.authenticateWithFlow(
			ctx,
			publicClient,
			a.flow,
			a.scopes,
		)
		if err != nil {
			return result, fmt.Errorf("could not acquire token: %w", err)
		}
	}

	return result, nil
}

func (a *Authenticator) authenticateWithFlow(
	ctx context.Context,
	publicClient public.Client,
	flow Flow,
	scopes []string,
) (public.AuthResult, error) {
	var (
		result public.AuthResult
		err    error
	)

	switch flow {
	case FlowInteractive:
		result, err = publicClient.AcquireTokenInteractive(ctx, scopes, public.WithRedirectURI(a.redirectURI))
	case FlowDeviceCode:
		var code public.DeviceCode

		code, err = publicClient.AcquireTokenByDeviceCode(ctx, scopes)
		if err != nil {
			return result, fmt.Errorf("could not acquire device code: %w", err)
		}

		if a.printer == nil {
			return result, errNoPrinter
		}

		err = a.printer(ctx, code.Result.Message)
		if err != nil {
			return result, redact.Errorf("could not print device code message: %w", redact.Safe(err))
		}

		// blocks until user has authenticated
		result, err = code.AuthenticationResult(ctx)
	default:
		return result, errFlowNotSupported
	}

	if err != nil {
		return result, redact.Errorf("could not acquire token: %w", redact.Safe(err))
	}

	return result, nil
}

// IsAuthenticated checks if the user is authenticated by checking if there are any
// cached accounts.
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

	if len(accounts) == 0 {
		return false, errNoAccounts
	}

	var result public.AuthResult

	result, err = publicClient.AcquireTokenSilent(
		ctx,
		a.scopes,
		public.WithSilentAccount(accounts[0]),
	)
	if err != nil {
		return false, fmt.Errorf("could not acquire token: %w", err)
	}

	if result.ExpiresOn.Before(time.Now()) {
		return false, errTokenExpired
	}

	for _, s := range a.scopes {
		if slices.Contains(result.DeclinedScopes, s) {
			return false, errDeclinedScopes
		}
	}

	return true, nil
}

type Printer func(ctx context.Context, message string) error
