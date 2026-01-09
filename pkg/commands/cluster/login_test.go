package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAPIURL(t *testing.T) {
	tests := []struct {
		name        string
		clusterName string
		tenantID    string
		want        string
	}{
		{
			name:        "mustafar tenant uses zone-1 URL",
			clusterName: "my-cluster",
			tenantID:    mustafarTenantID,
			want:        "https://api-my-cluster.clusters.zone-1.xcv.net",
		},
		{
			name:        "other tenant uses intilitycloud URL",
			clusterName: "my-cluster",
			tenantID:    "other-tenant-id",
			want:        "https://api-my-cluster.apps.intilitycloud.com",
		},
		{
			name:        "empty tenant ID uses intilitycloud URL",
			clusterName: "test-cluster",
			tenantID:    "",
			want:        "https://api-test-cluster.apps.intilitycloud.com",
		},
		{
			name:        "cluster name with suffix",
			clusterName: "prod-cluster-abc123",
			tenantID:    "some-tenant",
			want:        "https://api-prod-cluster-abc123.apps.intilitycloud.com",
		},
		{
			name:        "mustafar with complex cluster name",
			clusterName: "my-production-cluster-xyz789",
			tenantID:    mustafarTenantID,
			want:        "https://api-my-production-cluster-xyz789.clusters.zone-1.xcv.net",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getAPIURL(tt.clusterName, tt.tenantID)
			assert.Equal(t, tt.want, got)
		})
	}
}
