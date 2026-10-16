package completion

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/build"
	"github.com/intility/indev/pkg/clientset"
)

func completeNames(
	set clientset.ClientSet,
	resource string,
	listNames func(context.Context, clientset.ClientSet) ([]string, error),
) cobra.CompletionFunc {
	return func(cmd *cobra.Command, _ []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
		defer cancel()

		cache := set.NameCache

		key, resolved := nameCacheKey(ctx, set, resource)

		if cache != nil && resolved {
			if names, ok := cache.Get(key); ok {
				completions := filterCompletions(names, toComplete)

				if len(completions) > 0 {
					return completions, cobra.ShellCompDirectiveNoFileComp
				}
			}
		}

		if !signedIn(ctx, set) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		names, err := listNames(ctx, set)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		if cache != nil && resolved {
			cache.Set(key, names)
		}

		return filterCompletions(names, toComplete), cobra.ShellCompDirectiveNoFileComp
	}
}

// Invalidate removes the current account's cached names for a resource.
func Invalidate(ctx context.Context, set clientset.ClientSet, resource string) {
	if set.NameCache == nil {
		return
	}

	if key, ok := nameCacheKey(ctx, set, resource); ok {
		set.NameCache.Delete(key)
	}
}

// SetNames caches names for the current account and resource.
func SetNames(ctx context.Context, set clientset.ClientSet, resource string, names []string) {
	if set.NameCache == nil {
		return
	}

	if key, ok := nameCacheKey(ctx, set, resource); ok {
		set.NameCache.Set(key, names)
	}
}

func nameCacheKey(ctx context.Context, set clientset.ClientSet, resource string) (clientset.NameCacheKey, bool) {
	account, err := set.Authenticator.GetCurrentAccount(ctx)
	if err != nil {
		return clientset.NameCacheKey{Resource: resource, AccountID: "", Host: ""}, false
	}

	return clientset.NameCacheKey{
		Resource:  resource,
		AccountID: account.HomeAccountID,
		Host:      build.PlatformAPIHost(),
	}, true
}

func filterCompletions(names []string, toComplete string) []cobra.Completion {
	var completions []cobra.Completion

	for _, name := range names {
		if strings.HasPrefix(name, toComplete) {
			completions = append(completions, name)
		}
	}

	return completions
}
