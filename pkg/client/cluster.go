package client

type Cluster struct {
	Name string
}

type ClusterList []Cluster

type NewClusterRequest struct {
	Name string `json:"name"`
}
