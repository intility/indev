package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/intility/minctl/pkg/client"
	"github.com/intility/minctl/pkg/wizard"
)

var (
	clusterName string
)

// clusterCreateCmd represents the create command.
var clusterCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cluster",
	Long:  `Create a new cluster with the specified configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(client.WithDevConfig())

		if clusterName == "" {
			wz := wizard.NewWizard([]wizard.Input{
				{
					ID:          "name",
					Placeholder: "Cluster Name",
					Type:        wizard.InputTypeText,
					Limit:       0,
					Validator:   nil,
				},
			})

			result, err := wz.Run()
			if err != nil {
				return fmt.Errorf("could not gather information: %w", err)
			}

			if result.Cancelled() {
				return nil
			}

			clusterName := result.MustGetValue("name")

			if clusterName == "" {
				return errEmptyName
			}

			req := client.NewClusterRequest{Name: clusterName}
			cmd.SilenceUsage = true

			_, err = c.CreateCluster(cmd.Context(), req)
			if err != nil {
				return fmt.Errorf("could not create cluster: %w", err)
			}

			cmd.Printf("%s: cluster created\n", styleSuccess.Render("success"))

			return nil
		}

		req := client.NewClusterRequest{Name: clusterName}
		cluster, err := c.CreateCluster(cmd.Context(), req)

		// we are done with validating the input
		cmd.SilenceUsage = true

		if err != nil {
			return fmt.Errorf("could not create cluster: %w", err)
		}

		cmd.Println(cluster.Name)

		return nil
	},
}

func init() {
	clusterCreateCmd.Flags().StringVarP(&clusterName, "name", "n", "", "Name of the cluster to create")

	clusterCmd.AddCommand(clusterCreateCmd)
}
