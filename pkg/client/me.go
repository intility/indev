package client

import (
	"context"
	"fmt"
)

type Me struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	OrganizationRoles []string `json:"organizationRoles"`
	OrganizationName  string   `json:"organizationName"`
}

func (c *RestClient) GetMe(ctx context.Context) (Me, error) {
	var me Me

	req, err := c.createAuthenticatedRequest(ctx, "GET", c.baseURI+"/api/v1/me", nil)
	if err != nil {
		return me, err
	}

	if err = doRequest(c.httpClient, req, &me); err != nil {
		return me, fmt.Errorf("request failed: %w", err)
	}

	return me, nil
}
