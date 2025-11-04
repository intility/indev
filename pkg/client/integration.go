package client

import (
	"context"
	"fmt"
	"time"
)

type IntegrationInstance struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

func (c *RestClient) ListIntegrationInstances(ctx context.Context) ([]IntegrationInstance, error) {
	var instances []IntegrationInstance

	req, err := c.createAuthenticatedRequest(ctx, "GET", c.baseURI+"/api/v1/integrations/instances", nil)
	if err != nil {
		return instances, err
	}

	if err = doRequest(c.httpClient, req, &instances); err != nil {
		return instances, fmt.Errorf("request failed: %w", err)
	}

	return instances, nil
}
