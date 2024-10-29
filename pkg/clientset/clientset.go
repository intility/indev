package clientset

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/spf13/cobra"

	"github.com/intility/idpctl/internal/build"
	"github.com/intility/idpctl/internal/telemetry"
	"github.com/intility/idpctl/pkg/client"
)

var errNotAuthenticatedPreHook = errors.New("you need to sign in before executing this operation")

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

	for {
		if !current.HasParent() || current.Parent().Name() == build.AppName {
			break
		}

		parents = append(parents, current.Parent().Name())
		current = current.Parent()
	}

	name := strings.Join(parents, ".") + "." + cmd.Name()

	return name
}
