package rootcommand

import (
	"github.com/spf13/cobra"

	"github.com/intility/idpctl/internal/build"
	"github.com/intility/idpctl/internal/telemetry"
	"github.com/intility/idpctl/internal/ux"
	"github.com/intility/idpctl/pkg/authenticator"
	"github.com/intility/idpctl/pkg/client"
	"github.com/intility/idpctl/pkg/clientset"
	"github.com/intility/idpctl/pkg/commands/account"
	"github.com/intility/idpctl/pkg/commands/cluster"
)

func GetRootCommand() *cobra.Command {
	clients := clientset.ClientSet{
		Authenticator:  authenticator.NewAuthenticator(authenticator.ConfigFromBuildProps()),
		PlatformClient: client.New(),
	}

	rootCmd := &cobra.Command{
		Use:   build.AppName,
		Short: build.AppName + " controls your Intility Container Platform instance.",
		Long:  ``,
		Run:   showHelp,
	}

	rootCmd.AddCommand(getVersionCommand())
	rootCmd.AddCommand(account.NewLoginCommand(clients))
	rootCmd.AddCommand(account.NewLogoutCommand(clients))
	rootCmd.AddCommand(getClusterCommand(clients))
	rootCmd.AddCommand(getAccountCommand(clients))

	return rootCmd
}

func getVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  `Print the version information of idpctl.`,
		Run: func(cmd *cobra.Command, args []string) {
			_, span := telemetry.StartSpan(cmd.Context(), "version")
			defer span.End()

			ux.Fprint(cmd.OutOrStdout(), build.NameVersionString()+"\n")
		},
	}
}

func getClusterCommand(set clientset.ClientSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage your Intility Container Platform clusters",
		Long:  `Manage your Intility Container Platform clusters`,
		Run:   showHelp,
	}

	cmd.AddCommand(cluster.NewCreateCommand(set))
	cmd.AddCommand(cluster.NewDeleteCommand(set))
	cmd.AddCommand(cluster.NewListCommand(set))

	return cmd
}

func getAccountCommand(set clientset.ClientSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage your Intility Container Platform account",
		Long:  `Manage your Intility Container Platform account`,
		Run:   showHelp,
	}

	cmd.AddCommand(account.NewShowCommand(set))

	return cmd
}

func showHelp(cmd *cobra.Command, args []string) {
	_, span := telemetry.StartSpan(cmd.Context(), cmd.Use)
	defer span.End()

	_ = cmd.Help()
}
