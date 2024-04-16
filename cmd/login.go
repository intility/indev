package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/intility/minctl/pkg/credentialstore"
)

// loginCmd represents the login command.
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Intility Container Platform",
	Long:  `Login to Intility Container Platform using your Intility credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		credStore := credentialstore.New()

		if err := credStore.Set("foo", []byte("max.mekker@example.com")); err != nil {
			return fmt.Errorf("could not set credentials: %w", err)
		}

		fmt.Println("Successfully signed in")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
