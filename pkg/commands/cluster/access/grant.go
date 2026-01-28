package access

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/redact"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
)

var (
	errClusterRequired    = redact.Errorf("cluster name or ID is required")
	errSubjectRequired    = redact.Errorf("either user or team must be specified, but not both")
	errRoleRequired       = redact.Errorf("role must be specified")
	errInvalidClusterRole = redact.Errorf(
		"invalid role supplied, valid roles are: \"%s\"",
		strings.Join(client.GetClusterMemberRoleValues(), ", "),
	)
)

type GrantOptions struct {
	Cluster   string
	ClusterID string
	User      string
	UserID    string
	Team      string
	TeamID    string
	Role      client.ClusterMemberRole
}

func NewGrantCommand(set clientset.ClientSet) *cobra.Command {
	var options GrantOptions

	cmd := &cobra.Command{
		Use:     "grant",
		Short:   "Grant access to a cluster",
		Long:    `Grant a user or team access to a cluster with a specific role.`,
		PreRunE: set.EnsureSignedInPreHook,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := telemetry.StartSpan(cmd.Context(), "cluster.access.grant")
			defer span.End()

			cmd.SilenceUsage = true

			if err := validateGrantOptions(options); err != nil {
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
			var (
				subjectType string
				subjectID   string
				subjectName string
			)

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

			err := set.PlatformClient.AddClusterMember(ctx, options.ClusterID, []client.AddClusterMemberRequest{
				{
					Subject: client.AddClusterMemberSubject{
						Type: subjectType,
						ID:   subjectID,
					},
					Roles: []client.ClusterMemberRole{options.Role},
				},
			})

			if err != nil {
				if strings.Contains(err.Error(), "409 Conflict") {
					return redact.Errorf("%s %s already has access to cluster %s", subjectType, subjectName, options.Cluster)
				}

				return redact.Errorf("could not grant cluster access: %w", redact.Safe(err))
			}

			clusterDisplay := options.Cluster
			if clusterDisplay == "" {
				clusterDisplay = options.ClusterID
			}

			if subjectName == "" {
				subjectName = subjectID
			}

			ux.Fsuccess(
				cmd.OutOrStdout(),
				"Granted %s access to %s %s on cluster %s\n",
				options.Role, subjectType, subjectName, clusterDisplay,
			)

			return nil
		},
	}

	cmd.Flags().StringVarP(&options.Cluster, "cluster", "c", "", "Name of the cluster")
	cmd.Flags().StringVar(&options.ClusterID, "cluster-id", "", "ID of the cluster")
	cmd.Flags().StringVarP(&options.User, "user", "u", "", "UPN of the user to grant access")
	cmd.Flags().StringVar(&options.UserID, "user-id", "", "ID of the user to grant access")
	cmd.Flags().StringVarP(&options.Team, "team", "t", "", "Name of the team to grant access")
	cmd.Flags().StringVar(&options.TeamID, "team-id", "", "ID of the team to grant access")

	roleFlagDescription := "Role to grant. Valid roles are: " +
		strings.Join(client.GetClusterMemberRoleValues(), ", ")
	cmd.Flags().StringVarP((*string)(&options.Role), "role", "r", "", roleFlagDescription)

	return cmd
}

func validateGrantOptions(options GrantOptions) error {
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

	// Validate role is specified
	if options.Role == "" {
		return errRoleRequired
	}

	// Validate role is valid
	if !options.Role.IsValid() {
		return errInvalidClusterRole
	}

	return nil
}
