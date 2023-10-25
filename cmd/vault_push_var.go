package cmd

import (
	"fmt"
	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var vaultPushEnv = &cobra.Command{
	Use:   "push-secret",
	Short: "Push secret to vault",
	PreRun: func(cmd *cobra.Command, args []string) {

		for _, flag := range []string{"vault-addr", "vault-secret", "vault-secret-prefix", "vault-push-secret-key", "vault-push-secret-value"} {

			// Bind viper to flag
			err := viper.BindPFlag(flag, cmd.Flags().Lookup(flag))
			if err != nil {
				ErrorToEval(fmt.Errorf("failed to bind flag %s to viper: %s", flag, err))
				return
			}
		}

		for _, flag := range []string{"vault-addr", "vault-secret", "vault-push-secret-key", "vault-push-secret-value"} {

			// Check flag has a value
			if viper.GetString(flag) == "" {
				ErrorToEval(fmt.Errorf("flag %s must be defined", flag))
				return
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Create vault client
		vc, err := vault.NewClient(&vault.Config{Address: viper.GetString("vault-addr")})
		if err != nil {
			ErrorToEval(fmt.Errorf("failed to create vault client: %s", err))
			return
		}

		// Read vault token from env

		// Read vault token from env
		token, err := getVaultToken()
		if err != nil {
			ErrorToEval(fmt.Errorf("failed to get token: %s", err))
			return
		}
		// Set token to Vault client
		vc.SetToken(string(token))

		secretPath, err := getSecretPath()
		if err != nil {
			ErrorToEval(fmt.Errorf("failed to get secret path: %s", err))
			return
		}

		// Get Vault secret data
		data, err := getSecretData(vc, secretPath)
		if err != nil && err.Error() != "no secret found at path "+secretPath {
			ErrorToEval(fmt.Errorf("failed to get secret from Vault: %s", err))
			return
		}
		if err != nil && err.Error() == "no secret found at path "+secretPath {
			data = make(map[string]interface{})
		}

		key := viper.GetString("vault-push-secret-key")
		if key == "" {
			ErrorToEval(fmt.Errorf("failed to get key to push: %s", err))
			return
		}
		value := viper.GetString("vault-push-secret-value")
		if value == "" {
			ErrorToEval(fmt.Errorf("failed to get value to push: %s", err))
			return
		}

		//merge entries
		data[key] = value

		// Get Vault secret
		_, err = vc.Logical().Write(secretPath, map[string]interface{}{"data": data})
		if err != nil {
			ErrorToEval(fmt.Errorf("failed to push secret to Vault: %s", err))
			return
		}
	},
}

func init() {
	vaultPushEnv.Flags().String("vault-addr", "", "Vault server address [GOGCI_VAULT_ADDR]")
	vaultPushEnv.Flags().String("vault-secret", "", "Vault secret path [GOGCI_VAULT_SECRET]")
	vaultPushEnv.Flags().String("vault-secret-prefix", "", "Vault secret path prefix [GOGCI_VAULT_SECRET_PREFIX]")
	vaultPushEnv.Flags().String("vault-push-secret-key", "", "Key of the secret to push [GOGCI_VAULT_PUSH_SECRET_KEY]")
	vaultPushEnv.Flags().String("vault-push-secret-value", "", "Value of the secret to push [GOGCI_VAULT_PUSH_SECRET_VALUE]")

	vaultCmd.AddCommand(vaultPushEnv)
}
