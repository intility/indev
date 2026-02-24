package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/intility/indev/internal/build"
	"github.com/intility/indev/pkg/authenticator"
)

var ErrClusterNotFound = errors.New("cluster not found")

const (
	defaultHTTPTimeout = 10 * time.Second
	defaultAuthTimeout = 5 * time.Minute
)

type ClusterClient interface {
	ListClusters(ctx context.Context) (ClusterList, error)
	GetCluster(ctx context.Context, name string) (*Cluster, error)
	GetClusterStatus(ctx context.Context, clusterID string) (*Cluster, error)
	CreateCluster(ctx context.Context, request NewClusterRequest) (*Cluster, error)
	DeleteCluster(ctx context.Context, name string) error
	GetClusterMembers(ctx context.Context, clusterID string) ([]ClusterMember, error)
	AddClusterMember(ctx context.Context, clusterID string, request []AddClusterMemberRequest) error
	RemoveClusterMember(ctx context.Context, clusterID string, memberID string) error
}

type IntegrationClient interface {
	ListIntegrationInstances(ctx context.Context) ([]IntegrationInstance, error)
}

type MeClient interface {
	GetMe(ctx context.Context) (Me, error)
}

type TeamsClient interface {
	ListTeams(ctx context.Context) ([]Team, error)
	GetTeam(ctx context.Context, name string) (*Team, error)
	GetTeamMembers(ctx context.Context, teamID string) ([]TeamMember, error)
	CreateTeam(ctx context.Context, request NewTeamRequest) (*Team, error)
	DeleteTeam(ctx context.Context, request DeleteTeamRequest) error
}

type MemberClient interface {
	AddTeamMember(ctx context.Context, teamID string, request []AddTeamMemberRequest) error
	RemoveTeamMember(ctx context.Context, teamID string, memberID string) error
}

type UserClient interface {
	ListUsers(ctx context.Context) ([]User, error)
	GetUser(ctx context.Context, upn string) (*User, error)
}

type PullSecretClient interface {
	ListPullSecrets(ctx context.Context) ([]PullSecret, error)
	GetPullSecret(ctx context.Context, id string) (*PullSecret, error)
	CreatePullSecret(ctx context.Context, request NewPullSecretRequest) (*PullSecret, error)
	EditPullSecret(ctx context.Context, id string, request EditPullSecretRequest) (*PullSecret, error)
	DeletePullSecret(ctx context.Context, id string) error
	SetClusterPullSecret(ctx context.Context, clusterID string, pullSecretID string) error
}

type Client interface {
	ClusterClient
	IntegrationClient
	MeClient
	TeamsClient
	UserClient
	MemberClient
	PullSecretClient
}

type RestClientOption func(*RestClient)

type RestClient struct {
	baseURI       string
	httpClient    *http.Client
	authenticator *authenticator.Authenticator
}

var _ Client = New()

func New(options ...RestClientOption) *RestClient {
	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   defaultHTTPTimeout,
	}
	restClient := &RestClient{
		baseURI:       build.PlatformAPIHost(),
		httpClient:    client,
		authenticator: authenticator.NewAuthenticator(authenticator.ConfigFromBuildProps()),
	}

	for _, opt := range options {
		opt(restClient)
	}

	return restClient
}

