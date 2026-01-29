package member

import (
	"context"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/pkg/clientset"
)

func getTeamIDByName(ctx context.Context, set clientset.ClientSet, teamName string) (string, error) {
	team, err := set.PlatformClient.GetTeam(ctx, teamName)
	if err != nil {
		return "", redact.Errorf("could not get team: %w", redact.Safe(err))
	}

	if team == nil {
		return "", redact.Errorf("team not found: %s", teamName)
	}

	return team.ID, nil
}

func getUserIDByUpn(ctx context.Context, set clientset.ClientSet, upn string) (string, error) {
	user, err := set.PlatformClient.GetUser(ctx, upn)
	if err != nil {
		return "", redact.Errorf("could not get user: %w", redact.Safe(err))
	}

	if user == nil {
		return "", redact.Errorf("user not found: %s", upn)
	}

	return user.ID, nil
}
