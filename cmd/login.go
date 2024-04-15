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

		if err := credStore.Set("foo", []byte("success")); err != nil {
			return fmt.Errorf("could not set credentials: %w", err)
		}

		val, err := credStore.Get("foo")
		if err != nil {
			return fmt.Errorf("could not get credentials: %w", err)
		}

		fmt.Println(string(val))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
