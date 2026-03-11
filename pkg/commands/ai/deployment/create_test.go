package deployment

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
			name: "empty name returns error",
			options: CreateOptions{
				Name:  "",
				Model: "gpt-4o",
			},
			wantErr: errEmptyName,
		},
		{
			name: "empty model returns error",
			options: CreateOptions{
				Name:  "my-deployment",
				Model: "",
			},
			wantErr: errEmptyModel,
		},

		// Name length validation
		{
			name: "name at minimum length (3)",
			options: CreateOptions{
				Name:  "abc",
				Model: "gpt-4o",
			},
			wantErr: nil,
		},
		{
			name: "name below minimum length (2)",
			options: CreateOptions{
				Name:  "ab",
				Model: "gpt-4o",
			},
			wantErr: errInvalidNameLength,
		},
		{
			name: "name at maximum length (50)",
			options: CreateOptions{
				Name:  strings.Repeat("a", 50),
				Model: "gpt-4o",
			},
			wantErr: nil,
		},
		{
			name: "name above maximum length (51)",
			options: CreateOptions{
				Name:  strings.Repeat("a", 51),
				Model: "gpt-4o",
			},
			wantErr: errInvalidNameLength,
		},

		// Name format validation - valid patterns
		{
			name: "name with hyphens is valid",
			options: CreateOptions{
				Name:  "my-deployment",
				Model: "gpt-4o",
			},
			wantErr: nil,
		},
		{
			name: "name with numbers is valid",
			options: CreateOptions{
				Name:  "deploy123",
				Model: "gpt-4o",
			},
			wantErr: nil,
		},

		// Name format validation - invalid patterns
		{
			name: "name starting with hyphen is invalid",
			options: CreateOptions{
				Name:  "-deployment",
				Model: "gpt-4o",
			},
			wantErr: errInvalidNameFormat,
		},
		{
			name: "name ending with hyphen is invalid",
			options: CreateOptions{
				Name:  "deployment-",
				Model: "gpt-4o",
			},
			wantErr: errInvalidNameFormat,
		},
		{
			name: "name with uppercase is invalid",
			options: CreateOptions{
				Name:  "MyDeployment",
				Model: "gpt-4o",
			},
			wantErr: errInvalidNameFormat,
		},
		{
			name: "name with special characters is invalid",
			options: CreateOptions{
				Name:  "my@deployment",
				Model: "gpt-4o",
			},
			wantErr: errInvalidNameFormat,
		},
		{
			name: "name with underscore is invalid",
			options: CreateOptions{
				Name:  "my_deployment",
				Model: "gpt-4o",
			},
			wantErr: errInvalidNameFormat,
		},

		// Valid options
		{
			name: "valid options pass",
			options: CreateOptions{
				Name:  "my-deployment",
				Model: "gpt-4o",
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
