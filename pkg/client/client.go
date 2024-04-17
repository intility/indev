package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/intility/minctl/pkg/authenticator"
)

const (
	BaseURIDev         = "http://localhost:8080"
	BaseURI            = "https://containerplatform.intility.com"
	defaultHTTPTimeout = 10 * time.Second
	defaultAuthTimeout = 5 * time.Minute
)

type ClusterClient interface {
	ListClusters(ctx context.Context) (ClusterList, error)
	GetCluster(ctx context.Context, name string) (*Cluster, error)
	CreateCluster(ctx context.Context, request NewClusterRequest) error
	DeleteCluster(ctx context.Context, name string) error
}

type Client interface {
	ClusterClient
}

type RestClientOption func(*RestClient)

type RestClient struct {
	baseURI       string
	httpClient    *http.Client
	authenticator *authenticator.Authenticator
}

func New(options ...RestClientOption) *RestClient {
	client := &RestClient{
		baseURI:       BaseURI,
		httpClient:    &http.Client{Timeout: defaultHTTPTimeout},
		authenticator: authenticator.NewAuthenticator(authenticator.DefaultAuthConfig()),
	}

	for _, opt := range options {
		opt(client)
	}

	return client
}

func WithDevConfig() RestClientOption {
	return func(client *RestClient) {
		client.baseURI = BaseURIDev
	}
}

func WithAuthenticator(authenticator *authenticator.Authenticator) RestClientOption {
	return func(client *RestClient) {
		client.authenticator = authenticator
	}
}

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
		return clusters, err
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
	req, err := c.createAuthenticatedRequest(ctx, "GET", c.baseURI+"/api/v1/clusters/"+name, nil)
	if err != nil {
		return nil, err
	}

	var cluster Cluster
	if err = doRequest(c.httpClient, req, &cluster); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return &cluster, nil
}

func (c *RestClient) DeleteCluster(ctx context.Context, name string) error {
	req, err := c.createAuthenticatedRequest(ctx, "DELETE", c.baseURI+"/api/v1/clusters/"+name, nil)
	if err != nil {
		return err
	}

	if err = doRequest[any](c.httpClient, req, nil); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}
