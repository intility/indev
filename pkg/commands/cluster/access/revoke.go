package access

import (
	"context"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/clientset"
)

type RevokeOptions struct {
	Cluster   string
	ClusterID string
	User      string
	UserID    string
	Team      string
	TeamID    string
}

func NewRevokeCommand(set clientset.ClientSet) *cobra.Command {
	var options RevokeOptions

	cmd := &cobra.Command{
		Use:     "revoke",
		Short:   "Revoke access from a cluster",
		Long:    `Revoke a user's or team's access from a cluster.`,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "cluster.access.revoke")
			defer span.End()

			return runRevokeCommand(ctx, cmd.OutOrStdout(), set, &options)
		},
	}

	cmd.Flags().StringVarP(&options.Cluster, "cluster", "c", "", "Name of the cluster")
	cmd.Flags().StringVar(&options.ClusterID, "cluster-id", "", "ID of the cluster")
	cmd.Flags().StringVarP(&options.User, "user", "u", "", "UPN of the user to revoke access")
	cmd.Flags().StringVar(&options.UserID, "user-id", "", "ID of the user to revoke access")
	cmd.Flags().StringVarP(&options.Team, "team", "t", "", "Name of the team to revoke access")
	cmd.Flags().StringVar(&options.TeamID, "team-id", "", "ID of the team to revoke access")

	return cmd
}

func runRevokeCommand(ctx context.Context, out io.Writer, set clientset.ClientSet, options *RevokeOptions) error {
	if err := validateRevokeOptions(*options); err != nil {
		return err
	}

	// Resolve cluster ID
	if options.ClusterID == "" {
		clusterID, err := resolveClusterID(ctx, set, options.Cluster, options.ClusterID)
		if err != nil {
			return err
		}

		options.ClusterID = clusterID
	}

	// Determine subject type and resolve ID
	subject, err := resolveSubject(ctx, set, SubjectOptions{
		User:   options.User,
		UserID: options.UserID,
		Team:   options.Team,
		TeamID: options.TeamID,
	})
	if err != nil {
		return err
	}

	memberID := subject.Type + ":" + subject.ID

	err = set.PlatformClient.RemoveClusterMember(ctx, options.ClusterID, memberID)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return redact.Errorf("%s %s does not have access to cluster %s", subject.Type, subject.Name, options.Cluster)
		}

		return redact.Errorf("could not revoke cluster access: %w", redact.Safe(err))
	}

	clusterDisplay := options.Cluster
	if clusterDisplay == "" {
		clusterDisplay = options.ClusterID
	}

	subjectName := subject.Name
	if subjectName == "" {
		subjectName = subject.ID
	}

	ux.Fsuccessf(out, "Revoked access for %s %s from cluster %s\n",
		subject.Type, subjectName, clusterDisplay)

	return nil
}

func validateRevokeOptions(options RevokeOptions) error {
	// Validate cluster is specified
	if options.ClusterID == "" && options.Cluster == "" {
		return errClusterRequired
	}

	// Validate exactly one subject type is specified
	hasUser := options.User != "" || options.UserID != ""
	hasTeam := options.Team != "" || options.TeamID != ""

	if (!hasUser && !hasTeam) || (hasUser && hasTeam) {
		return errSubjectRequired
	}

	return nil
}
