package cmd

import (
	"fmt"
	"os"

	"github.com/Ouest-France/gogci/command"
	"github.com/Ouest-France/gogci/gitlab"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// tfInitCmd represents the "tf init" command
var tfInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Launch terraform init",
	PreRunE: func(cmd *cobra.Command, args []string) error {

		for _, flag := range []string{"gitlab-url", "gitlab-token"} {

			// Bind viper to flag
			err := viper.BindPFlag(flag, cmd.Flags().Lookup(flag))
			if err != nil {
				return fmt.Errorf("Error binding viper to flag %q: %w", flag, err)
			}

			// Check flag has a value
			if viper.GetString(flag) == "" {
				return fmt.Errorf("Flag %q must be defined", flag)
			}
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		// Create gitlab client
		gc := gitlab.Client{Token: viper.GetString("gitlab-token"), URL: viper.GetString("gitlab-url")}

		// Execute init
		_, _, _, err := command.Run("terraform", []string{"init"})
		if err != nil {
			errGit := gc.TerraformInitFailed()
			if errGit != nil {
				return fmt.Errorf("error sending terraform init failed notification: %s: %w", errGit, err)
			}
			return fmt.Errorf("error during terraform apply: %w", err)
		}

		return nil
	},
}

func init() {
	tfInitCmd.Flags().String("gitlab-url", os.Getenv("CI_API_V4_URL"), "Gitlab API url (default: CI_API_V4_URL) [GOGCI_GITLAB_URL]")
	tfInitCmd.Flags().String("gitlab-token", "", "Gitlab API token [GOGCI_GITLAB_TOKEN]")

	tfCmd.AddCommand(tfInitCmd)
}
