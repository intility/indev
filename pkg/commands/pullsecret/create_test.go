package pullsecret

import (
	"errors"
	"testing"
)

func TestParseRegistryFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    RegistryEntry
		wantErr error
	}{
		{
			name:  "valid address username password",
			input: "ghcr.io:myuser:mypass",
			want:  RegistryEntry{Address: "ghcr.io", Username: "myuser", Password: "mypass"},
		},
		{
			name:  "password with colons",
			input: "ghcr.io:myuser:pa:ss:w0rd",
			want:  RegistryEntry{Address: "ghcr.io", Username: "myuser", Password: "pa:ss:w0rd"},
		},
		{
			name:    "missing password",
			input:   "ghcr.io:myuser",
			wantErr: errInvalidRegistry,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: errInvalidRegistry,
		},
		{
			name:    "address only",
			input:   "ghcr.io",
			wantErr: errInvalidRegistry,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseRegistryFlag(tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("parseRegistryFlag(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("parseRegistryFlag(%q) unexpected error: %v", tt.input, err)
			}

			if got != tt.want {
				t.Errorf("parseRegistryFlag(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateCreateOptions(t *testing.T) {
	t.Parallel()

	validEntry := RegistryEntry{Address: "ghcr.io", Username: "user", Password: "pass"}

	tests := []struct {
		name    string
		options CreateOptions
		wantErr error
	}{
		{
			name:    "empty name",
			options: CreateOptions{Name: "", Registries: []RegistryEntry{validEntry}},
			wantErr: errEmptyPullSecretName,
		},
		{
			name:    "no registries",
			options: CreateOptions{Name: "mysecret", Registries: []RegistryEntry{}},
			wantErr: errNoRegistries,
		},
		{
			name: "empty address",
			options: CreateOptions{
				Name:       "mysecret",
				Registries: []RegistryEntry{{Address: "", Username: "user", Password: "pass"}},
			},
			wantErr: errEmptyAddress,
		},
		{
			name: "empty username",
			options: CreateOptions{
				Name:       "mysecret",
				Registries: []RegistryEntry{{Address: "ghcr.io", Username: "", Password: "pass"}},
			},
			wantErr: errEmptyUsername,
		},
		{
			name: "empty password",
			options: CreateOptions{
				Name:       "mysecret",
				Registries: []RegistryEntry{{Address: "ghcr.io", Username: "user", Password: ""}},
			},
			wantErr: errEmptyPassword,
		},
		{
			name: "duplicate address",
			options: CreateOptions{
				Name:       "mysecret",
				Registries: []RegistryEntry{validEntry, validEntry},
			},
			wantErr: errDuplicateAddress,
		},
		{
			name:    "valid single registry",
			options: CreateOptions{Name: "mysecret", Registries: []RegistryEntry{validEntry}},
		},
		{
			name: "valid multiple registries",
			options: CreateOptions{
				Name: "mysecret",
				Registries: []RegistryEntry{
					{Address: "ghcr.io", Username: "user1", Password: "pass1"},
					{Address: "docker.io", Username: "user2", Password: "pass2"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateCreateOptions(tt.options)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("validateCreateOptions() error = %v, wantErr %v", err, tt.wantErr)
				}

				return
			}

			if err != nil {
				t.Errorf("validateCreateOptions() unexpected error: %v", err)
			}
		})
	}
}