//goland:noinspection GoUnusedExportedFunction
func WithAuthenticator(authenticator *authenticator.Authenticator) RestClientOption {
	return func(client *RestClient) {
		client.authenticator = authenticator
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithHTTPClient(httpClient *http.Client) RestClientOption {
	return func(client *RestClient) {
		client.httpClient = httpClient
	}
}

func (c *RestClient) createAuthenticatedRequest(
	ctx context.Context,
	method,
	path string,
	body io.Reader,
) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, path, body)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	authContext, cancel := context.WithTimeout(ctx, defaultAuthTimeout)
	defer cancel()

	authResult, err := c.authenticator.Authenticate(authContext)
	if err != nil {
		return nil, fmt.Errorf("could not authenticate: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+authResult.AccessToken)

	return req, nil
}

func (c *RestClient) ListClusters(ctx context.Context) (ClusterList, error) {
	var clusters ClusterList

	req, err := c.createAuthenticatedRequest(ctx, "GET", c.baseURI+"/api/v1/clusters", nil)
	if err != nil {
		return clusters, err
	}

	if err = doRequest(c.httpClient, req, &clusters); err != nil {
		return clusters, fmt.Errorf("request failed: %w", err)
	}

	return clusters, nil
}

func (c *RestClient) CreateCluster(ctx context.Context, request NewClusterRequest) (*Cluster, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("could not marshal request: %w", err)
	}

	req, err := c.createAuthenticatedRequest(ctx, "POST", c.baseURI+"/api/v1/clusters", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var result Cluster
	if err = doRequest(c.httpClient, req, &result); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return &result, nil
}

func (c *RestClient) GetCluster(ctx context.Context, name string) (*Cluster, error) {
	req, err := c.createAuthenticatedRequest(ctx, "GET", c.baseURI+"/api/v1/clusters/by-name/"+name, nil)
	if err != nil {
		return nil, err
	}

	var cluster Cluster
	if err = doRequest(c.httpClient, req, &cluster); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if cluster.Name != name {
		return nil, fmt.Errorf("%w: %s", ErrClusterNotFound, name)
	}

	return &cluster, nil
}

func (c *RestClient) GetClusterStatus(ctx context.Context, clusterID string) (*Cluster, error) {
	req, err := c.createAuthenticatedRequest(ctx, "GET", c.baseURI+"/api/v1/clusters/"+clusterID+"/status", nil)
	if err != nil {
		return nil, err
	}

	var cluster Cluster
	if err = doRequest(c.httpClient, req, &cluster); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return &cluster, nil
}

func (c *RestClient) DeleteCluster(ctx context.Context, id string) error {
	req, err := c.createAuthenticatedRequest(ctx, "DELETE", c.baseURI+"/api/v1/clusters/"+id, nil)
	if err != nil {
		return err
	}

	if err = doRequest[any](c.httpClient, req, nil); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}

func (c *RestClient) GetClusterMembers(ctx context.Context, clusterID string) ([]ClusterMember, error) {
	var members []ClusterMember

	req, err := c.createAuthenticatedRequest(ctx, "GET", c.baseURI+"/api/v1/clusters/"+clusterID+"/members", nil)
	if err != nil {
		return members, err
	}

	if err = doRequest(c.httpClient, req, &members); err != nil {
		return members, fmt.Errorf("request failed: %w", err)
	}

	return members, nil
}

func (c *RestClient) AddClusterMember(ctx context.Context, clusterID string, request []AddClusterMemberRequest) error {
	payload := AddClusterMembersPayload{Values: request}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("could not marshal request: %w", err)
	}

	endpoint := c.baseURI + "/api/v1/clusters/" + clusterID + "/members"

	req, err := c.createAuthenticatedRequest(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}

	if err = doRequest[any](c.httpClient, req, nil); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}

func (c *RestClient) RemoveClusterMember(ctx context.Context, clusterID string, memberID string) error {
	endpoint := c.baseURI + "/api/v1/clusters/" + clusterID + "/members/" + memberID

	req, err := c.createAuthenticatedRequest(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}

	if err = doRequest[any](c.httpClient, req, nil); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}

func (c *RestClient) ListPullSecrets(ctx context.Context) ([]PullSecret, error) {
	var pullSecrets []PullSecret

	req, err := c.createAuthenticatedRequest(ctx, "GET", c.baseURI+"/api/v1/settings/image-pull-secrets", nil)
	if err != nil {
		return pullSecrets, err
	}

	if err = doRequest(c.httpClient, req, &pullSecrets); err != nil {
		return pullSecrets, fmt.Errorf("request failed: %w", err)
	}

	return pullSecrets, nil
}

func (c *RestClient) GetPullSecret(ctx context.Context, id string) (*PullSecret, error) {
	req, err := c.createAuthenticatedRequest(ctx, "GET", c.baseURI+"/api/v1/settings/image-pull-secrets/"+id, nil)
	if err != nil {
		return nil, err
	}

	var pullSecret PullSecret
	if err = doRequest(c.httpClient, req, &pullSecret); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return &pullSecret, nil
}

func (c *RestClient) CreatePullSecret(ctx context.Context, request NewPullSecretRequest) (*PullSecret, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("could not marshal request: %w", err)
	}

	endpoint := c.baseURI + "/api/v1/settings/image-pull-secrets"

	req, err := c.createAuthenticatedRequest(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var pullSecret PullSecret
	if err = doRequest(c.httpClient, req, &pullSecret); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return &pullSecret, nil
}

func (c *RestClient) EditPullSecret(
	ctx context.Context,
	id string,
	request EditPullSecretRequest,
) (*PullSecret, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("could not marshal request: %w", err)
	}

	endpoint := c.baseURI + "/api/v1/settings/image-pull-secrets/" + id

	req, err := c.createAuthenticatedRequest(ctx, "PATCH", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var pullSecret PullSecret
	if err = doRequest(c.httpClient, req, &pullSecret); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return &pullSecret, nil
}

func (c *RestClient) DeletePullSecret(ctx context.Context, id string) error {
	req, err := c.createAuthenticatedRequest(ctx, "DELETE", c.baseURI+"/api/v1/settings/image-pull-secrets/"+id, nil)
	if err != nil {
		return err
	}

	if err = doRequest[any](c.httpClient, req, nil); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}

func (c *RestClient) SetClusterPullSecret(ctx context.Context, clusterID string, pullSecretID string) error {
	endpoint := c.baseURI + "/api/v1/clusters/" + clusterID + "/image-pull-secrets/" + pullSecretID

	req, err := c.createAuthenticatedRequest(ctx, "PUT", endpoint, nil)
	if err != nil {
		return err
	}

	if err = doRequest[any](c.httpClient, req, nil); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}
