package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/intility/minctl/pkg/tokencache"
)

// logoutCmd represents the logout command.
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout the current account",
	Long:  `Logout the current account.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := tokencache.New().Clear()
		if err != nil {
			return fmt.Errorf("logout failed: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
