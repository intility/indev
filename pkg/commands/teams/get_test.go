package teams

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/intility/indev/pkg/client"
)

func TestGetTeamRole(t *testing.T) {
	tests := []struct {
		name  string
		roles []client.MemberRole
		want  string
	}{
		{
			name:  "owner role returns Owner",
			roles: []client.MemberRole{client.MemberRoleOwner},
			want:  "Owner",
		},
		{
			name:  "member role returns Member",
			roles: []client.MemberRole{client.MemberRoleMember},
			want:  "Member",
		},
		{
			name:  "owner takes precedence over member",
			roles: []client.MemberRole{client.MemberRoleMember, client.MemberRoleOwner},
			want:  "Owner",
		},
		{
			name:  "empty roles returns None",
			roles: []client.MemberRole{},
			want:  "None",
		},
		{
			name:  "nil roles returns None",
			roles: nil,
			want:  "None",
		},
		{
			name:  "unknown role returns Unknown",
			roles: []client.MemberRole{"admin"},
			want:  "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTeamRole(tt.roles)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrintTeamDetails(t *testing.T) {
	tests := []struct {
		name    string
		team    *client.Team
		members []client.TeamMember
		check   func(t *testing.T, output string)
	}{
		{
			name: "prints team with members",
			team: &client.Team{
				ID:          "team-123",
				Name:        "Platform Team",
				Description: "The platform team",
			},
			members: []client.TeamMember{
				{
					Subject: client.Subject{Name: "Alice"},
					Roles:   []client.MemberRole{client.MemberRoleOwner},
				},
				{
					Subject: client.Subject{Name: "Bob"},
					Roles:   []client.MemberRole{client.MemberRoleMember},
				},
			},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "team-123")
				assert.Contains(t, output, "Platform Team")
				assert.Contains(t, output, "The platform team")
				assert.Contains(t, output, "Alice")
				assert.Contains(t, output, "Bob")
				assert.Contains(t, output, "Owner")
				assert.Contains(t, output, "Member")
			},
		},
		{
			name: "prints team with no members",
			team: &client.Team{
				ID:          "team-456",
				Name:        "Empty Team",
				Description: "A team with no members",
			},
			members: []client.TeamMember{},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "team-456")
				assert.Contains(t, output, "Empty Team")
				assert.Contains(t, output, "Members:")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printTeamDetails(&buf, tt.team, tt.members)

			tt.check(t, buf.String())
		})
	}
}
