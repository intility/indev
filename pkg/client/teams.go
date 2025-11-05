package client

import (
	"context"
	"fmt"
)

type Team struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Role        []string `json:"roles"`
}

func (c *RestClient) ListTeams(ctx context.Context) ([]Team, error) {
	var teams []Team

	req, err := c.createAuthenticatedRequest(ctx, "GET", c.baseURI+"/api/v1/teams", nil)
	if err != nil {
		return teams, err
	}

	if err = doRequest(c.httpClient, req, &teams); err != nil {
		return teams, fmt.Errorf("request failed: %w", err)
	}

	return teams, nil
}
