package completion

import (
	"context"
	"errors"
	"testing"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/intility/indev/internal/build"
	"github.com/intility/indev/mocks"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

type fakeAuthenticator struct {
	authenticated bool
	err           error
	account       public.Account
	accountErr    error
}

func (f fakeAuthenticator) IsAuthenticated(context.Context) (bool, error) {
	return f.authenticated, f.err
}

func (f fakeAuthenticator) GetCurrentAccount(context.Context) (public.Account, error) {
	return f.account, f.accountErr
}

type fakeNameCache struct {
	entries map[clientset.NameCacheKey][]string
	gets    []clientset.NameCacheKey
	sets    map[clientset.NameCacheKey][]string
	deletes []clientset.NameCacheKey
}

func newFakeNameCache() *fakeNameCache {
	return &fakeNameCache{
		entries: map[clientset.NameCacheKey][]string{},
		sets:    map[clientset.NameCacheKey][]string{},
	}
}

func (f *fakeNameCache) Get(key clientset.NameCacheKey) ([]string, bool) {
	f.gets = append(f.gets, key)

	names, ok := f.entries[key]

	return names, ok
}

func (f *fakeNameCache) Set(key clientset.NameCacheKey, names []string) {
	f.sets[key] = names
}

func (f *fakeNameCache) Delete(key clientset.NameCacheKey) {
	f.deletes = append(f.deletes, key)
}

func newCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	return cmd
}

func expectedKey(resource, accountID string) clientset.NameCacheKey {
	return clientset.NameCacheKey{Resource: resource, AccountID: accountID, Host: build.PlatformAPIHost()}
}

func TestClusterNames(t *testing.T) {
	tests := []struct {
		name          string
		authenticated bool
		authErr       error
		clusters      client.ClusterList
		listErr       error
		toComplete    string
		want          []cobra.Completion
	}{
		{
			name:          "empty input returns all clusters",
			authenticated: true,
			clusters:      client.ClusterList{{Name: "prod-east"}, {Name: "prod-west"}, {Name: "staging"}},
			toComplete:    "",
			want:          []cobra.Completion{"prod-east", "prod-west", "staging"},
		},
		{
			name:          "prefix filters clusters",
			authenticated: true,
			clusters:      client.ClusterList{{Name: "prod-east"}, {Name: "prod-west"}, {Name: "staging"}},
			toComplete:    "prod",
			want:          []cobra.Completion{"prod-east", "prod-west"},
		},
		{
			name:          "unique prefix returns single cluster",
			authenticated: true,
			clusters:      client.ClusterList{{Name: "prod-east"}, {Name: "prod-west"}, {Name: "staging"}},
			toComplete:    "st",
			want:          []cobra.Completion{"staging"},
		},
		{
			name:          "no match returns nothing",
			authenticated: true,
			clusters:      client.ClusterList{{Name: "prod-east"}, {Name: "prod-west"}, {Name: "staging"}},
			toComplete:    "dev",
			want:          nil,
		},
		{
			name:          "list failure returns nothing",
			authenticated: true,
			listErr:       errors.New("boom"),
			want:          nil,
		},
		{
			name:          "not signed in returns nothing",
			authenticated: false,
			want:          nil,
		},
		{
			name:    "auth check failure returns nothing",
			authErr: errors.New("token expired"),
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := mocks.NewClient(t)

			// Without a valid sign-in, the mock fails the test if ListClusters is called.
			if tt.authenticated && tt.authErr == nil {
				mc.EXPECT().ListClusters(mock.Anything).Return(tt.clusters, tt.listErr)
			}

			set := clientset.ClientSet{
				Authenticator:  fakeAuthenticator{authenticated: tt.authenticated, err: tt.authErr},
				PlatformClient: mc,
			}

			got, directive := ClusterNames(set)(newCommand(), nil, tt.toComplete)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
		})
	}
}

