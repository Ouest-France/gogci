package cmd

import (
	"fmt"
	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// vaultEnv represents the env command
var vaultEnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Get a Vault secret and export all keys as env vars",
	PreRun: func(cmd *cobra.Command, args []string) {

		for _, flag := range []string{"vault-addr", "vault-secret", "vault-secret-prefix"} {

			// Bind viper to flag
			err := viper.BindPFlag(flag, cmd.Flags().Lookup(flag))
			if err != nil {
				ErrorToEval(fmt.Errorf("failed to bind flag %s to viper: %s", flag, err))
				return
			}
		}

		for _, flag := range []string{"vault-addr", "vault-secret"} {

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
		if err != nil {
			ErrorToEval(fmt.Errorf("failed to get secret from Vault: %s", err))
			return
		}

		// Export keys as env vars
		for key, value := range data {
			envName, err := convertToEnvName(key)
			if err != nil {
				ErrorToEval(fmt.Errorf("failed to calculate env var name from vault secret: %s", err))
				return
			}
			fmt.Printf("export %q=%q\n", envName, value)
		}
	},
}

func init() {
	vaultEnvCmd.Flags().String("vault-addr", "", "Vault server address [GOGCI_VAULT_ADDR]")
	vaultEnvCmd.Flags().String("vault-secret", "", "Vault secret path [GOGCI_VAULT_SECRET]")
	vaultEnvCmd.Flags().String("vault-secret-prefix", "", "Vault secret path prefix [GOGCI_VAULT_SECRET_PREFIX]")

	vaultCmd.AddCommand(vaultEnvCmd)
}
