package account

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOrganizationalRole(t *testing.T) {
	tests := []struct {
		name  string
		roles []string
		want  string
	}{
		{
			name:  "owner role returns Admin",
			roles: []string{"owner"},
			want:  "Admin",
		},
		{
			name:  "member role returns Member",
			roles: []string{"member"},
			want:  "Member",
		},
		{
			name:  "owner takes precedence over member",
			roles: []string{"member", "owner"},
			want:  "Admin",
		},
		{
			name:  "owner with other roles returns Admin",
			roles: []string{"viewer", "owner", "contributor"},
			want:  "Admin",
		},
		{
			name:  "member with other non-owner roles returns Member",
			roles: []string{"viewer", "member", "contributor"},
			want:  "Member",
		},
		{
			name:  "empty roles returns Unknown",
			roles: []string{},
			want:  "Unknown",
		},
		{
			name:  "nil roles returns Unknown",
			roles: nil,
			want:  "Unknown",
		},
		{
			name:  "unrecognized role returns Unknown",
			roles: []string{"admin", "superuser"},
			want:  "Unknown",
		},
		{
			name:  "single unrecognized role returns Unknown",
			roles: []string{"viewer"},
			want:  "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getOrganizationalRole(tt.roles)
			assert.Equal(t, tt.want, got)
		})
	}
}
