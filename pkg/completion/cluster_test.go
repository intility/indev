package completion

import (
	"context"
	"errors"
	"testing"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/intility/indev/mocks"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

type fakeAuthenticator struct {
	authenticated bool
	err           error
}

func (f fakeAuthenticator) IsAuthenticated(context.Context) (bool, error) {
	return f.authenticated, f.err
}

func (f fakeAuthenticator) GetCurrentAccount(context.Context) (public.Account, error) {
	return public.Account{}, nil
}

func newCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	return cmd
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
