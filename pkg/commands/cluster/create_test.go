package cluster

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/intility/indev/mocks"
	"github.com/intility/indev/pkg/client"
	"github.com/stretchr/testify/mock"
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



func TestSelectSSOProvisioner(t *testing.T) {
	tests := []struct {
		name      string
		instances []client.IntegrationInstance
		err       error
		wantID    string
		wantErr   bool
		errSubstr string
	}{
		{
			name: "single EntraID provisioner returns its ID",
			instances: []client.IntegrationInstance{
				{ID: "prov-123", Type: "EntraID", Name: "My SSO"},
			},
			wantID:  "prov-123",
			wantErr: false,
		},
		{
			name: "multiple EntraID provisioners returns first",
			instances: []client.IntegrationInstance{
				{ID: "prov-123", Type: "EntraID", Name: "Primary SSO"},
				{ID: "prov-456", Type: "EntraID", Name: "Secondary SSO"},
			},
			wantID:  "prov-123",
			wantErr: false,
		},
		{
			name: "no EntraID provisioners returns error",
			instances: []client.IntegrationInstance{
				{ID: "prov-123", Type: "LDAP", Name: "LDAP Provider"},
			},
			wantID:    "",
			wantErr:   true,
			errSubstr: "no SSO provisioner configured",
		},
		{
			name:      "empty provisioners list returns error",
			instances: []client.IntegrationInstance{},
			wantID:    "",
			wantErr:   true,
			errSubstr: "no SSO provisioner configured",
		},
		{
			name:      "API error returns error",
			err:       errors.New("connection refused"),
			wantID:    "",
			wantErr:   true,
			errSubstr: "failed to list integration instances",
		},
		{
			name: "filters non-EntraID types",
			instances: []client.IntegrationInstance{
				{ID: "prov-100", Type: "LDAP", Name: "LDAP"},
				{ID: "prov-200", Type: "EntraID", Name: "Azure AD"},
				{ID: "prov-300", Type: "SAML", Name: "SAML Provider"},
			},
			wantID:  "prov-200",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := mocks.NewClient(t)
			mc.EXPECT().ListIntegrationInstances(mock.Anything).Return(tt.instances, tt.err)

			var buf bytes.Buffer
			id, err := selectSSOProvisioner(context.Background(), mc, &buf)

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
