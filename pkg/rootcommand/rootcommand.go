package rootcommand

import (
	"github.com/spf13/cobra"

	"github.com/intility/indev/internal/build"
	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/authenticator"
	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/clientset"
	"github.com/intility/indev/pkg/commands/account"
	"github.com/intility/indev/pkg/commands/cluster"
	"github.com/intility/indev/pkg/commands/teams"
	"github.com/intility/indev/pkg/commands/teams/member"
	"github.com/intility/indev/pkg/commands/user"
)

func GetRootCommand() *cobra.Command {
	clients := clientset.ClientSet{
		Authenticator:  authenticator.NewAuthenticator(authenticator.ConfigFromBuildProps()),
		PlatformClient: client.New(),
	}

	rootCmd := &cobra.Command{
		Use:           build.AppName,
		Short:         build.AppName + " controls your Intility Developer Platform instance.",
		Long:          ``,
		Run:           showHelp,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(getVersionCommand())
	rootCmd.AddCommand(account.NewLoginCommand(clients))
	rootCmd.AddCommand(account.NewLogoutCommand(clients))
	rootCmd.AddCommand(getClusterCommand(clients))
	rootCmd.AddCommand(getAccountCommand(clients))
	rootCmd.AddCommand(getTeamsCommand(clients))
	rootCmd.AddCommand(getUserCommand(clients))

	return rootCmd
}

func getVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  `Print the version information of indev.`,
		Run: func(cmd *cobra.Command, args []string) {
			_, span := telemetry.StartSpan(cmd.Context(), "version")
			defer span.End()

			ux.Fprint(cmd.OutOrStdout(), "%s\n", build.NameVersionString())
		},
	}
}

func getClusterCommand(set clientset.ClientSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage your Intility Developer Platform clusters",
		Long:  `Manage your Intility Developer Platform clusters`,
		Run:   showHelp,
	}

	cmd.AddCommand(cluster.NewCreateCommand(set))
	cmd.AddCommand(cluster.NewDeleteCommand(set))
	cmd.AddCommand(cluster.NewGetCommand(set))
	cmd.AddCommand(cluster.NewListCommand(set))
	cmd.AddCommand(cluster.NewStatusCommand(set))

	return cmd
}

func getAccountCommand(set clientset.ClientSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage your Intility Developer Platform account",
		Long:  `Manage your Intility Developer Platform account`,
		Run:   showHelp,
	}

	cmd.AddCommand(account.NewShowCommand(set))

	return cmd
}

func getTeamsCommand(set clientset.ClientSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "team",
		Short: "Manage Intility Developer Platform teams",
		Long:  "Manage Intility Developer Platform teams",
		Run:   showHelp,
	}

	cmd.AddCommand(teams.NewListCommand(set))
	cmd.AddCommand(teams.NewGetCommand(set))
	cmd.AddCommand(teams.NewCreateCommand(set))
	cmd.AddCommand(teams.NewDeleteCommand(set))

	cmd.AddCommand(getMemberCommand(set))

	return cmd
}

func getMemberCommand(set clientset.ClientSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member",
		Short: "Manage Intility Developer Platform team members",
		Long:  "Manage Intility Developer Platform team members",
		Run:   showHelp,
	}

	cmd.AddCommand(member.NewAddCommand(set))
	cmd.AddCommand(member.NewRemoveCommand(set))

	return cmd
}

func getUserCommand(set clientset.ClientSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage Intility Developer Platform users",
		Long:  "Manage Intility Developer Platform users",
		Run:   showHelp,
	}

	cmd.AddCommand(user.NewListCommand(set))

	return cmd
}

func showHelp(cmd *cobra.Command, args []string) {
	_, span := telemetry.StartSpan(cmd.Context(), cmd.Use)
	defer span.End()

	_ = cmd.Help()
}
