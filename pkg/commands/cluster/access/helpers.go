package access

import (
	"context"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/pkg/clientset"
)

// SubjectInfo contains resolved information about a user or team subject.
type SubjectInfo struct {
	Type string // "user" or "team"
	ID   string
	Name string
}

// SubjectOptions contains the user/team options from command flags.
type SubjectOptions struct {
	User   string
	UserID string
	Team   string
	TeamID string
}

// resolveSubject resolves the subject (user or team) from the provided options.
func resolveSubject(ctx context.Context, set clientset.ClientSet, opts SubjectOptions) (SubjectInfo, error) {
	if opts.User != "" || opts.UserID != "" {
		return resolveUserSubject(ctx, set, opts.User, opts.UserID)
	}

	return resolveTeamSubject(ctx, set, opts.Team, opts.TeamID)
}

func resolveUserSubject(ctx context.Context, set clientset.ClientSet, upn, userID string) (SubjectInfo, error) {
	info := SubjectInfo{Type: "user", ID: "", Name: upn}

	if userID != "" {
		info.ID = userID

		return info, nil
	}

	resolvedID, err := getUserIDByUPN(ctx, set, upn)
	if err != nil {
		return info, err
	}

	info.ID = resolvedID

	return info, nil
}

func resolveTeamSubject(ctx context.Context, set clientset.ClientSet, teamName, teamID string) (SubjectInfo, error) {
	info := SubjectInfo{Type: "team", ID: "", Name: teamName}

	if teamID != "" {
		info.ID = teamID

		return info, nil
	}

	resolvedID, err := getTeamIDByName(ctx, set, teamName)
	if err != nil {
		return info, err
	}

	info.ID = resolvedID

	return info, nil
}

func resolveClusterID(ctx context.Context, set clientset.ClientSet, clusterName, clusterID string) (string, error) {
	// If cluster ID is provided directly, use it
	if clusterID != "" {
		return clusterID, nil
	}

	// If no cluster name provided, return error
	if clusterName == "" {
		return "", redact.Errorf("cluster name or ID is required")
	}

	// Get cluster by name
	cluster, err := set.PlatformClient.GetCluster(ctx, clusterName)
	if err != nil {
		return "", redact.Errorf("could not get cluster: %w", redact.Safe(err))
	}

	if cluster == nil {
		return "", redact.Errorf("cluster not found: %s", clusterName)
	}

	return cluster.ID, nil
}

func getUserIDByUPN(ctx context.Context, set clientset.ClientSet, upn string) (string, error) {
	user, err := set.PlatformClient.GetUser(ctx, upn)
	if err != nil {
		return "", redact.Errorf("could not get user: %w", redact.Safe(err))
	}

	if user == nil {
		return "", redact.Errorf("user not found: %s", upn)
	}

	return user.ID, nil
}

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
