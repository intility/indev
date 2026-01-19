package cluster

import (
	"bytes"
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/intility/indev/pkg/client"
)

func TestValidateOptions(t *testing.T) {
	tests := []struct {
		name    string
		options CreateOptions
		wantErr error
	}{
		// Name validation
		{
			name: "empty name returns error",
			options: CreateOptions{
				Name:      "",
				Preset:    "minimal",
				NodeCount: 2,
			},
			wantErr: errEmptyName,
		},
		{
			name: "valid name passes",
			options: CreateOptions{
				Name:      "test-cluster",
				Preset:    "minimal",
				NodeCount: 2,
			},
			wantErr: nil,
		},

		// Preset validation
		{
			name: "valid preset minimal",
			options: CreateOptions{
				Name:      "test-cluster",
				Preset:    "minimal",
				NodeCount: 2,
			},
			wantErr: nil,
		},
		{
			name: "valid preset balanced",
			options: CreateOptions{
				Name:      "test-cluster",
				Preset:    "balanced",
				NodeCount: 2,
			},
			wantErr: nil,
		},
		{
			name: "valid preset performance",
			options: CreateOptions{
				Name:      "test-cluster",
				Preset:    "performance",
				NodeCount: 2,
			},
			wantErr: nil,
		},
		{
			name: "invalid preset returns error",
			options: CreateOptions{
				Name:      "test-cluster",
				Preset:    "invalid",
				NodeCount: 2,
			},
			wantErr: errInvalidPreset,
		},
		{
			name: "empty preset returns error",
			options: CreateOptions{
				Name:      "test-cluster",
				Preset:    "",
				NodeCount: 2,
			},
			wantErr: errInvalidPreset,
		},

		// Non-autoscaling node count validation
		{
			name: "node count at minimum boundary (2)",
			options: CreateOptions{
				Name:              "test-cluster",
				Preset:            "minimal",
				EnableAutoscaling: false,
				NodeCount:         2,
			},
			wantErr: nil,
		},
		{
			name: "node count at maximum boundary (8)",
			options: CreateOptions{
				Name:              "test-cluster",
				Preset:            "minimal",
				EnableAutoscaling: false,
				NodeCount:         8,
			},
			wantErr: nil,
		},
		{
			name: "node count below minimum (1)",
			options: CreateOptions{
				Name:              "test-cluster",
				Preset:            "minimal",
				EnableAutoscaling: false,
				NodeCount:         1,
			},
			wantErr: errInvalidNodeCount,
		},
		{
			name: "node count above maximum (9)",
			options: CreateOptions{
				Name:              "test-cluster",
				Preset:            "minimal",
				EnableAutoscaling: false,
				NodeCount:         9,
			},
			wantErr: errInvalidNodeCount,
		},
		{
			name: "node count zero",
			options: CreateOptions{
				Name:              "test-cluster",
				Preset:            "minimal",
				EnableAutoscaling: false,
				NodeCount:         0,
			},
			wantErr: errInvalidNodeCount,
		},

		// Autoscaling validation - MinNodes
		{
			name: "autoscaling min nodes at boundary (2)",
			options: CreateOptions{
				Name:              "test-cluster",
				Preset:            "minimal",
				EnableAutoscaling: true,
				MinNodes:          2,
				MaxNodes:          4,
			},
			wantErr: nil,
		},
		{
			name: "autoscaling min nodes below boundary (1)",
			options: CreateOptions{
				Name:              "test-cluster",
				Preset:            "minimal",
				EnableAutoscaling: true,
				MinNodes:          1,
				MaxNodes:          4,
			},
			wantErr: errInvalidMinNodes,
		},
		{
			name: "autoscaling min nodes above boundary (9)",
			options: CreateOptions{
				Name:              "test-cluster",
				Preset:            "minimal",
				EnableAutoscaling: true,
				MinNodes:          9,
				MaxNodes:          9,
			},
			wantErr: errInvalidMinNodes,
		},

		// Autoscaling validation - MaxNodes
		{
			name: "autoscaling max nodes at boundary (8)",
			options: CreateOptions{
				Name:              "test-cluster",
				Preset:            "minimal",
				EnableAutoscaling: true,
				MinNodes:          2,
				MaxNodes:          8,
			},
			wantErr: nil,
		},
		{
			name: "autoscaling max nodes above boundary (9)",
			options: CreateOptions{
				Name:              "test-cluster",
				Preset:            "minimal",
				EnableAutoscaling: true,
				MinNodes:          2,
				MaxNodes:          9,
			},
			wantErr: errInvalidMaxNodes,
		},

		// Autoscaling validation - Min > Max
		{
			name: "autoscaling min greater than max",
			options: CreateOptions{
				Name:              "test-cluster",
				Preset:            "minimal",
				EnableAutoscaling: true,
				MinNodes:          6,
				MaxNodes:          4,
			},
			wantErr: errMinGreaterThanMax,
		},
		{
			name: "autoscaling min equals max (valid)",
			options: CreateOptions{
				Name:              "test-cluster",
				Preset:            "minimal",
				EnableAutoscaling: true,
				MinNodes:          4,
				MaxNodes:          4,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOptions(tt.options)

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

func TestGenerateSuffix(t *testing.T) {
	t.Run("returns valid format", func(t *testing.T) {
		suffix := generateSuffix()
		want := regexp.MustCompile(`^[a-z0-9]{6}$`)
		assert.Regexp(t, want, suffix)
	})

	t.Run("returns correct length", func(t *testing.T) {
		suffix := generateSuffix()
		assert.Len(t, suffix, 6)
	})

	t.Run("generates unique suffixes", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 100; i++ {
			suffix := generateSuffix()
			seen[suffix] = true
		}
		// With 36^6 possible combinations, 100 samples should be mostly unique
		assert.Greater(t, len(seen), 90, "suffixes should be mostly unique")
	})
}

// mockClient implements client.Client for testing selectSSOProvisioner.
//
// TODO: Replace with generated mock once mockery compatibility is fixed.
// Steps to migrate:
//  1. Run `make mocks` to generate mocks/Client.go
//  2. Remove this inline mockClient type and all its method implementations
//  3. Update imports to include:
//     "github.com/intility/indev/mocks"
//     "github.com/stretchr/testify/mock"
//  4. Replace test setup with:
//     mockClient := mocks.NewClient(t)
//     mockClient.EXPECT().ListIntegrationInstances(mock.Anything).Return(instances, nil)
type mockClient struct {
	instances []client.IntegrationInstance
	err       error
}

func (m *mockClient) ListIntegrationInstances(_ context.Context) ([]client.IntegrationInstance, error) {
	return m.instances, m.err
}

// Stub implementations for other Client interface methods
func (m *mockClient) ListClusters(_ context.Context) (client.ClusterList, error) {
	return client.ClusterList{}, nil
}
func (m *mockClient) GetCluster(_ context.Context, _ string) (*client.Cluster, error) {
	return nil, nil
}
func (m *mockClient) GetClusterStatus(_ context.Context, _ string) (*client.Cluster, error) {
	return nil, nil
}
func (m *mockClient) CreateCluster(_ context.Context, _ client.NewClusterRequest) (*client.Cluster, error) {
	return nil, nil
}
func (m *mockClient) DeleteCluster(_ context.Context, _ string) error { return nil }
func (m *mockClient) GetClusterMembers(_ context.Context, _ string) ([]client.ClusterMember, error) {
	return nil, nil
}
func (m *mockClient) AddClusterMember(_ context.Context, _ string, _ []client.AddClusterMemberRequest) error {
	return nil
}
func (m *mockClient) GetMe(_ context.Context) (client.Me, error) { return client.Me{}, nil }
func (m *mockClient) ListTeams(_ context.Context) ([]client.Team, error) {
	return nil, nil
}
func (m *mockClient) GetTeamMembers(_ context.Context, _ string) ([]client.TeamMember, error) {
	return nil, nil
}
func (m *mockClient) CreateTeam(_ context.Context, _ client.NewTeamRequest) (*client.Team, error) {
	return nil, nil
}
func (m *mockClient) DeleteTeam(_ context.Context, _ client.DeleteTeamRequest) error { return nil }
func (m *mockClient) AddTeamMember(_ context.Context, _ string, _ []client.AddTeamMemberRequest) error {
	return nil
}
func (m *mockClient) RemoveTeamMember(_ context.Context, _, _ string) error { return nil }
func (m *mockClient) ListUsers(_ context.Context) ([]client.User, error)    { return nil, nil }

func TestSelectSSOProvisioner(t *testing.T) {
	tests := []struct {
		name      string
		client    *mockClient
		wantID    string
		wantErr   bool
		errSubstr string
	}{
		{
			name: "single EntraID provisioner returns its ID",
			client: &mockClient{
				instances: []client.IntegrationInstance{
					{ID: "prov-123", Type: "EntraID", Name: "My SSO"},
				},
			},
			wantID:  "prov-123",
			wantErr: false,
		},
		{
			name: "multiple EntraID provisioners returns first",
			client: &mockClient{
				instances: []client.IntegrationInstance{
					{ID: "prov-123", Type: "EntraID", Name: "Primary SSO"},
					{ID: "prov-456", Type: "EntraID", Name: "Secondary SSO"},
				},
			},
			wantID:  "prov-123",
			wantErr: false,
		},
		{
			name: "no EntraID provisioners returns error",
			client: &mockClient{
				instances: []client.IntegrationInstance{
					{ID: "prov-123", Type: "LDAP", Name: "LDAP Provider"},
				},
			},
			wantID:    "",
			wantErr:   true,
			errSubstr: "no SSO provisioner configured",
		},
		{
			name: "empty provisioners list returns error",
			client: &mockClient{
				instances: []client.IntegrationInstance{},
			},
			wantID:    "",
			wantErr:   true,
			errSubstr: "no SSO provisioner configured",
		},
		{
			name: "API error returns error",
			client: &mockClient{
				err: errors.New("connection refused"),
			},
			wantID:    "",
			wantErr:   true,
			errSubstr: "failed to list integration instances",
		},
		{
			name: "filters non-EntraID types",
			client: &mockClient{
				instances: []client.IntegrationInstance{
					{ID: "prov-100", Type: "LDAP", Name: "LDAP"},
					{ID: "prov-200", Type: "EntraID", Name: "Azure AD"},
					{ID: "prov-300", Type: "SAML", Name: "SAML Provider"},
				},
			},
			wantID:  "prov-200",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			id, err := selectSSOProvisioner(context.Background(), tt.client, &buf)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, id)
			}
		})
	}
}
