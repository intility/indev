package client

type PullSecret struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	CreatedAt  string   `json:"createdAt"`
	Registries []string `json:"registries"`
}

type PullSecretCredential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type NewPullSecretRequest struct {
	Name       string                          `json:"name"`
	Registries map[string]PullSecretCredential `json:"registries"`
}

type EditPullSecretRequest struct {
	Registries map[string]*PullSecretCredential `json:"registries"`
}

type ImagePullSecret struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Registries []string `json:"registries"`
}
