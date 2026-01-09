package cluster

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/outputformat"
)

func TestStatusString(t *testing.T) {
	tests := []struct {
		name    string
		cluster client.Cluster
		want    string
	}{
		{
			name: "ready cluster returns Ready",
			cluster: client.Cluster{
				Status: client.ClusterStatus{
					Ready: client.StatusReady{Status: true},
				},
			},
			want: "Ready",
		},
		{
			name: "active deployment returns In Deployment",
			cluster: client.Cluster{
				Status: client.ClusterStatus{
					Ready:      client.StatusReady{Status: false},
					Deployment: client.StatusDeployment{Active: true},
				},
			},
			want: "In Deployment",
		},
		{
			name: "failed deployment returns Deployment Failed",
			cluster: client.Cluster{
				Status: client.ClusterStatus{
					Ready:      client.StatusReady{Status: false},
					Deployment: client.StatusDeployment{Failed: true},
				},
			},
			want: "Deployment Failed",
		},
		{
			name: "not ready and not deploying returns Not Ready",
			cluster: client.Cluster{
				Status: client.ClusterStatus{
					Ready:      client.StatusReady{Status: false},
					Deployment: client.StatusDeployment{Active: false, Failed: false},
				},
			},
			want: "Not Ready",
		},
		{
			name: "ready takes precedence over deployment active",
			cluster: client.Cluster{
				Status: client.ClusterStatus{
					Ready:      client.StatusReady{Status: true},
					Deployment: client.StatusDeployment{Active: true},
				},
			},
			want: "Ready",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statusString(tt.cluster)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStatusMessage(t *testing.T) {
	tests := []struct {
		name    string
		cluster client.Cluster
		want    string
	}{
		{
			name: "returns status message",
			cluster: client.Cluster{
				Status: client.ClusterStatus{
					Ready: client.StatusReady{Message: "All nodes healthy"},
				},
			},
			want: "All nodes healthy",
		},
		{
			name: "returns empty string when no message",
			cluster: client.Cluster{
				Status: client.ClusterStatus{
					Ready: client.StatusReady{Message: ""},
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statusMessage(tt.cluster)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNodePoolSummary(t *testing.T) {
	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name    string
		cluster client.Cluster
		want    string
	}{
		{
			name: "no node pools returns 0",
			cluster: client.Cluster{
				NodePools: []client.NodePool{},
			},
			want: "0",
		},
		{
			name: "single pool with replicas",
			cluster: client.Cluster{
				NodePools: []client.NodePool{
					{Replicas: intPtr(3)},
				},
			},
			want: "1 pool(s), 3 node(s)",
		},
		{
			name: "multiple pools with replicas",
			cluster: client.Cluster{
				NodePools: []client.NodePool{
					{Replicas: intPtr(3)},
					{Replicas: intPtr(5)},
				},
			},
			want: "2 pool(s), 8 node(s)",
		},
		{
			name: "pool without replicas (nil)",
			cluster: client.Cluster{
				NodePools: []client.NodePool{
					{Replicas: nil},
				},
			},
			want: "1 pool(s), 0 node(s)",
		},
		{
			name: "mixed pools with and without replicas",
			cluster: client.Cluster{
				NodePools: []client.NodePool{
					{Replicas: intPtr(2)},
					{Replicas: nil},
					{Replicas: intPtr(4)},
				},
			},
			want: "3 pool(s), 6 node(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nodePoolSummary(tt.cluster)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRolesString(t *testing.T) {
	tests := []struct {
		name    string
		cluster client.Cluster
		want    string
	}{
		{
			name: "no roles returns empty string",
			cluster: client.Cluster{
				Roles: []string{},
			},
			want: "",
		},
		{
			name: "single role",
			cluster: client.Cluster{
				Roles: []string{"admin"},
			},
			want: "[admin]",
		},
		{
			name: "multiple roles",
			cluster: client.Cluster{
				Roles: []string{"admin", "developer"},
			},
			want: "[admin developer]",
		},
		{
			name: "nil roles returns empty string",
			cluster: client.Cluster{
				Roles: nil,
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rolesString(tt.cluster)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrintClusterList(t *testing.T) {
	intPtr := func(i int) *int { return &i }

	sampleClusters := client.ClusterList{
		{
			ID:         "cluster-1",
			Name:       "prod-cluster",
			Version:    "4.14",
			ConsoleURL: "https://console.prod.example.com",
			NodePools: []client.NodePool{
				{Replicas: intPtr(3)},
			},
			Status: client.ClusterStatus{
				Ready: client.StatusReady{Status: true},
			},
			Roles: []string{"admin"},
		},
		{
			ID:         "cluster-2",
			Name:       "dev-cluster",
			Version:    "4.13",
			ConsoleURL: "https://console.dev.example.com",
			NodePools: []client.NodePool{
				{Replicas: intPtr(2)},
			},
			Status: client.ClusterStatus{
				Ready:      client.StatusReady{Status: false},
				Deployment: client.StatusDeployment{Active: true},
			},
		},
	}

	t.Run("default format shows name, version, status, node pools", func(t *testing.T) {
		var buf bytes.Buffer
		err := printClusterList(&buf, outputformat.Format(""), sampleClusters)

		assert.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "prod-cluster")
		assert.Contains(t, output, "dev-cluster")
		assert.Contains(t, output, "4.14")
		assert.Contains(t, output, "Ready")
		// Default format should NOT contain Console URL
		assert.NotContains(t, output, "console.prod.example.com")
	})

	t.Run("wide format includes console URL and roles", func(t *testing.T) {
		var buf bytes.Buffer
		err := printClusterList(&buf, outputformat.Format("wide"), sampleClusters)

		assert.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "prod-cluster")
		assert.Contains(t, output, "console.prod.example.com")
		assert.Contains(t, output, "[admin]")
	})

	t.Run("json format outputs valid JSON", func(t *testing.T) {
		var buf bytes.Buffer
		err := printClusterList(&buf, outputformat.Format("json"), sampleClusters)

		assert.NoError(t, err)

		var decoded client.ClusterList
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Len(t, decoded, 2)
		assert.Equal(t, "prod-cluster", decoded[0].Name)
	})

	t.Run("yaml format outputs valid YAML", func(t *testing.T) {
		var buf bytes.Buffer
		err := printClusterList(&buf, outputformat.Format("yaml"), sampleClusters)

		assert.NoError(t, err)

		var decoded client.ClusterList
		err = yaml.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Len(t, decoded, 2)
	})

	t.Run("handles empty cluster list", func(t *testing.T) {
		var buf bytes.Buffer
		err := printClusterList(&buf, outputformat.Format("json"), client.ClusterList{})

		assert.NoError(t, err)

		var decoded client.ClusterList
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Len(t, decoded, 0)
	})
}
