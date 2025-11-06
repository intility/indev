package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Team struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Role        []string `json:"roles"`
}

type Subject struct {
	Type    string    `json:"type"`
	Name    string    `json:"name"`
	Details string    `json:"details"`
	ID      uuid.UUID `json:"id"`
}

type MemberRole string

const (
	MemberRoleOwner  MemberRole = "owner"
	MemberRoleMember MemberRole = "member"
)

type TeamMember struct {
	Subject Subject      `json:"subject"`
	Roles   []MemberRole `json:"roles"`
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

func (c *RestClient) GetTeamMembers(ctx context.Context, teamId string) ([]TeamMember, error) {
	var members []TeamMember

	req, err := c.createAuthenticatedRequest(ctx, "GET", c.baseURI+"/api/v1/teams/"+teamId+"/members", nil)
	if err != nil {
		return members, err
	}

	if err = doRequest(c.httpClient, req, &members); err != nil {
		return members, fmt.Errorf("request failed: %w", err)
	}

	return members, nil
}
