package authenticator

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"

	"github.com/intility/icpctl/internal/build"
	"github.com/intility/icpctl/internal/redact"
	"github.com/intility/icpctl/internal/telemetry"
	"github.com/intility/icpctl/pkg/tokencache"
)

var (
	ErrNoPrinter        = errors.New("printer is required for device code flow")
	ErrFlowNotSupported = errors.New("unsupported authentication flow")
	ErrNoAccounts       = errors.New("no accounts found")
	ErrTokenExpired     = errors.New("token has expired")
	ErrDeclinedScopes   = errors.New("scopes have been declined")
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

	ctx, span := telemetry.StartSpan(ctx, "Authenticate")
	defer span.End()

	publicClient, err := a.createPublicClient(ctx)
	if err != nil {
		return result, err
	}

	accounts, err := getCachedAccounts(publicClient, ctx)
	if len(accounts) > 0 {
		ctx, span = telemetry.StartSpan(ctx, "SilentAcquisition")
		defer span.End()

		result, err = a.acquireTokenSilent(ctx, publicClient, accounts[0])
	}

	if err != nil || len(accounts) == 0 {
		result, err = a.authenticateWithFlow(
			ctx,
			publicClient,
			a.flow,
			a.scopes,
		)
		if err != nil {
			span.RecordError(err)
			return result, fmt.Errorf("could not acquire token: %w", err)
		}
	}

	return result, nil
}

func (a *Authenticator) createPublicClient(ctx context.Context) (public.Client, error) {
	ctx, span := telemetry.StartSpan(ctx, "CreatePublicClient")
	defer span.End()

	client, err := public.New(
		a.clientID,
		public.WithAuthority(a.authority),
		public.WithCache(a.cache),
	)
	if err != nil {
		return client, fmt.Errorf("could not create public client: %w", err)
	}

	return client, nil
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
		ctx, span := telemetry.StartSpan(ctx, "InteractiveAcquisition")
		defer span.End()

		result, err = publicClient.AcquireTokenInteractive(ctx, scopes, public.WithRedirectURI(a.redirectURI))
	case FlowDeviceCode:
		ctx, span := telemetry.StartSpan(ctx, "DeviceCodeAcquisition")
		defer span.End()

		var code public.DeviceCode

		code, err = publicClient.AcquireTokenByDeviceCode(ctx, scopes)
		if err != nil {
			return result, fmt.Errorf("could not acquire device code: %w", err)
		}

		if a.printer == nil {
			return result, ErrNoPrinter
		}

		err = a.printer(ctx, code.Result.Message)
		if err != nil {
			return result, redact.Errorf("could not print device code message: %w", redact.Safe(err))
		}

		ctx, span = telemetry.StartSpan(ctx, "DeviceCodeAuthentication")
		defer span.End()

		// blocks until user has authenticated
		result, err = code.AuthenticationResult(ctx)
	default:
		return result, ErrFlowNotSupported
	}

	if err != nil {
		return result, redact.Errorf("could not acquire token: %w", redact.Safe(err))
	}

	return result, nil
}

// IsAuthenticated checks if the user is authenticated by checking if there are any
// cached accounts.
func (a *Authenticator) IsAuthenticated(ctx context.Context) (bool, error) {
	ctx, span := telemetry.StartSpan(ctx, "IsAuthenticated")
	defer span.End()

	publicClient, err := a.createPublicClient(ctx)
	if err != nil {
		return false, err
	}

	accounts, err := getCachedAccounts(publicClient, ctx)
	if err != nil {
		return false, err
	}

	if len(accounts) == 0 {
		return false, ErrNoAccounts
	}

	result, err := a.acquireTokenSilent(ctx, publicClient, accounts[0])
	if err != nil {
		return false, fmt.Errorf("could not acquire token: %w", err)
	}

	if result.ExpiresOn.Before(time.Now()) {
		return false, ErrTokenExpired
	}

	for _, s := range a.scopes {
		if slices.Contains(result.DeclinedScopes, s) {
			return false, ErrDeclinedScopes
		}
	}

	return true, nil
}

func (a *Authenticator) GetCurrentAccount(ctx context.Context) (public.Account, error) {
	ctx, span := telemetry.StartSpan(ctx, "GetCurrentAccount")
	defer span.End()

	publicClient, err := a.createPublicClient(ctx)
	if err != nil {
		return public.Account{}, err
	}

	accounts, err := getCachedAccounts(publicClient, ctx)
	if err != nil {
		return public.Account{}, err
	}

	if len(accounts) == 0 {
		return public.Account{}, ErrNoAccounts
	}

	return accounts[0], nil
}

func (a *Authenticator) acquireTokenSilent(
	ctx context.Context,
	client public.Client,
	account public.Account,
) (public.AuthResult, error) {
	ctx, span := telemetry.StartSpan(ctx, "SilentAcquisition")
	defer span.End()

	result, err := client.AcquireTokenSilent(
		ctx,
		a.scopes,
		public.WithSilentAccount(account))

	if err != nil {
		span.RecordError(err)
		return result, fmt.Errorf("could not acquire token silently: %w", err)
	}

	span.AddEvent("acquired token silently")

	return result, nil
}

func getCachedAccounts(client public.Client, ctx context.Context) ([]public.Account, error) {
	ctx, span := telemetry.StartSpan(ctx, "GetCachedAccounts")
	defer span.End()

	accounts, err := client.Accounts(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("could not get accounts: %w", err)
	}

	if len(accounts) == 0 {
		span.AddEvent("no cached accounts found")
		return accounts, nil
	}

	span.AddEvent("found cached accounts")

	return accounts, nil
}

type Printer func(ctx context.Context, message string) error
