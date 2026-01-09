package teams

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
		// Name validation - empty
		{
			name: "empty name returns error",
			options: CreateOptions{
				Name:        "",
				Description: "A valid description",
			},
			wantErr: errEmptyName,
		},

		// Description validation - empty
		{
			name: "empty description returns error",
			options: CreateOptions{
				Name:        "valid-team",
				Description: "",
			},
			wantErr: errEmptyDescription,
		},

		// Name length validation
		{
			name: "name at minimum length (3)",
			options: CreateOptions{
				Name:        "abc",
				Description: "A valid description",
			},
			wantErr: nil,
		},
		{
			name: "name below minimum length (2)",
			options: CreateOptions{
				Name:        "ab",
				Description: "A valid description",
			},
			wantErr: errInvalidNameLength,
		},
		{
			name: "name at maximum length (50)",
			options: CreateOptions{
				Name:        strings.Repeat("a", 50),
				Description: "A valid description",
			},
			wantErr: nil,
		},
		{
			name: "name above maximum length (51)",
			options: CreateOptions{
				Name:        strings.Repeat("a", 51),
				Description: "A valid description",
			},
			wantErr: errInvalidNameLength,
		},

		// Description length validation
		{
			name: "description at minimum length (1)",
			options: CreateOptions{
				Name:        "valid-team",
				Description: "A",
			},
			wantErr: nil,
		},
		{
			name: "description at maximum length (100)",
			options: CreateOptions{
				Name:        "valid-team",
				Description: strings.Repeat("a", 100),
			},
			wantErr: nil,
		},
		{
			name: "description above maximum length (101)",
			options: CreateOptions{
				Name:        "valid-team",
				Description: strings.Repeat("a", 101),
			},
			wantErr: errInvalidDescLength,
		},

		// Valid options
		{
			name: "valid options pass",
			options: CreateOptions{
				Name:        "my-team",
				Description: "This is a great team",
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
