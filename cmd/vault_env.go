package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	vault "github.com/hashicorp/vault/api"
	homedir "github.com/mitchellh/go-homedir"
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
		token := os.Getenv("VAULT_TOKEN")

		if token == "" {
			// Read vault token on disk
			tokenPath, err := homedir.Expand("~/.vault-token")
			if err != nil {
				ErrorToEval(fmt.Errorf("failed to construct vault token path: %s", err))
				return
			}

			tokenFile, err := ioutil.ReadFile(tokenPath)
			if err != nil {
				ErrorToEval(fmt.Errorf("failed to read token: %s", err))
				return
			}

			token = string(tokenFile)
		}

		// Set token to Vault client
		vc.SetToken(string(token))

		// Template prefix
		prefixTmpl, err := template.New("prefix").Funcs(sprig.TxtFuncMap()).Parse(viper.GetString("vault-secret-prefix"))
		if err != nil {
			ErrorToEval(fmt.Errorf("failed to create prefix template: %s", err))
			return
		}

		var prefix bytes.Buffer
		if err := prefixTmpl.Execute(&prefix, nil); err != nil {
			ErrorToEval(fmt.Errorf("failed to execute prefix template: %s", err))
			return
		}

		// Template secret
		secretPathTmpl, err := template.New("secret").Funcs(sprig.TxtFuncMap()).Parse(viper.GetString("vault-secret"))
		if err != nil {
			ErrorToEval(fmt.Errorf("failed to create secretPath template: %s", err))
			return
		}

		var secretPath bytes.Buffer
		if err := secretPathTmpl.Execute(&secretPath, nil); err != nil {
			ErrorToEval(fmt.Errorf("failed to execute secretPath template: %s", err))
			return
		}

		// Get Vault secret
		secret, err := vc.Logical().Read(fmt.Sprintf("%s/%s", prefix.String(), secretPath.String()))
		if err != nil {
			ErrorToEval(fmt.Errorf("failed to get secret from Vault: %s", err))
			return
		}

		// Check if secret exists
		if secret == nil {
			ErrorToEval(fmt.Errorf("no secret found at path %s/%s", prefix.String(), secretPath.String()))
			return
		}

		// Check if data entry exists
		data, ok := secret.Data["data"]
		if !ok {
			ErrorToEval(fmt.Errorf("no data found at path %s/%s", prefix.String(), secretPath.String()))
			return
		}

		// Export keys as env vars
		for key, value := range data.(map[string]interface{}) {
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
