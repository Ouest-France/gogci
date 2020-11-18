package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	vault "github.com/hashicorp/vault/api"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// vaultLoginCmd represents the vault login command
var vaultLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Get Vault token and write it to .vault-token file",
	PreRunE: func(cmd *cobra.Command, args []string) error {

		for _, flag := range []string{
			"vault-addr",
			"vault-method",
			"vault-role-id",
			"vault-secret-id",
			"vault-kubernetes-path",
			"vault-kubernetes-role",
			"vault-jwt-path",
			"vault-jwt-role",
			"vault-jwt-token",
			"vault-jwt-token-envvar",
			"export-token",
		} {
			// Bind viper to flag
			err := viper.BindPFlag(flag, cmd.Flags().Lookup(flag))
			if err != nil {
				return fmt.Errorf("Error binding viper to flag %q: %s", flag, err)
			}
		}

		for _, flag := range []string{"vault-addr", "vault-method"} {
			// Check flag has a value
			if viper.GetString(flag) == "" {
				return fmt.Errorf("Flag %q must be defined", flag)
			}
		}

		if viper.GetString("vault-jwt-token") == "" && viper.GetString("vault-jwt-token-envvar") == "" {
			return errors.New("One of vault-jwt-token or vault-jwt-token-envvar flags must be defined")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		// Create vault client
		vc, err := vault.NewClient(&vault.Config{Address: viper.GetString("vault-addr")})
		if err != nil {
			return fmt.Errorf("failed to create vault client: %w", err)
		}

		// Login
		secret := &vault.Secret{}

		switch method := viper.GetString("vault-method"); method {
		case "approle":
			// AppRole login
			approle := map[string]interface{}{
				"role_id":   viper.GetString("vault-role-id"),
				"secret_id": viper.GetString("vault-secret-id"),
			}
			secret, err = vc.Logical().Write("auth/approle/login", approle)
			if err != nil {
				return fmt.Errorf("failed Vault login by approle: %w", err)
			}

		case "kubernetes":
			// Get kubernetes service account
			token, err := ioutil.ReadFile("/run/secrets/kubernetes.io/serviceaccount/token")
			if err != nil {
				return fmt.Errorf("failed to read kubernetes service account token: %s", err)
			}

			// Kubernetes login
			kubeAuth := map[string]interface{}{
				"role": viper.GetString("vault-kubernetes-role"),
				"jwt":  string(token),
			}
			secret, err = vc.Logical().Write(fmt.Sprintf("auth/%s/login", viper.GetString("vault-kubernetes-path")), kubeAuth)
			if err != nil {
				return fmt.Errorf("failed Vault login by kubernetes: %w", err)
			}

		case "jwt":
			// Get JWT token
			var token string
			if viper.GetString("vault-jwt-token-envvar") == "" {
				token := viper.GetString("vault-jwt-token")
				if token == "" {
					return fmt.Errorf("empty jwt token read from flag %q", viper.GetString("vault-jwt-token"))
				}
			} else {
				token := os.Getenv(viper.GetString("vault-jwt-token-envvar"))
				if token == "" {
					return fmt.Errorf("empty jwt token read from env var %q", viper.GetString("vault-jwt-token-envvar"))
				}
			}

			// JWT login
			jwtAuth := map[string]interface{}{
				"role": viper.GetString("vault-jwt-role"),
				"jwt":  string(token),
			}
			secret, err = vc.Logical().Write(fmt.Sprintf("auth/%s/login", viper.GetString("vault-jwt-path")), jwtAuth)
			if err != nil {
				return fmt.Errorf("failed Vault login by JWT: %w", err)
			}
		}

		if viper.GetBool("export-token") {
			fmt.Printf("export %q=%q\n", "VAULT_TOKEN", secret.Auth.ClientToken)
		} else {
			// Expand home token path
			tokenPath, err := homedir.Expand("~/.vault-token")
			if err != nil {
				return fmt.Errorf("failed to construct vault token path: %w", err)
			}

			// Write Vault token
			err = ioutil.WriteFile(tokenPath, []byte(secret.Auth.ClientToken), 0600)
			if err != nil {
				return fmt.Errorf("failed to write Vault token to disk: %w", err)
			}
		}

		return nil
	},
}

func init() {
	vaultLoginCmd.Flags().String("vault-addr", "", "Vault server address [GOGCI_VAULT_ADDR]")
	vaultLoginCmd.Flags().String("vault-method", "approle", "Vault login method (default: approle) [GOGCI_VAULT_METHOD]")
	vaultLoginCmd.Flags().String("vault-role-id", "", "Vault AppRole Role ID [GOGCI_VAULT_ROLE_ID]")
	vaultLoginCmd.Flags().String("vault-secret-id", "", "Vault AppRole Secret ID [GOGCI_VAULT_SECRET_ID]")
	vaultLoginCmd.Flags().String("vault-kubernetes-path", "kubernetes", "Vault Kubernetes login mount path [GOGCI_VAULT_KUBERNETES_PATH]")
	vaultLoginCmd.Flags().String("vault-kubernetes-role", "", "Vault Kubernetes login role [GOGCI_VAULT_KUBERNETES_ROLE]")
	vaultLoginCmd.Flags().String("vault-jwt-path", "jwt", "Vault JWT login mount path [GOGCI_VAULT_JWT_PATH]")
	vaultLoginCmd.Flags().String("vault-jwt-role", "", "Vault JWT login role [GOGCI_VAULT_JWT_ROLE]")
	vaultLoginCmd.Flags().String("vault-jwt-token", "", "Vault JWT environment token [GOGCI_VAULT_JWT_TOKEN]")
	vaultLoginCmd.Flags().String("vault-jwt-token-envvar", "", "Vault JWT environment var name for token (override vault-jwt-token) [GOGCI_VAULT_JWT_TOKEN_ENVVAR]")
	vaultLoginCmd.Flags().Bool("export-token", false, "Export Vault Token [GOGCI_EXPORT_TOKEN]")

	vaultCmd.AddCommand(vaultLoginCmd)
}
