package client

import "github.com/google/uuid"

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

type ClusterMemberRole string

const (
	ClusterMemberRoleAdmin  ClusterMemberRole = "admin"
	ClusterMemberRoleReader ClusterMemberRole = "reader"
)

type ClusterMemberSubject struct {
	Type    string    `json:"type"    yaml:"type"`
	Name    string    `json:"name"    yaml:"name"`
	Details string    `json:"details" yaml:"details"`
	ID      uuid.UUID `json:"id"      yaml:"id"`
}

type ClusterMember struct {
	Subject ClusterMemberSubject `json:"subject" yaml:"subject"`
	Roles   []ClusterMemberRole  `json:"roles"   yaml:"roles"`
}

// String returns the string representation of ClusterMemberRole.
func (r ClusterMemberRole) String() string {
	return string(r)
}

// IsValid checks if the ClusterMemberRole is valid.
func (r ClusterMemberRole) IsValid() bool {
	switch r {
	case ClusterMemberRoleAdmin, ClusterMemberRoleReader:
		return true
	default:
		return false
	}
}

// GetClusterMemberRoleValues returns a slice of valid ClusterMemberRole values.
func GetClusterMemberRoleValues() []string {
	return []string{
		string(ClusterMemberRoleAdmin),
		string(ClusterMemberRoleReader),
	}
}

type AddClusterMemberSubject struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type AddClusterMemberRequest struct {
	Subject AddClusterMemberSubject `json:"subject"`
	Roles   []ClusterMemberRole     `json:"roles"`
}

type AddClusterMembersPayload struct {
	Values []AddClusterMemberRequest `json:"values"`
}
