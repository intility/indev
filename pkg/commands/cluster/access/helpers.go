package access

import (
	"context"
	"strings"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

func resolveClusterID(ctx context.Context, set clientset.ClientSet, clusterName, clusterID string) (string, error) {
	// If cluster ID is provided directly, use it
	if clusterID != "" {
		return clusterID, nil
	}

	// If no cluster name provided, return error
	if clusterName == "" {
		return "", redact.Errorf("cluster name or ID is required")
	}

	// List clusters to find the one by name
	clusters, err := set.PlatformClient.ListClusters(ctx)
	if err != nil {
		return "", redact.Errorf("could not list clusters: %w", redact.Safe(err))
	}

	// Find the cluster with the matching name
	for _, c := range clusters {
		if strings.EqualFold(c.Name, clusterName) {
			return c.ID, nil
		}
	}

	return "", redact.Errorf("cluster not found: %s", clusterName)
}

func getUserIDByUPN(ctx context.Context, set clientset.ClientSet, upn string) (string, error) {
	users, err := set.PlatformClient.ListUsers(ctx)
	if err != nil {
		return "", redact.Errorf("could not list users: %w", redact.Safe(err))
	}

	var user *client.User
	for _, u := range users {
		if strings.EqualFold(u.UPN, upn) {
			user = &u
			break
		}
	}

	if user == nil {
		return "", redact.Errorf("user not found: %s", upn)
	}

	return user.ID, nil
}

func getTeamIDByName(ctx context.Context, set clientset.ClientSet, teamName string) (string, error) {
	teams, err := set.PlatformClient.ListTeams(ctx)
	if err != nil {
		return "", redact.Errorf("could not list teams: %w", redact.Safe(err))
	}

	var team *client.Team
	for _, t := range teams {
		if strings.EqualFold(t.Name, teamName) {
			team = &t
			break
		}
	}

	if team == nil {
		return "", redact.Errorf("team not found: %s", teamName)
	}

	return team.ID, nil
}
