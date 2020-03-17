package cmd

import (
	"fmt"
	"io/ioutil"

	vault "github.com/hashicorp/vault/api"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// vaultLoginCmd represents the sts command
var vaultLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Get Vault token and write it to .vault-token file",
	PreRunE: func(cmd *cobra.Command, args []string) error {

		for _, flag := range []string{"vault-addr", "vault-role-id", "vault-secret-id"} {

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

		// Create vault client
		vc, err := vault.NewClient(&vault.Config{Address: viper.GetString("vault-addr")})
		if err != nil {
			return fmt.Errorf("failed to create vault client: %s", err)
		}

		// AppRole login
		approle := map[string]interface{}{
			"role_id":   viper.GetString("vault-role-id"),
			"secret_id": viper.GetString("vault-secret-id"),
		}
		secret, err := vc.Logical().Write("auth/approle/login", approle)
		if err != nil {
			return fmt.Errorf("failed Vault login by approle: %s", err)
		}

		// Expand home token path
		tokenPath, err := homedir.Expand("~/.vault-token")
		if err != nil {
			return fmt.Errorf("failed to construct vault token path: %s", err)
		}

		// Write Vault token
		err = ioutil.WriteFile(tokenPath, []byte(secret.Auth.ClientToken), 0600)
		if err != nil {
			return fmt.Errorf("failed to write Vault token to disk: %s", err)
		}

		return nil
	},
}

func init() {
	vaultLoginCmd.Flags().String("vault-addr", "", "Vault server address [GOGCI_VAULT_ADDR]")
	vaultLoginCmd.Flags().String("vault-role-id", "", "Vault AppRole Role ID [GOGCI_VAULT_ROLE_ID]")
	vaultLoginCmd.Flags().String("vault-secret-id", "", "Vault AppRole Secret ID [GOGCI_VAULT_SECRET_ID]")

	vaultCmd.AddCommand(vaultLoginCmd)
}