func TestClusterNameArgOnlyCompletesFirstArg(t *testing.T) {
	set := clientset.ClientSet{
		Authenticator:  fakeAuthenticator{authenticated: true},
		PlatformClient: mocks.NewClient(t),
	}

	got, directive := ClusterNameArg(set)(newCommand(), []string{"prod-east"}, "")

	assert.Empty(t, got)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestClusterNamesServesFreshCacheWithoutListing(t *testing.T) {
	cache := newFakeNameCache()
	cache.entries[expectedKey(ResourceClusters, "user-oid.tenant")] = []string{"prod-east", "staging"}

	set := clientset.ClientSet{
		Authenticator: fakeAuthenticator{
			authenticated: true,
			account:       public.Account{HomeAccountID: "user-oid.tenant"},
		},
		PlatformClient: mocks.NewClient(t),
		NameCache:      cache,
	}

	got, directive := ClusterNames(set)(newCommand(), nil, "prod")

	assert.Equal(t, []cobra.Completion{"prod-east"}, got)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	assert.Equal(t, []clientset.NameCacheKey{expectedKey(ResourceClusters, "user-oid.tenant")}, cache.gets)
}

func TestClusterNamesCacheHitSkipsAuthCheck(t *testing.T) {
	cache := newFakeNameCache()
	cache.entries[expectedKey(ResourceClusters, "user-oid.tenant")] = []string{"prod-east"}

	set := clientset.ClientSet{
		Authenticator: fakeAuthenticator{
			authenticated: false,
			account:       public.Account{HomeAccountID: "user-oid.tenant"},
		},
		PlatformClient: mocks.NewClient(t),
		NameCache:      cache,
	}

	got, _ := ClusterNames(set)(newCommand(), nil, "")

	assert.Equal(t, []cobra.Completion{"prod-east"}, got)
}

func TestClusterNamesPopulatesCacheAfterLiveList(t *testing.T) {
	mc := mocks.NewClient(t)
	mc.EXPECT().ListClusters(mock.Anything).Return(client.ClusterList{{Name: "prod-east"}, {Name: "staging"}}, nil)

	cache := newFakeNameCache()

	set := clientset.ClientSet{
		Authenticator: fakeAuthenticator{
			authenticated: true,
			account:       public.Account{HomeAccountID: "user-oid.tenant"},
		},
		PlatformClient: mc,
		NameCache:      cache,
	}

	got, _ := ClusterNames(set)(newCommand(), nil, "")

	assert.Equal(t, []cobra.Completion{"prod-east", "staging"}, got)
	assert.Equal(t, map[clientset.NameCacheKey][]string{expectedKey(ResourceClusters, "user-oid.tenant"): {"prod-east", "staging"}}, cache.sets)
}

func TestClusterNamesFallsThroughToLiveListOnPrefixMiss(t *testing.T) {
	mc := mocks.NewClient(t)
	mc.EXPECT().ListClusters(mock.Anything).Return(client.ClusterList{{Name: "prod-east"}}, nil)

	cache := newFakeNameCache()
	cache.entries[expectedKey(ResourceClusters, "user-oid.tenant")] = []string{"staging"}

	set := clientset.ClientSet{
		Authenticator: fakeAuthenticator{
			authenticated: true,
			account:       public.Account{HomeAccountID: "user-oid.tenant"},
		},
		PlatformClient: mc,
		NameCache:      cache,
	}

	got, _ := ClusterNames(set)(newCommand(), nil, "prod")

	assert.Equal(t, []cobra.Completion{"prod-east"}, got)
	assert.Equal(t, map[clientset.NameCacheKey][]string{expectedKey(ResourceClusters, "user-oid.tenant"): {"prod-east"}}, cache.sets)
}

func TestClusterNamesIgnoresCacheWhenAccountUnresolved(t *testing.T) {
	mc := mocks.NewClient(t)
	mc.EXPECT().ListClusters(mock.Anything).Return(client.ClusterList{{Name: "prod-east"}}, nil)

	cache := newFakeNameCache()

	set := clientset.ClientSet{
		Authenticator: fakeAuthenticator{
			authenticated: true,
			accountErr:    errors.New("no account"),
		},
		PlatformClient: mc,
		NameCache:      cache,
	}

	got, _ := ClusterNames(set)(newCommand(), nil, "")

	assert.Equal(t, []cobra.Completion{"prod-east"}, got)
	assert.Empty(t, cache.gets)
	assert.Empty(t, cache.sets)
}

func TestInvalidate(t *testing.T) {
	cache := newFakeNameCache()

	set := clientset.ClientSet{
		Authenticator: fakeAuthenticator{account: public.Account{HomeAccountID: "user-oid.tenant"}},
		NameCache:     cache,
	}

	Invalidate(context.Background(), set, ResourceClusters)

	assert.Equal(t, []clientset.NameCacheKey{expectedKey(ResourceClusters, "user-oid.tenant")}, cache.deletes)
}

func TestInvalidateWithoutCacheOrAccount(t *testing.T) {
	Invalidate(context.Background(), clientset.ClientSet{}, ResourceClusters)

	cache := newFakeNameCache()
	unresolved := clientset.ClientSet{
		Authenticator: fakeAuthenticator{accountErr: errors.New("no account")},
		NameCache:     cache,
	}

	Invalidate(context.Background(), unresolved, ResourceClusters)

	assert.Empty(t, cache.deletes)
}

func TestSetNames(t *testing.T) {
	cache := newFakeNameCache()

	set := clientset.ClientSet{
		Authenticator: fakeAuthenticator{account: public.Account{HomeAccountID: "user-oid.tenant"}},
		NameCache:     cache,
	}

	SetNames(context.Background(), set, ResourceClusters, []string{"prod-east", "staging"})

	assert.Equal(t, map[clientset.NameCacheKey][]string{expectedKey(ResourceClusters, "user-oid.tenant"): {"prod-east", "staging"}}, cache.sets)
}

func TestSetNamesWithoutCacheOrAccount(t *testing.T) {
	SetNames(context.Background(), clientset.ClientSet{}, ResourceClusters, []string{"prod-east"})

	cache := newFakeNameCache()
	unresolved := clientset.ClientSet{
		Authenticator: fakeAuthenticator{accountErr: errors.New("no account")},
		NameCache:     cache,
	}

	SetNames(context.Background(), unresolved, ResourceClusters, []string{"prod-east"})

	assert.Empty(t, cache.sets)
}
