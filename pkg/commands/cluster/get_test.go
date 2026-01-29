package cluster

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/intility/indev/pkg/client"
)

func TestPrintClusterDetails(t *testing.T) {
	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name    string
		cluster *client.Cluster
		check   func(t *testing.T, output string)
	}{
		{
			name: "prints basic cluster information",
			cluster: &client.Cluster{
				ID:         "cluster-abc-123",
				Name:       "my-production-cluster",
				Version:    "4.14.5",
				ConsoleURL: "https://console.example.com",
				Status: client.ClusterStatus{
					Ready: client.StatusReady{Status: true},
				},
			},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "cluster-abc-123")
				assert.Contains(t, output, "my-production-cluster")
				assert.Contains(t, output, "4.14.5")
				assert.Contains(t, output, "https://console.example.com")
				assert.Contains(t, output, "Ready")
			},
		},
		{
			name: "prints roles when present",
			cluster: &client.Cluster{
				ID:      "cluster-1",
				Name:    "test-cluster",
				Version: "4.14",
				Roles:   []string{"admin", "developer"},
				Status: client.ClusterStatus{
					Ready: client.StatusReady{Status: true},
				},
			},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "Roles:")
				assert.Contains(t, output, "admin")
				assert.Contains(t, output, "developer")
			},
		},
		{
			name: "does not print roles section when empty",
			cluster: &client.Cluster{
				ID:      "cluster-1",
				Name:    "test-cluster",
				Version: "4.14",
				Roles:   []string{},
				Status: client.ClusterStatus{
					Ready: client.StatusReady{Status: true},
				},
			},
			check: func(t *testing.T, output string) {
				// Should not have Roles line when empty
				assert.NotContains(t, output, "Roles:")
			},
		},
		{
			name: "prints deployment status",
			cluster: &client.Cluster{
				ID:      "cluster-1",
				Name:    "deploying-cluster",
				Version: "4.14",
				Status: client.ClusterStatus{
					Ready:      client.StatusReady{Status: false},
					Deployment: client.StatusDeployment{Active: true},
				},
			},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "In Deployment")
			},
		},
		{
			name: "prints failed deployment status",
			cluster: &client.Cluster{
				ID:      "cluster-1",
				Name:    "failed-cluster",
				Version: "4.14",
				Status: client.ClusterStatus{
					Ready:      client.StatusReady{Status: false},
					Deployment: client.StatusDeployment{Failed: true},
				},
			},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "Deployment Failed")
			},
		},
		{
			name: "prints message and reason when present",
			cluster: &client.Cluster{
				ID:      "cluster-1",
				Name:    "test-cluster",
				Version: "4.14",
				Status: client.ClusterStatus{
					Ready: client.StatusReady{
						Status:  false,
						Message: "Waiting for nodes",
						Reason:  "NodesNotReady",
					},
				},
			},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "Waiting for nodes")
				assert.Contains(t, output, "NodesNotReady")
			},
		},
		{
			name: "prints node pool details",
			cluster: &client.Cluster{
				ID:      "cluster-1",
				Name:    "test-cluster",
				Version: "4.14",
				NodePools: []client.NodePool{
					{
						ID:       "pool-1",
						Name:     "worker-pool",
						Preset:   "balanced",
						Replicas: intPtr(3),
						Compute: &client.ComputeResources{
							Cores:  4,
							Memory: "16Gi",
						},
						AutoscalingEnabled: false,
					},
				},
				Status: client.ClusterStatus{
					Ready: client.StatusReady{Status: true},
				},
			},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "Node Pools:")
				assert.Contains(t, output, "Pool 1:")
				assert.Contains(t, output, "worker-pool")
				assert.Contains(t, output, "pool-1")
				assert.Contains(t, output, "balanced")
				assert.Contains(t, output, "Replicas:")
				assert.Contains(t, output, "3")
				assert.Contains(t, output, "Cores:")
				assert.Contains(t, output, "4")
				assert.Contains(t, output, "Memory:")
				assert.Contains(t, output, "16Gi")
			},
		},
		{
			name: "prints autoscaling details when enabled",
			cluster: &client.Cluster{
				ID:      "cluster-1",
				Name:    "test-cluster",
				Version: "4.14",
				NodePools: []client.NodePool{
					{
						Name:               "autoscale-pool",
						Preset:             "minimal",
						AutoscalingEnabled: true,
						MinCount:           intPtr(2),
						MaxCount:           intPtr(8),
					},
				},
				Status: client.ClusterStatus{
					Ready: client.StatusReady{Status: true},
				},
			},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "Autoscaling:")
				assert.Contains(t, output, "true")
				assert.Contains(t, output, "Min Count:")
				assert.Contains(t, output, "2")
				assert.Contains(t, output, "Max Count:")
				assert.Contains(t, output, "8")
			},
		},
		{
			name: "prints multiple node pools",
			cluster: &client.Cluster{
				ID:      "cluster-1",
				Name:    "multi-pool-cluster",
				Version: "4.14",
				NodePools: []client.NodePool{
					{Name: "pool-a", Preset: "minimal", Replicas: intPtr(2)},
					{Name: "pool-b", Preset: "performance", Replicas: intPtr(4)},
				},
				Status: client.ClusterStatus{
					Ready: client.StatusReady{Status: true},
				},
			},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "Pool 1:")
				assert.Contains(t, output, "pool-a")
				assert.Contains(t, output, "Pool 2:")
				assert.Contains(t, output, "pool-b")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printClusterDetails(&buf, tt.cluster)
			tt.check(t, buf.String())
		})
	}
}
