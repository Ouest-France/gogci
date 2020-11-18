package cmd

import (
	"fmt"

	"github.com/Ouest-France/gogci/awsconfig"
	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// vaultAwsStsCmd represents the sts command
var vaultAwsStsCmd = &cobra.Command{
	Use:   "sts",
	Short: "Get STS credentials from AWS vault secret backend and write them to .aws/credentials file",
	PreRunE: func(cmd *cobra.Command, args []string) error {

		for _, flag := range []string{"vault-addr", "vault-role-id", "vault-secret-id", "vault-aws-path", "vault-aws-sts-role", "aws-profile"} {

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

		// Create vault client
		vc, err := vault.NewClient(&vault.Config{Address: viper.GetString("vault-addr")})
		if err != nil {
			return fmt.Errorf("failed to create vault client: %w", err)
		}

		// AppRole login
		approle := map[string]interface{}{
			"role_id":   viper.GetString("vault-role-id"),
			"secret_id": viper.GetString("vault-secret-id"),
		}
		secret, err := vc.Logical().Write("auth/approle/login", approle)
		if err != nil {
			return fmt.Errorf("failed Vault login by approle: %w", err)
		}
		vc.SetToken(secret.Auth.ClientToken)

		// Get AWS STS credentials
		secret, err = vc.Logical().Read(fmt.Sprintf("%s/sts/%s", viper.GetString("vault-aws-path"), viper.GetString("vault-aws-sts-role")))
		if err != nil {
			return fmt.Errorf("failed to get STS credentials from Vault: %w", err)
		}

		// Write AWS credentials file
		accessKey := secret.Data["access_key"].(string)
		secretKey := secret.Data["secret_key"].(string)
		sessionToken := secret.Data["security_token"].(string)
		err = awsconfig.WriteCredentials(viper.GetString("aws-profile"), accessKey, secretKey, sessionToken)
		if err != nil {
			return fmt.Errorf("failed to write aws credentials to disk: %w", err)
		}

		return nil
	},
}

func init() {
	vaultAwsStsCmd.Flags().String("vault-addr", "", "Vault server address [GOGCI_VAULT_ADDR]")
	vaultAwsStsCmd.Flags().String("vault-role-id", "", "Vault AppRole Role ID [GOGCI_VAULT_ROLE_ID]")
	vaultAwsStsCmd.Flags().String("vault-secret-id", "", "Vault AppRole Secret ID [GOGCI_VAULT_SECRET_ID]")
	vaultAwsStsCmd.Flags().String("vault-aws-path", "aws", "Vault AWS backend mount [GOGCI_VAULT_AWS_PATH]")
	vaultAwsStsCmd.Flags().String("vault-aws-sts-role", "", "Vault AWS STS role [GOGCI_VAULT_AWS_STS_ROLE]")
	vaultAwsStsCmd.Flags().String("aws-profile", "default", "AWS config/credentials profile [GOGCI_AWS_PROFILE]")

	vaultAwsCmd.AddCommand(vaultAwsStsCmd)
}
