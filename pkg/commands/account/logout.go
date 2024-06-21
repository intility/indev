package account

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/intility/icpctl/internal/telemetry"
	"github.com/intility/icpctl/pkg/clientset"
	"github.com/intility/icpctl/pkg/tokencache"
)

func NewLogoutCommand(set clientset.ClientSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout the current account",
		Long:  `Logout the current account.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, span := telemetry.StartSpan(cmd.Context(), "account.logout")
			defer span.End()

			err := tokencache.New().Clear()
			if err != nil {
				return fmt.Errorf("logout failed: %w", err)
			}

			return nil
		},
	}

	return cmd
}
