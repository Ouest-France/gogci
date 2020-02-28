package cmd

import (
	"fmt"
	"os"

	"github.com/Ouest-France/gogci/command"
	"github.com/Ouest-France/gogci/gitlab"
	"github.com/acarl005/stripansi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// tfApplyCmd represents the "tf plan" command
var tfApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Launch terraform apply and send output to Gitlab MR comment",
	PreRunE: func(cmd *cobra.Command, args []string) error {

		for _, flag := range []string{"gitlab-url", "gitlab-token"} {

			// Bind viper to flag
			err := viper.BindPFlag(flag, cmd.Flags().Lookup(flag))
			if err != nil {
				return fmt.Errorf("Error binding viper to flag %q: %s", flag, err)
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

		// Notify apply start
		err := gc.TerraformApplyRunning()
		if err != nil {
			return err
		}

		// Execute Apply
		stdout, stderr, _, err := command.Run("terraform", append([]string{"apply"}, args...))
		if err != nil {
			errGit := gc.TerraformApplyFailed(stripansi.Strip(string(stderr)))
			if errGit != nil {
				return fmt.Errorf("%s: %s", errGit, err)
			}
			return err
		}

		// Notify apply summary
		err = gc.TerraformApplySummary(stripansi.Strip(string(stdout)))
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	tfApplyCmd.Flags().String("gitlab-url", os.Getenv("CI_API_V4_URL"), "Gitlab API url (default: CI_API_V4_URL) [GOGCI_GITLAB_URL]")
	tfApplyCmd.Flags().String("gitlab-token", "", "Gitlab API token [GOGCI_GITLAB_TOKEN]")

	tfCmd.AddCommand(tfApplyCmd)
}
