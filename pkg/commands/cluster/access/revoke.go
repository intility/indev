package access

import (
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

			cmd.SilenceUsage = true

			if err := validateRevokeOptions(options); err != nil {
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
			var subjectType string
			var subjectID string
			var subjectName string

			if options.User != "" || options.UserID != "" {
				subjectType = "user"
				subjectName = options.User
				if options.UserID != "" {
					subjectID = options.UserID
				} else {
					userID, err := getUserIDByUPN(ctx, set, options.User)
					if err != nil {
						return err
					}
					subjectID = userID
				}
			} else {
				subjectType = "team"
				subjectName = options.Team
				if options.TeamID != "" {
					subjectID = options.TeamID
				} else {
					teamID, err := getTeamIDByName(ctx, set, options.Team)
					if err != nil {
						return err
					}
					subjectID = teamID
				}
			}

			memberID := subjectType + ":" + subjectID

			err := set.PlatformClient.RemoveClusterMember(ctx, options.ClusterID, memberID)
			if err != nil {
				if strings.Contains(err.Error(), "404 Not Found") {
					return redact.Errorf("%s %s does not have access to cluster %s", subjectType, subjectName, options.Cluster)
				}
				return redact.Errorf("could not revoke cluster access: %w", redact.Safe(err))
			}

			clusterDisplay := options.Cluster
			if clusterDisplay == "" {
				clusterDisplay = options.ClusterID
			}

			if subjectName == "" {
				subjectName = subjectID
			}

			ux.Fsuccess(cmd.OutOrStdout(), "Revoked access for %s %s from cluster %s\n", subjectType, subjectName, clusterDisplay)

			return nil
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
