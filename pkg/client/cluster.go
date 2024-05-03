package client

type Cluster struct {
	Name       string        `json:"name"`
	ConsoleURL string        `json:"consoleUrl"`
	Status     ClusterStatus `json:"status"`
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
	Name string `json:"name"`
}
