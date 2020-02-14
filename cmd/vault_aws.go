package cmd

import (
	"github.com/spf13/cobra"
)

// vaultAwsCmd represents the aws command
var vaultAwsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Vault AWS helpers",
}

func init() {
	vaultCmd.AddCommand(vaultAwsCmd)
}
