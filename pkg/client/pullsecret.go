package client

type PullSecret struct {
	ID         string   `json:"id"         yaml:"id"`
	Name       string   `json:"name"       yaml:"name"`
	CreatedAt  string   `json:"createdAt"  yaml:"createdAt"`
	Registries []string `json:"registries" yaml:"registries"`
}

type PullSecretCredential struct {
	Username string `json:"username"`
	Password string `json:"password"` //nolint:gosec // G117: this is a credential payload, not a hardcoded secret
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
