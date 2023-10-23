package cmd

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"text/template"
)

func getVaultToken() (string, error) {

	token := os.Getenv("VAULT_TOKEN")

	if token == "" {
		// Read vault token on disk
		tokenPath, err := homedir.Expand("~/.vault-token")
		if err != nil {
			return "", fmt.Errorf("failed to construct vault token path: %s", err)
		}

		tokenFile, err := ioutil.ReadFile(tokenPath)
		if err != nil {
			return "", fmt.Errorf("failed to read token: %s", err)
		}

		token = string(tokenFile)
	}

	return token, nil
}

func getSecretPath() (string, error) {
	// Template prefix
	prefixTmpl, err := template.New("prefix").Funcs(sprig.TxtFuncMap()).Parse(viper.GetString("vault-secret-prefix"))
	if err != nil {
		return "", fmt.Errorf("failed to create prefix template: %s", err)
	}

	var prefix bytes.Buffer
	if err := prefixTmpl.Execute(&prefix, nil); err != nil {
		return "", fmt.Errorf("failed to execute prefix template: %s", err)
	}

	// Template secret
	secretPathTmpl, err := template.New("secret").Funcs(sprig.TxtFuncMap()).Parse(viper.GetString("vault-secret"))
	if err != nil {
		return "", fmt.Errorf("failed to create secretPath template: %s", err)
	}

	var secretPath bytes.Buffer
	if err := secretPathTmpl.Execute(&secretPath, nil); err != nil {
		return "", fmt.Errorf("failed to execute secretPath template: %s", err)
	}

	return fmt.Sprintf("%s/%s", prefix.String(), secretPath.String()), nil
}

func getSecretData(vc *api.Client, secretPath string) (map[string]interface{}, error) {
	// Get Vault secret
	secret, err := vc.Logical().Read(secretPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret from Vault: %s", err)
	}

	// Check if secret exists
	if secret == nil {
		return nil, fmt.Errorf("no secret found at path %s", secretPath)
	}

	// Check if data entry exists
	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no data found at path %s", secretPath)
	}

	return data, nil
}
