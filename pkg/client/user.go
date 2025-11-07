package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type User struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	UPN   string    `json:"upn"`
	Roles []string  `json:"roles"`
}

func (c *RestClient) ListUsers(ctx context.Context) ([]User, error) {
	var users []User

	req, err := c.createAuthenticatedRequest(ctx, "GET", c.baseURI+"/api/v1/users", nil)
	if err != nil {
		return users, err
	}

	if err = doRequest(c.httpClient, req, &users); err != nil {
		return users, fmt.Errorf("request failed: %w", err)
	}

	return users, nil
}
