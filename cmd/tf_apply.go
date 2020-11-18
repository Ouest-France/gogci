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

		// Bind string flags
		for _, flag := range []string{"gitlab-url", "gitlab-token"} {

			// Bind viper to flag
			err := viper.BindPFlag(flag, cmd.Flags().Lookup(flag))
			if err != nil {
				return fmt.Errorf("error binding viper to flag %q: %w", flag, err)
			}

			// Check flag has a value
			if viper.GetString(flag) == "" {
				return fmt.Errorf("flag %q must be defined", flag)
			}
		}

		// Bind bool flags
		for _, flag := range []string{"approved", "oldest"} {

			// Bind viper to flag
			err := viper.BindPFlag(flag, cmd.Flags().Lookup(flag))
			if err != nil {
				return fmt.Errorf("error binding viper to flag %q: %w", flag, err)
			}
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		// Create gitlab client
		gc := gitlab.Client{Token: viper.GetString("gitlab-token"), URL: viper.GetString("gitlab-url")}

		// Check merge request approval when approved flag is set
		if viper.GetBool("approved") {
			approved, err := gc.CheckMergeRequestApproved()
			if err != nil {
				return fmt.Errorf("failed to check merge request approval: %w", err)
			}

			if !approved {
				err = gc.TerraformApplyNotApproved()
				if err != nil {
					return fmt.Errorf("failed to send 'terraform apply not approved' comment: %w", err)
				}

				return fmt.Errorf("merge request must be approved to execute 'terraform apply'")
			}
		}

		// Check that no older merge request are in open state
		if viper.GetBool("oldest") {
			oldest, err := gc.CheckOldestMergeRequest()
			if err != nil {
				return fmt.Errorf("failed to check is current merge request is the oldest still open")
			}
			if !oldest {
				if err != nil {
					return fmt.Errorf("failed to send 'terraform apply blocked' comment: %w", err)
				}

				return fmt.Errorf("all older merge requests must be closed to launch 'terraform apply'")
			}
		}

		// Notify apply start
		err := gc.TerraformApplyRunning()
		if err != nil {
			return fmt.Errorf("error sending terraform apply notification: %w", err)
		}

		// Execute Apply
		stdout, stderr, _, err := command.Run("terraform", append([]string{"apply"}, args...))
		if err != nil {
			errGit := gc.TerraformApplyFailed(stripansi.Strip(string(stderr)))
			if errGit != nil {
				return fmt.Errorf("error during terraform apply: %s: %w", errGit, err)
			}
			return fmt.Errorf("error during terraform apply: %w", err)
		}

		// Notify apply summary
		err = gc.TerraformApplySummary(stripansi.Strip(string(stdout)))
		if err != nil {
			return fmt.Errorf("error sending apply summery notification: %w", err)
		}

		return nil
	},
}

func init() {
	tfApplyCmd.Flags().String("gitlab-url", os.Getenv("CI_API_V4_URL"), "Gitlab API url (default: CI_API_V4_URL) [GOGCI_GITLAB_URL]")
	tfApplyCmd.Flags().String("gitlab-token", "", "Gitlab API token [GOGCI_GITLAB_TOKEN]")
	tfApplyCmd.Flags().Bool("approved", true, "Execute apply only when the merge request is approved [GOGCI_APPROVED]")
	tfApplyCmd.Flags().Bool("oldest", true, "Execute apply only when no older merge requests is in open state [GOGCI_OLDEST]")

	tfCmd.AddCommand(tfApplyCmd)
}
