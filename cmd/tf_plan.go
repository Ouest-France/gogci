package cmd

import (
	"fmt"
	"os"

	"github.com/Ouest-France/gogci/command"
	"github.com/Ouest-France/gogci/notify"
	"github.com/acarl005/stripansi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// tfPlanCmd represents the "tf plan" command
var tfPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Launch terraform plan and send output to Gitlab MR comment",
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
		git := notify.Gitlab{Token: viper.GetString("gitlab-token"), URL: viper.GetString("gitlab-url")}

		// Notify plan start
		err := git.TerraformPlanRunning()
		if err != nil {
			return err
		}

		// Execute plan
		stdout, _, _, err := command.Run("terraform", append([]string{"plan"}, args...))
		if err != nil {
			return err
		}

		// Notify plan summary
		err = git.TerraformPlanSummary(stripansi.Strip(string(stdout)))
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	tfPlanCmd.Flags().String("gitlab-url", os.Getenv("CI_API_V4_URL"), "Gitlab API url (default: CI_API_V4_URL) [GOGCI_GITLAB_URL]")
	tfPlanCmd.Flags().String("gitlab-token", "", "Gitlab API token [GOGCI_GITLAB_TOKEN]")

	tfCmd.AddCommand(tfPlanCmd)
}
