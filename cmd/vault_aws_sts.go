package cmd

import (
	"fmt"
	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// vaultAwsStsCmd represents the sts command
var vaultAwsStsCmd = &cobra.Command{
	Use:   "sts",
	Short: "Get STS credentials from AWS vault secret backend and write them to .aws/credentials file",
	PreRun: func(cmd *cobra.Command, args []string) {

		for _, flag := range []string{"vault-addr", "vault-aws-path", "vault-aws-sts-role"} {

			// Bind viper to flag
			err := viper.BindPFlag(flag, cmd.Flags().Lookup(flag))
			if err != nil {
				ErrorToEval(fmt.Errorf("failed to bind flag %s to viper: %s", flag, err))
				return
			}

			// Check flag has a value
			if viper.GetString(flag) == "" {
				ErrorToEval(fmt.Errorf("flag %s must be defined", flag))
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Create vault client
		vc, err := vault.NewClient(&vault.Config{Address: viper.GetString("vault-addr")})
		if err != nil {
			ErrorToEval(fmt.Errorf("failed to create vault client: %w", err))
			return

		}

		// Read vault token from env
		token, err := getVaultToken()
		if err != nil {
			ErrorToEval(fmt.Errorf("failed to get token: %s", err))
			return
		}

		// Set token to Vault client
		vc.SetToken(string(token))

		// Get AWS STS credentials
		secret, err := vc.Logical().Write(fmt.Sprintf("%s/sts/%s", viper.GetString("vault-aws-path"), viper.GetString("vault-aws-sts-role")), nil)
		if err != nil {
			ErrorToEval(fmt.Errorf("failed to get STS credentials from Vault: %w", err))
			return
		}

		// Export AWS credentials as environment variables
		fmt.Printf("export AWS_ACCESS_KEY_ID=%q\n", secret.Data["access_key"].(string))
		fmt.Printf("export AWS_SECRET_ACCESS_KEY=%q\n", secret.Data["secret_key"].(string))
		fmt.Printf("export AWS_SESSION_TOKEN=%q\n", secret.Data["security_token"].(string))
	},
}

func init() {
	vaultAwsStsCmd.Flags().String("vault-addr", "", "Vault server address [GOGCI_VAULT_ADDR]")
	vaultAwsStsCmd.Flags().String("vault-aws-path", "aws_sts", "Vault AWS backend mount [GOGCI_VAULT_AWS_PATH]")
	vaultAwsStsCmd.Flags().String("vault-aws-sts-role", "", "Vault AWS STS role [GOGCI_VAULT_AWS_STS_ROLE]")

	vaultAwsCmd.AddCommand(vaultAwsStsCmd)
}
