package client

import (
	"bytes"
	"context"
	"encoding/json"
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

// String returns the string representation of MemberRole.
func (mr MemberRole) String() string {
	return string(mr)
}

// IsValid checks if the MemberRole is valid.
func (mr MemberRole) IsValid() bool {
	switch mr {
	case MemberRoleOwner, MemberRoleMember:
		return true
	default:
		return false
	}
}

// GetMemberRoleValues returns a slice of valid MemberRole values.
func GetMemberRoleValues() []string {
	return []string{
		string(MemberRoleOwner),
		string(MemberRoleMember),
	}
}

type TeamMember struct {
	Subject Subject      `json:"subject"`
	Roles   []MemberRole `json:"roles"`
}

type NewTeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type DeleteTeamRequest struct {
	TeamId string `json:"teamId"`
}

type AddMemberSubject struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type AddTeamMemberRequest struct {
	Roles   []MemberRole     `json:"roles"`
	Subject AddMemberSubject `json:"subject"`
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

func (c *RestClient) CreateTeam(ctx context.Context, request NewTeamRequest) (*Team, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("could not marshal request: %w", err)
	}

	req, err := c.createAuthenticatedRequest(ctx, "POST", c.baseURI+"/api/v1/teams", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var team Team
	if err = doRequest(c.httpClient, req, &team); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return &team, nil
}

func (c *RestClient) DeleteTeam(ctx context.Context, request DeleteTeamRequest) error {
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("could not marshal request: %w", err)
	}

	req, err := c.createAuthenticatedRequest(ctx, "DELETE", c.baseURI+"/api/v1/teams/"+request.TeamId, bytes.NewReader(body))
	if err != nil {
		return err
	}

	if err = doRequest[any](c.httpClient, req, nil); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}

func (c *RestClient) AddTeamMember(ctx context.Context, teamId string, request []AddTeamMemberRequest) error {
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("could not marshal request: %w", err)
	}

	req, err := c.createAuthenticatedRequest(ctx, "POST", c.baseURI+"/api/v1/teams/"+teamId+"/members", bytes.NewReader(body))
	if err != nil {
		return err
	}

	if err = doRequest[any](c.httpClient, req, nil); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}

func (c *RestClient) RemoveTeamMember(ctx context.Context, teamId string, memberId string) error {
	req, err := c.createAuthenticatedRequest(ctx, "DELETE", c.baseURI+"/api/v1/teams/"+teamId+"/members/"+memberId, nil)
	if err != nil {
		return err
	}

	if err = doRequest[any](c.httpClient, req, nil); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}
