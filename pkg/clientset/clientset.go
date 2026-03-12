package clientset

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/build"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/pkg/client"
)

// homeAccountIDParts is the expected number of parts when splitting
// HomeAccountID by ".". Format is "<oid>.<tid>".
const homeAccountIDParts = 2

var (
	errNotAuthenticatedPreHook = errors.New("you need to sign in before executing this operation")
	errInvalidHomeAccountID    = errors.New("invalid HomeAccountID format")
	errFeatureNotAvailable     = errors.New("this feature is not available for your tenant")
)

type Authenticator interface {
	IsAuthenticated(ctx context.Context) (bool, error)
	GetCurrentAccount(ctx context.Context) (public.Account, error)
}

type ClientSet struct {
	Authenticator  Authenticator
	PlatformClient client.Client
}

func (c *ClientSet) EnsureSignedIn(cmd *cobra.Command, _ []string) error {
	ctx, span := telemetry.StartSpan(cmd.Context(), "EnsureSignedIn")
	defer span.End()

	isAuthenticated, err := c.Authenticator.IsAuthenticated(ctx)
	if !isAuthenticated || err != nil {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true

		return errNotAuthenticatedPreHook
	}

	return nil
}

type Hook func(cmd *cobra.Command, args []string) error

func (c *ClientSet) EnsureSignedInPreHook(cmd *cobra.Command, args []string) error {
	return c.PreHooks(c.EnsureSignedIn)(cmd, args)
}

type hookType string

const (
	PreRun  hookType = "pre"
	PostRun hookType = "post"
)

func (ht hookType) String() string {
	return string(ht)
}

func (c *ClientSet) PreHooks(hooks ...Hook) Hook {
	return c.hooks(PreRun, hooks...)
}

func (c *ClientSet) PostHooks(hooks ...Hook) Hook {
	return c.hooks(PostRun, hooks...)
}

func (c *ClientSet) hooks(typ hookType, hooks ...Hook) Hook {
	return func(cmd *cobra.Command, args []string) error {
		cmdName := fqcn(cmd)

		ctx, span := telemetry.StartSpan(cmd.Context(), fmt.Sprintf("%s.%s%s", cmdName, typ, "Hook"))
		defer span.End()

		cmd.SetContext(ctx)

		for _, hook := range hooks {
			if err := hook(cmd, args); err != nil {
				return err
			}
		}

		return nil
	}
}

// fqcn returns the fully qualified command name based on its parents.
func fqcn(cmd *cobra.Command) string {
	parents := make([]string, 0)

	current := cmd

	for current.HasParent() && current.Parent().Name() != build.AppName {
		parents = append(parents, current.Parent().Name())
		current = current.Parent()
	}

	name := strings.Join(parents, ".") + "." + cmd.Name()

	return name
}

// EnsureAITenantPreHook is a pre-run hook that checks if the current tenant
// has access to AI features. It composes EnsureSignedIn so auth is checked first.
func (c *ClientSet) EnsureAITenantPreHook(cmd *cobra.Command, args []string) error {
	return c.PreHooks(c.EnsureSignedIn, c.ensureAITenant)(cmd, args)
}

func (c *ClientSet) ensureAITenant(cmd *cobra.Command, _ []string) error {
	allowedTenants := map[string]bool{
		"9b5ff18e-53c0-45a2-8bc2-9c0c8f60b2c6": true, // intility
		"f0a51c83-6cc9-4830-9d01-d4d18c068e32": true, // skoglab
	}

	tenantID, err := c.GetTenantID(cmd.Context())
	if err != nil {
		return fmt.Errorf("could not determine tenant: %w", err)
	}

	if !allowedTenants[tenantID] {
		cmd.SilenceUsage = true

		return errFeatureNotAvailable
	}

	return nil
}

// GetTenantID extracts the tenant ID from the current account's HomeAccountID.
// The HomeAccountID format is "<oid>.<tid>" where tid is the tenant ID.
func (c *ClientSet) GetTenantID(ctx context.Context) (string, error) {
	account, err := c.Authenticator.GetCurrentAccount(ctx)
	if err != nil {
		return "", fmt.Errorf("could not get current account: %w", err)
	}

	parts := strings.Split(account.HomeAccountID, ".")
	if len(parts) < homeAccountIDParts {
		return "", fmt.Errorf("%w: %s", errInvalidHomeAccountID, account.HomeAccountID)
	}

	return parts[1], nil
}
