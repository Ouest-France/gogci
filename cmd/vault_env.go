package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	vault "github.com/hashicorp/vault/api"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// vaultEnv represents the env command
var vaultEnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Get a Vault secret and export all keys as env vars",
	PreRunE: func(cmd *cobra.Command, args []string) error {

		for _, flag := range []string{"vault-addr", "vault-secret", "vault-secret-prefix"} {

			// Bind viper to flag
			err := viper.BindPFlag(flag, cmd.Flags().Lookup(flag))
			if err != nil {
				return fmt.Errorf("Error binding viper to flag %q: %s", flag, err)
			}
		}

		for _, flag := range []string{"vault-addr", "vault-secret"} {

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

		// Read vault token from env
		token := os.Getenv("VAULT_TOKEN")

		if token == "" {
			// Read vault token on disk
			tokenPath, err := homedir.Expand("~/.vault-token")
			if err != nil {
				return fmt.Errorf("failed to construct vault token path: %s", err)
			}

			tokenFile, err := ioutil.ReadFile(tokenPath)
			if err != nil {
				return fmt.Errorf("failed to read token: %s", err)
			}

			token = string(tokenFile)
		}

		// Set token to Vault client
		vc.SetToken(string(token))

		// Get Vault secret
		secret, err := vc.Logical().Read(fmt.Sprintf("%s/%s", viper.GetString("vault-secret-prefix"), viper.GetString("vault-secret")))
		if err != nil {
			return fmt.Errorf("failed to get secret from Vault: %s", err)
		}

		// Check if data entry exists
		data, ok := secret.Data["data"]
		if !ok {
			return fmt.Errorf("no data found at path %s/%s", viper.GetString("vault-secret-prefix"), viper.GetString("vault-secret"))
		}

		// Export keys as env vars
		for key, value := range data.(map[string]interface{}) {
			envName, err := convertToEnvName(key)
			if err != nil {
				return fmt.Errorf("failed to calculate env var name from vault secret: %s", err)
			}
			fmt.Printf("export %q=%q\n", envName, value)
		}

		return nil
	},
}

func convertToEnvName(name string) (string, error) {

	// Remove characters that are not alphanum or . _ -
	r, err := regexp.Compile("[a-zA-Z0-9._-]")
	if err != nil {
		return "", fmt.Errorf("failed to compile regex: %s", err)
	}
	var sanitizedName string
	for _, l := range name {
		if r.MatchString(string(l)) {
			sanitizedName = sanitizedName + string(l)
		}
	}

	// Error if sanitized name is empty
	if sanitizedName == "" {
		return "", errors.New("empty sanitized name")
	}

	// Set name to be in upper case
	capitalizedName := strings.ToUpper(sanitizedName)

	// Replace - and . with _
	substituedName := capitalizedName
	for old, new := range map[string]string{"-": "_", ".": "_"} {
		substituedName = strings.ReplaceAll(substituedName, old, new)
	}

	// Add VAULTENV prefix
	finalName := "VAULTENV_" + substituedName

	return finalName, nil
}

func init() {
	vaultEnvCmd.Flags().String("vault-addr", "", "Vault server address [GOGCI_VAULT_ADDR]")
	vaultEnvCmd.Flags().String("vault-secret", "", "Vault secret path [GOGCI_VAULT_SECRET]")
	vaultEnvCmd.Flags().String("vault-secret-prefix", "", "Vault secret path prefix [GOGCI_VAULT_SECRET_PREFIX]")

	vaultCmd.AddCommand(vaultEnvCmd)
}
