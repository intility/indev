package client

import (
	"context"
	"errors"
	"fmt"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID    string   `json:"id"    yaml:"id"`
	Name  string   `json:"name"  yaml:"name"`
	UPN   string   `json:"upn"   yaml:"upn"`
	Roles []string `json:"roles" yaml:"roles"`
}

type UserList []User

func (c *RestClient) ListUsers(ctx context.Context) ([]User, error) {
	var users UserList

	req, err := c.createAuthenticatedRequest(ctx, "GET", c.baseURI+"/api/v1/users", nil)
	if err != nil {
		return users, err
	}

	if err = doRequest(c.httpClient, req, &users); err != nil {
		return users, fmt.Errorf("request failed: %w", err)
	}

	return users, nil
}

func (c *RestClient) GetUser(ctx context.Context, upn string) (*User, error) {
	req, err := c.createAuthenticatedRequest(ctx, "GET", c.baseURI+"/api/v1/users/by-upn/"+upn, nil)
	if err != nil {
		return nil, err
	}

	var user User
	if err = doRequest(c.httpClient, req, &user); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if user.UPN != upn {
		return nil, fmt.Errorf("%w: %s", ErrUserNotFound, upn)
	}

	return &user, nil
}
