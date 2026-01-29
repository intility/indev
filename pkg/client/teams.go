package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var ErrTeamNotFound = errors.New("team not found")

type Team struct {
	ID          string   `json:"id"          yaml:"id"`
	Name        string   `json:"name"        yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	Role        []string `json:"roles"       yaml:"roles"`
}

type TeamList []Team

type Subject struct {
	Type    string    `json:"type"    yaml:"type"`
	Name    string    `json:"name"    yaml:"name"`
	Details string    `json:"details" yaml:"details"`
	ID      uuid.UUID `json:"id"      yaml:"id"`
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
	TeamID string `json:"teamId"`
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
	var teams TeamList

	req, err := c.createAuthenticatedRequest(ctx, "GET", c.baseURI+"/api/v1/teams", nil)
	if err != nil {
		return teams, err
	}

	if err = doRequest(c.httpClient, req, &teams); err != nil {
		return teams, fmt.Errorf("request failed: %w", err)
	}

	return teams, nil
}

func (c *RestClient) GetTeam(ctx context.Context, name string) (*Team, error) {
	req, err := c.createAuthenticatedRequest(ctx, "GET", c.baseURI+"/api/v1/teams/by-name/"+name, nil)
	if err != nil {
		return nil, err
	}

	var team Team
	if err = doRequest(c.httpClient, req, &team); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if team.Name != name {
		return nil, fmt.Errorf("%w: %s", ErrTeamNotFound, name)
	}

	return &team, nil
}

func (c *RestClient) GetTeamMembers(ctx context.Context, teamID string) ([]TeamMember, error) {
	var members []TeamMember

	req, err := c.createAuthenticatedRequest(ctx, "GET", c.baseURI+"/api/v1/teams/"+teamID+"/members", nil)
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
	req, err := c.createAuthenticatedRequest(ctx, "DELETE", c.baseURI+"/api/v1/teams/"+request.TeamID, nil)
	if err != nil {
		return err
	}

	if err = doRequest[any](c.httpClient, req, nil); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}

func (c *RestClient) AddTeamMember(ctx context.Context, teamID string, request []AddTeamMemberRequest) error {
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("could not marshal request: %w", err)
	}

	endpoint := c.baseURI + "/api/v1/teams/" + teamID + "/members"

	req, err := c.createAuthenticatedRequest(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}

	if err = doRequest[any](c.httpClient, req, nil); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}

func (c *RestClient) RemoveTeamMember(ctx context.Context, teamID string, memberID string) error {
	req, err := c.createAuthenticatedRequest(ctx, "DELETE", c.baseURI+"/api/v1/teams/"+teamID+"/members/"+memberID, nil)
	if err != nil {
		return err
	}

	if err = doRequest[any](c.httpClient, req, nil); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}
