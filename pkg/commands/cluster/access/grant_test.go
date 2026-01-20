package access

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/intility/indev/pkg/client"
)

func TestValidateGrantOptions(t *testing.T) {
	tests := []struct {
		name    string
		options GrantOptions
		wantErr error
	}{
		// Cluster validation
		{
			name: "missing both cluster and cluster-id returns error",
			options: GrantOptions{
				Cluster:   "",
				ClusterID: "",
				User:      "user@example.com",
				Role:      client.ClusterMemberRoleAdmin,
			},
			wantErr: errClusterRequired,
		},
		{
			name: "cluster name provided is valid",
			options: GrantOptions{
				Cluster: "my-cluster",
				User:    "user@example.com",
				Role:    client.ClusterMemberRoleAdmin,
			},
			wantErr: nil,
		},
		{
			name: "cluster-id provided is valid",
			options: GrantOptions{
				ClusterID: "cluster-123",
				User:      "user@example.com",
				Role:      client.ClusterMemberRoleAdmin,
			},
			wantErr: nil,
		},
		{
			name: "both cluster and cluster-id provided is valid",
			options: GrantOptions{
				Cluster:   "my-cluster",
				ClusterID: "cluster-123",
				User:      "user@example.com",
				Role:      client.ClusterMemberRoleAdmin,
			},
			wantErr: nil,
		},

		// Subject validation (user OR team, exactly one)
		{
			name: "missing both user/user-id and team/team-id returns error",
			options: GrantOptions{
				Cluster: "my-cluster",
				User:    "",
				UserID:  "",
				Team:    "",
				TeamID:  "",
				Role:    client.ClusterMemberRoleAdmin,
			},
			wantErr: errSubjectRequired,
		},
		{
			name: "both user and team specified returns error",
			options: GrantOptions{
				Cluster: "my-cluster",
				User:    "user@example.com",
				Team:    "my-team",
				Role:    client.ClusterMemberRoleAdmin,
			},
			wantErr: errSubjectRequired,
		},
		{
			name: "both user-id and team-id specified returns error",
			options: GrantOptions{
				Cluster: "my-cluster",
				UserID:  "user-123",
				TeamID:  "team-123",
				Role:    client.ClusterMemberRoleAdmin,
			},
			wantErr: errSubjectRequired,
		},
		{
			name: "user UPN provided is valid",
			options: GrantOptions{
				Cluster: "my-cluster",
				User:    "user@example.com",
				Role:    client.ClusterMemberRoleAdmin,
			},
			wantErr: nil,
		},
		{
			name: "user-id provided is valid",
			options: GrantOptions{
				Cluster: "my-cluster",
				UserID:  "user-123",
				Role:    client.ClusterMemberRoleAdmin,
			},
			wantErr: nil,
		},
		{
			name: "team name provided is valid",
			options: GrantOptions{
				Cluster: "my-cluster",
				Team:    "my-team",
				Role:    client.ClusterMemberRoleAdmin,
			},
			wantErr: nil,
		},
		{
			name: "team-id provided is valid",
			options: GrantOptions{
				Cluster: "my-cluster",
				TeamID:  "team-123",
				Role:    client.ClusterMemberRoleAdmin,
			},
			wantErr: nil,
		},

		// Role validation
		{
			name: "missing role returns error",
			options: GrantOptions{
				Cluster: "my-cluster",
				User:    "user@example.com",
				Role:    "",
			},
			wantErr: errRoleRequired,
		},
		{
			name: "admin role is valid",
			options: GrantOptions{
				Cluster: "my-cluster",
				User:    "user@example.com",
				Role:    client.ClusterMemberRoleAdmin,
			},
			wantErr: nil,
		},
		{
			name: "reader role is valid",
			options: GrantOptions{
				Cluster: "my-cluster",
				User:    "user@example.com",
				Role:    client.ClusterMemberRoleReader,
			},
			wantErr: nil,
		},
		{
			name: "invalid role returns error",
			options: GrantOptions{
				Cluster: "my-cluster",
				User:    "user@example.com",
				Role:    "owner",
			},
			wantErr: errInvalidClusterRole,
		},
		{
			name: "another invalid role returns error",
			options: GrantOptions{
				Cluster: "my-cluster",
				User:    "user@example.com",
				Role:    "member",
			},
			wantErr: errInvalidClusterRole,
		},

		// Complete valid options
		{
			name: "all valid options with cluster name and user UPN",
			options: GrantOptions{
				Cluster: "production-cluster",
				User:    "alice@example.com",
				Role:    client.ClusterMemberRoleAdmin,
			},
			wantErr: nil,
		},
		{
			name: "all valid options with IDs",
			options: GrantOptions{
				ClusterID: "cluster-abc-123",
				UserID:    "user-xyz-789",
				Role:      client.ClusterMemberRoleReader,
			},
			wantErr: nil,
		},
		{
			name: "all valid options with team",
			options: GrantOptions{
				Cluster: "production-cluster",
				Team:    "platform-team",
				Role:    client.ClusterMemberRoleReader,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGrantOptions(tt.options)

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}
