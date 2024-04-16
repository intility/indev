package cmd

import (
	"fmt"
	"github.com/intility/minctl/pkg/credentialstore"
	"github.com/spf13/cobra"
)

// showCmd represents the show command.
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show account information",
	Long:  `Show information about the current logged in account.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		credStore := credentialstore.New()
		cred, err := credStore.Get("foo")
		if err != nil {
			return fmt.Errorf("could not get credentials: %w", err)
		}

		fmt.Println("Retrieved credentials: ", string(cred))

		return nil
	},
}

func init() {
	accountCmd.AddCommand(showCmd)
}
