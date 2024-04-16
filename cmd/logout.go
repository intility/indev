package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"

	"github.com/intility/minctl/pkg/credentialstore"
	"github.com/intility/minctl/pkg/tokencache"
)

// logoutCmd represents the logout command.
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout the current account",
	Long:  `Logout the current account.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		cacheFile := filepath.Join(xdg.DataHome, "minctl", "msal.cache")
		credStore := credentialstore.NewFilesystemCredentialStore(cacheFile)
		cache := tokencache.New(tokencache.WithCredentialStore(credStore))

		err := cache.Clear()
		if err != nil {
			return fmt.Errorf("could not clear token cache: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
