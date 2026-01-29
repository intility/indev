package access

import (
	"context"
	"io"
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

			return runGrantCommand(ctx, cmd.OutOrStdout(), set, &options)
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

func runGrantCommand(ctx context.Context, out io.Writer, set clientset.ClientSet, options *GrantOptions) error {
	if err := validateGrantOptions(*options); err != nil {
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

	err = set.PlatformClient.AddClusterMember(ctx, options.ClusterID, []client.AddClusterMemberRequest{
		{
			Subject: client.AddClusterMemberSubject{
				Type: subject.Type,
				ID:   subject.ID,
			},
			Roles: []client.ClusterMemberRole{options.Role},
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "409 Conflict") {
			return redact.Errorf("%s %s already has access to cluster %s", subject.Type, subject.Name, options.Cluster)
		}

		return redact.Errorf("could not grant cluster access: %w", redact.Safe(err))
	}

	clusterDisplay := options.Cluster
	if clusterDisplay == "" {
		clusterDisplay = options.ClusterID
	}

	subjectName := subject.Name
	if subjectName == "" {
		subjectName = subject.ID
	}

	ux.Fsuccessf(out, "Granted %s access to %s %s on cluster %s\n",
		options.Role, subject.Type, subjectName, clusterDisplay)

	return nil
}

//nolint:cyclop // validation logic is inherently sequential
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
