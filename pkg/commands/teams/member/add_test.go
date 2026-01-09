package member

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/intility/indev/pkg/client"
)

func TestValidateAddOptions(t *testing.T) {
	tests := []struct {
		name    string
		options AddMemberOptions
		wantErr error
	}{
		// Team validation
		{
			name: "missing both team and team-id returns error",
			options: AddMemberOptions{
				Team:   "",
				TeamId: "",
				User:   "user@example.com",
				Role:   client.MemberRoleMember,
			},
			wantErr: errTeamRequired,
		},
		{
			name: "team name provided is valid",
			options: AddMemberOptions{
				Team:   "my-team",
				TeamId: "",
				User:   "user@example.com",
				Role:   client.MemberRoleMember,
			},
			wantErr: nil,
		},
		{
			name: "team-id provided is valid",
			options: AddMemberOptions{
				Team:   "",
				TeamId: "team-123",
				User:   "user@example.com",
				Role:   client.MemberRoleMember,
			},
			wantErr: nil,
		},
		{
			name: "both team and team-id provided is valid",
			options: AddMemberOptions{
				Team:   "my-team",
				TeamId: "team-123",
				User:   "user@example.com",
				Role:   client.MemberRoleMember,
			},
			wantErr: nil,
		},

		// User validation
		{
			name: "missing both user and user-id returns error",
			options: AddMemberOptions{
				Team:   "my-team",
				User:   "",
				UserId: "",
				Role:   client.MemberRoleMember,
			},
			wantErr: errUserRequired,
		},
		{
			name: "user UPN provided is valid",
			options: AddMemberOptions{
				Team:   "my-team",
				User:   "user@example.com",
				UserId: "",
				Role:   client.MemberRoleMember,
			},
			wantErr: nil,
		},
		{
			name: "user-id provided is valid",
			options: AddMemberOptions{
				Team:   "my-team",
				User:   "",
				UserId: "user-123",
				Role:   client.MemberRoleMember,
			},
			wantErr: nil,
		},

		// Role validation
		{
			name: "missing role returns error",
			options: AddMemberOptions{
				Team: "my-team",
				User: "user@example.com",
				Role: "",
			},
			wantErr: errRoleRequired,
		},
		{
			name: "owner role is valid",
			options: AddMemberOptions{
				Team: "my-team",
				User: "user@example.com",
				Role: client.MemberRoleOwner,
			},
			wantErr: nil,
		},
		{
			name: "member role is valid",
			options: AddMemberOptions{
				Team: "my-team",
				User: "user@example.com",
				Role: client.MemberRoleMember,
			},
			wantErr: nil,
		},
		{
			name: "invalid role returns error",
			options: AddMemberOptions{
				Team: "my-team",
				User: "user@example.com",
				Role: "admin",
			},
			wantErr: errInvalidRole,
		},
		{
			name: "another invalid role returns error",
			options: AddMemberOptions{
				Team: "my-team",
				User: "user@example.com",
				Role: "guest",
			},
			wantErr: errInvalidRole,
		},

		// Complete valid options
		{
			name: "all valid options with team name and user UPN",
			options: AddMemberOptions{
				Team: "platform-team",
				User: "alice@example.com",
				Role: client.MemberRoleOwner,
			},
			wantErr: nil,
		},
		{
			name: "all valid options with IDs",
			options: AddMemberOptions{
				TeamId: "team-abc-123",
				UserId: "user-xyz-789",
				Role:   client.MemberRoleMember,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAddOptions(tt.options)

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}
