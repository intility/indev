package access

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRevokeOptions(t *testing.T) {
	tests := []struct {
		name    string
		options RevokeOptions
		wantErr error
	}{
		// Cluster validation
		{
			name: "missing both cluster and cluster-id returns error",
			options: RevokeOptions{
				Cluster:   "",
				ClusterID: "",
				User:      "user@example.com",
			},
			wantErr: errClusterRequired,
		},
		{
			name: "cluster name provided is valid",
			options: RevokeOptions{
				Cluster: "my-cluster",
				User:    "user@example.com",
			},
			wantErr: nil,
		},
		{
			name: "cluster-id provided is valid",
			options: RevokeOptions{
				ClusterID: "cluster-123",
				User:      "user@example.com",
			},
			wantErr: nil,
		},
		{
			name: "both cluster and cluster-id provided is valid",
			options: RevokeOptions{
				Cluster:   "my-cluster",
				ClusterID: "cluster-123",
				User:      "user@example.com",
			},
			wantErr: nil,
		},

		// Subject validation (user OR team, exactly one)
		{
			name: "missing both user/user-id and team/team-id returns error",
			options: RevokeOptions{
				Cluster: "my-cluster",
				User:    "",
				UserID:  "",
				Team:    "",
				TeamID:  "",
			},
			wantErr: errSubjectRequired,
		},
		{
			name: "both user and team specified returns error",
			options: RevokeOptions{
				Cluster: "my-cluster",
				User:    "user@example.com",
				Team:    "my-team",
			},
			wantErr: errSubjectRequired,
		},
		{
			name: "both user-id and team-id specified returns error",
			options: RevokeOptions{
				Cluster: "my-cluster",
				UserID:  "user-123",
				TeamID:  "team-123",
			},
			wantErr: errSubjectRequired,
		},
		{
			name: "user UPN provided is valid",
			options: RevokeOptions{
				Cluster: "my-cluster",
				User:    "user@example.com",
			},
			wantErr: nil,
		},
		{
			name: "user-id provided is valid",
			options: RevokeOptions{
				Cluster: "my-cluster",
				UserID:  "user-123",
			},
			wantErr: nil,
		},
		{
			name: "team name provided is valid",
			options: RevokeOptions{
				Cluster: "my-cluster",
				Team:    "my-team",
			},
			wantErr: nil,
		},
		{
			name: "team-id provided is valid",
			options: RevokeOptions{
				Cluster: "my-cluster",
				TeamID:  "team-123",
			},
			wantErr: nil,
		},

		// Complete valid options
		{
			name: "all valid options with cluster name and user UPN",
			options: RevokeOptions{
				Cluster: "production-cluster",
				User:    "alice@example.com",
			},
			wantErr: nil,
		},
		{
			name: "all valid options with IDs",
			options: RevokeOptions{
				ClusterID: "cluster-abc-123",
				UserID:    "user-xyz-789",
			},
			wantErr: nil,
		},
		{
			name: "all valid options with team",
			options: RevokeOptions{
				Cluster: "production-cluster",
				Team:    "platform-team",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRevokeOptions(tt.options)

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}
