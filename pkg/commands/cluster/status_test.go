package cluster

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/intility/indev/pkg/client"
)

func TestPrintClusterStatus(t *testing.T) {
	tests := []struct {
		name    string
		cluster *client.Cluster
		want    string
	}{
		{
			name: "ready cluster prints Ready",
			cluster: &client.Cluster{
				Status: client.ClusterStatus{
					Ready: client.StatusReady{Status: true},
				},
			},
			want: "Ready\n",
		},
		{
			name: "active deployment prints In Deployment",
			cluster: &client.Cluster{
				Status: client.ClusterStatus{
					Ready:      client.StatusReady{Status: false},
					Deployment: client.StatusDeployment{Active: true},
				},
			},
			want: "In Deployment\n",
		},
		{
			name: "failed deployment prints Deployment Failed",
			cluster: &client.Cluster{
				Status: client.ClusterStatus{
					Ready:      client.StatusReady{Status: false},
					Deployment: client.StatusDeployment{Failed: true},
				},
			},
			want: "Deployment Failed\n",
		},
		{
			name: "not ready prints Not Ready",
			cluster: &client.Cluster{
				Status: client.ClusterStatus{
					Ready:      client.StatusReady{Status: false},
					Deployment: client.StatusDeployment{Active: false, Failed: false},
				},
			},
			want: "Not Ready\n",
		},
		{
			name: "ready takes precedence over active deployment",
			cluster: &client.Cluster{
				Status: client.ClusterStatus{
					Ready:      client.StatusReady{Status: true},
					Deployment: client.StatusDeployment{Active: true},
				},
			},
			want: "Ready\n",
		},
		{
			name: "active deployment takes precedence over failed",
			cluster: &client.Cluster{
				Status: client.ClusterStatus{
					Ready:      client.StatusReady{Status: false},
					Deployment: client.StatusDeployment{Active: true, Failed: true},
				},
			},
			want: "In Deployment\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printClusterStatus(&buf, tt.cluster)

			assert.Equal(t, tt.want, buf.String())
		})
	}
}
