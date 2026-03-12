package apikey

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCreateOptions(t *testing.T) {
	tests := []struct {
		name    string
		options CreateOptions
		wantErr error
	}{
		// Empty field validation
		{
			name: "empty deployment returns error",
			options: CreateOptions{
				Deployment: "",
				Name:       "my-key",
				TTLDays:    30,
			},
			wantErr: errEmptyDeployment,
		},
		{
			name: "empty name returns error",
			options: CreateOptions{
				Deployment: "my-deployment",
				Name:       "",
				TTLDays:    30,
			},
			wantErr: errEmptyName,
		},

		// TTL validation
		{
			name: "zero TTL returns error",
			options: CreateOptions{
				Deployment: "my-deployment",
				Name:       "my-key",
				TTLDays:    0,
			},
			wantErr: errInvalidTTL,
		},
		{
			name: "negative TTL returns error",
			options: CreateOptions{
				Deployment: "my-deployment",
				Name:       "my-key",
				TTLDays:    -1,
			},
			wantErr: errInvalidTTL,
		},
		{
			name: "TTL above 365 returns error",
			options: CreateOptions{
				Deployment: "my-deployment",
				Name:       "my-key",
				TTLDays:    366,
			},
			wantErr: errInvalidTTL,
		},
		{
			name: "TTL at 1 is valid",
			options: CreateOptions{
				Deployment: "my-deployment",
				Name:       "my-key",
				TTLDays:    1,
			},
			wantErr: nil,
		},
		{
			name: "TTL at 365 is valid",
			options: CreateOptions{
				Deployment: "my-deployment",
				Name:       "my-key",
				TTLDays:    365,
			},
			wantErr: nil,
		},

		// Name length validation
		{
			name: "name at minimum length (3)",
			options: CreateOptions{
				Deployment: "my-deployment",
				Name:       "abc",
				TTLDays:    30,
			},
			wantErr: nil,
		},
		{
			name: "name below minimum length (2)",
			options: CreateOptions{
				Deployment: "my-deployment",
				Name:       "ab",
				TTLDays:    30,
			},
			wantErr: errInvalidNameLength,
		},
		{
			name: "name at maximum length (50)",
			options: CreateOptions{
				Deployment: "my-deployment",
				Name:       strings.Repeat("a", 50),
				TTLDays:    30,
			},
			wantErr: nil,
		},
		{
			name: "name above maximum length (51)",
			options: CreateOptions{
				Deployment: "my-deployment",
				Name:       strings.Repeat("a", 51),
				TTLDays:    30,
			},
			wantErr: errInvalidNameLength,
		},

		// Name format validation - valid patterns
		{
			name: "name with hyphens is valid",
			options: CreateOptions{
				Deployment: "my-deployment",
				Name:       "my-key",
				TTLDays:    30,
			},
			wantErr: nil,
		},
		{
			name: "name with numbers is valid",
			options: CreateOptions{
				Deployment: "my-deployment",
				Name:       "key123",
				TTLDays:    30,
			},
			wantErr: nil,
		},

		// Name format validation - invalid patterns
		{
			name: "name starting with hyphen is invalid",
			options: CreateOptions{
				Deployment: "my-deployment",
				Name:       "-my-key",
				TTLDays:    30,
			},
			wantErr: errInvalidNameFormat,
		},
		{
			name: "name ending with hyphen is invalid",
			options: CreateOptions{
				Deployment: "my-deployment",
				Name:       "my-key-",
				TTLDays:    30,
			},
			wantErr: errInvalidNameFormat,
		},
		{
			name: "name with uppercase is invalid",
			options: CreateOptions{
				Deployment: "my-deployment",
				Name:       "MyKey",
				TTLDays:    30,
			},
			wantErr: errInvalidNameFormat,
		},
		{
			name: "name with special characters is invalid",
			options: CreateOptions{
				Deployment: "my-deployment",
				Name:       "my@key",
				TTLDays:    30,
			},
			wantErr: errInvalidNameFormat,
		},

		// Valid options
		{
			name: "valid options pass",
			options: CreateOptions{
				Deployment: "my-deployment",
				Name:       "my-key",
				TTLDays:    30,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreateOptions(tt.options)

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}
