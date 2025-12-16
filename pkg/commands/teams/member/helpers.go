package member

import (
	"context"
	"strings"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

func getTeamIdByName(ctx context.Context, set clientset.ClientSet, teamName string) (string, error) {
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

func getUserIdByUpn(ctx context.Context, set clientset.ClientSet, upn string) (string, error) {
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
