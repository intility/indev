package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/adrg/xdg"
	"github.com/spf13/cobra"

	"github.com/intility/minctl/pkg/credentialstore"
	"github.com/intility/minctl/pkg/tokencache"
)

// loginCmd represents the login command.
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Intility Container Platform",
	Long:  `Login to Intility Container Platform using your Intility credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// confidential clients have a credential, such as a secret or a certificate
		clientID := "b65cf9b0-290c-4b44-a4b1-0b02b7752b3c"
		authority := "https://login.microsoftonline.com/intility.no"

		cacheFile := filepath.Join(xdg.DataHome, "minctl", "msal.cache")
		credStore := credentialstore.NewFilesystemCredentialStore(cacheFile)
		cache := tokencache.New(tokencache.WithCredentialStore(credStore))

		publicClient, err := public.New(
			clientID,
			public.WithAuthority(authority),
			public.WithCache(cache),
		)
		if err != nil {
			return fmt.Errorf("could not create public client: %w", err)
		}

		scopes := []string{"api://containerplatform.intility.com/user_impersonation"}

		var result public.AuthResult

		accounts, err := publicClient.Accounts(cmd.Context())
		if len(accounts) > 0 {
			result, err = publicClient.AcquireTokenSilent(
				cmd.Context(),
				scopes,
				public.WithSilentAccount(accounts[0]),
			)
		}

		if err != nil || len(accounts) == 0 {
			result, err = publicClient.AcquireTokenInteractive(
				cmd.Context(),
				scopes,
				public.WithRedirectURI("http://localhost:42069"),
			)
			if err != nil {
				return fmt.Errorf("could not acquire token: %w", err)
			}
		}

		fmt.Println("Access token: ", result.AccessToken)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
