package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type AIModel struct {
	ID            string `json:"id"            yaml:"id"`
	DisplayName   string `json:"displayName"   yaml:"displayName"`
	Slug          string `json:"slug"          yaml:"slug"`
	Description   string `json:"description"   yaml:"description"`
	ContextLength int    `json:"contextLength" yaml:"contextLength"`
}

type AIDeploymentCreatedBy struct {
	ID   string `json:"id"   yaml:"id"`
	Name string `json:"name" yaml:"name"`
	UPN  string `json:"upn"  yaml:"upn"`
}

type AIDeployment struct {
	ID        string                `json:"id"        yaml:"id"`
	Name      string                `json:"name"      yaml:"name"`
	Model     string                `json:"model"     yaml:"model"`
	Endpoint  string                `json:"endpoint"  yaml:"endpoint"`
	CreatedBy AIDeploymentCreatedBy `json:"createdBy" yaml:"createdBy"`
}

type NewAIDeploymentRequest struct {
	Name  string `json:"name"  validate:"required,min=3,max=63"`
	Model string `json:"model" validate:"required"`
}

func (c *RestClient) CreateAIDeployment(ctx context.Context, request NewAIDeploymentRequest) (*AIDeployment, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("could not marshal request: %w", err)
	}

	endpoint := c.baseURIBlurite + "/api/v1/blurite/llm-deployments"

	req, err := c.createAuthenticatedRequest(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var deploy AIDeployment
	if err = doRequest(c.httpClient, req, &deploy); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return &deploy, nil
}

func (c *RestClient) ListAIDeployments(ctx context.Context) ([]AIDeployment, error) {
	req, err := c.createAuthenticatedRequest(ctx, "GET", c.baseURIBlurite+"/api/v1/blurite/llm-deployments", nil)
	if err != nil {
		return nil, err
	}

	var deployments []AIDeployment
	if err = doRequest(c.httpClient, req, &deployments); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return deployments, nil
}

func (c *RestClient) GetAIDeployment(ctx context.Context, name string) (*AIDeployment, error) {
	endpoint := c.baseURIBlurite + "/api/v1/blurite/llm-deployments/by-name/" + name

	req, err := c.createAuthenticatedRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var deploy AIDeployment
	if err = doRequest(c.httpClient, req, &deploy); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return &deploy, nil
}

func (c *RestClient) DeleteAIDeployment(ctx context.Context, id string) error {
	req, err := c.createAuthenticatedRequest(ctx, "DELETE", c.baseURIBlurite+"/api/v1/blurite/llm-deployments/"+id, nil)
	if err != nil {
		return err
	}

	if err = doRequest[any](c.httpClient, req, nil); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}

func (c *RestClient) ListAIModels(ctx context.Context) ([]AIModel, error) {
	req, err := c.createAuthenticatedRequest(ctx, "GET", c.baseURIBlurite+"/api/v1/blurite/models", nil)
	if err != nil {
		return nil, err
	}

	var models []AIModel
	if err = doRequest(c.httpClient, req, &models); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return models, nil
}
