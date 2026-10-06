// Package completion provides dynamic shell completion functions for commands.
package completion

import (
	"context"
	"time"

	"github.com/intility/indev/pkg/clientset"
)

// timeout bounds how long a completion request may block the shell.
const timeout = 3 * time.Second

// signedIn reports whether the user is signed in.
// It never triggers an interactive sign-in, which would hang the shell.
func signedIn(ctx context.Context, set clientset.ClientSet) bool {
	authenticated, err := set.Authenticator.IsAuthenticated(ctx)

	return err == nil && authenticated
}
