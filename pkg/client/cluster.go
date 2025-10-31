package client

type Cluster struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Version    string        `json:"version"`
	ConsoleURL string        `json:"consoleUrl"`
	NodePools  NodePools     `json:"nodePools"`
	Status     ClusterStatus `json:"status"`
	Roles      []string      `json:"roles"`
}

type ClusterStatus struct {
	Ready      StatusReady      `json:"ready"`
	Deployment StatusDeployment `json:"deployment"`
}

type StatusReady struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Reason  string `json:"reason"`
}

type StatusDeployment struct {
	Active bool `json:"active"`
	Failed bool `json:"failed"`
}

type ClusterList []Cluster

type NewClusterRequest struct {
	Name           string    `json:"name"`
	SSOProvisioner string    `json:"ssoProvisioner"`
	NodePools      NodePools `json:"nodepools,omitempty"`
	Version        string    `json:"version,omitempty"`
	Environment    string    `json:"environment,omitempty"`
	PullSecretRef  *string   `json:"pullSecretRef,omitempty"`
}

type NodePools []NodePool

type NodePool struct {
	ID                 string            `json:"id,omitempty"`
	Name               string            `json:"name,omitempty"`
	Preset             string            `json:"preset,omitempty"`
	Replicas           *int              `json:"replicas,omitempty"`
	Compute            *ComputeResources `json:"compute,omitempty"`
	AutoscalingEnabled bool              `json:"autoscalingEnabled,omitempty"`
	MinCount           *int              `json:"minCount,omitempty"`
	MaxCount           *int              `json:"maxCount,omitempty"`
}

type ComputeResources struct {
	Cores  int    `json:"cores"`
	Memory string `json:"memory"`
}
