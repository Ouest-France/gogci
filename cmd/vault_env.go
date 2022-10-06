package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
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
				fmt.Printf("echo \"echo error binding viper to flag %q: %s\"", flag, err)
				return
			}
		}

		for _, flag := range []string{"vault-addr", "vault-secret"} {

			// Check flag has a value
			if viper.GetString(flag) == "" {
				fmt.Printf("echo \"echo flag %s must be defined\"", flag)
				return
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Create vault client
		vc, err := vault.NewClient(&vault.Config{Address: viper.GetString("vault-addr")})
		if err != nil {
			fmt.Printf("echo \"echo failed to create vault client: %s\"", err)
			return
		}

		// Read vault token from env
		token := os.Getenv("VAULT_TOKEN")

		if token == "" {
			// Read vault token on disk
			tokenPath, err := homedir.Expand("~/.vault-token")
			if err != nil {
				fmt.Printf("echo \"echo failed to construct vault token path: %s\"", err)
				return
			}

			tokenFile, err := ioutil.ReadFile(tokenPath)
			if err != nil {
				fmt.Printf("echo \"echo failed to read token: %s\"", err)
				return
			}

			token = string(tokenFile)
		}

		// Set token to Vault client
		vc.SetToken(string(token))

		// Template prefix
		prefixTmpl, err := template.New("prefix").Funcs(sprig.TxtFuncMap()).Parse(viper.GetString("vault-secret-prefix"))
		if err != nil {
			fmt.Printf("echo \"echo failed to create prefix template: %s\"", err)
			return
		}

		var prefix bytes.Buffer
		if err := prefixTmpl.Execute(&prefix, nil); err != nil {
			fmt.Printf("echo \"echo %s\"", err)
			return
		}

		// Template secret
		secretPathTmpl, err := template.New("secret").Funcs(sprig.TxtFuncMap()).Parse(viper.GetString("vault-secret"))
		if err != nil {
			fmt.Printf("echo \"echo failed to create secretPath template: %s\"", err)
			return
		}

		var secretPath bytes.Buffer
		if err := secretPathTmpl.Execute(&secretPath, nil); err != nil {
			fmt.Printf("echo \"echo %s\"", err)
			return
		}

		// Get Vault secret
		secret, err := vc.Logical().Read(fmt.Sprintf("%s/%s", prefix.String(), secretPath.String()))
		if err != nil {
			fmt.Printf("echo \"echo failed to get secret from Vault: %s\"", err)
			return
		}

		// Check if secret exists
		if secret == nil {
			fmt.Printf("echo \"echo no secret found at path %s/%s\"", prefix.String(), secretPath.String())
			return
		}

		// Check if data entry exists
		data, ok := secret.Data["data"]
		if !ok {
			fmt.Printf("echo \"echo no data found at path %s/%s\"", prefix.String(), secretPath.String())
			return
		}

		// Export keys as env vars
		for key, value := range data.(map[string]interface{}) {
			envName, err := convertToEnvName(key)
			if err != nil {
				fmt.Printf("echo \"echo failed to calculate env var name from vault secret: %s\"", err)
				return
			}
			fmt.Printf("export %q=%q\n", envName, value)
		}
	},
}

func convertToEnvName(name string) (string, error) {

	// Remove characters that are not alphanum or . _ -
	r, err := regexp.Compile("[a-zA-Z0-9._-]")
	if err != nil {
		return "", fmt.Errorf("failed to compile regex: %w", err)
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
