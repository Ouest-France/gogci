package cmd

import (
	"github.com/spf13/cobra"
)

// tfCmd represents the tf command
var tfCmd = &cobra.Command{
	Use:   "tf",
	Short: "Terraform helpers",
}

func init() {
	rootCmd.AddCommand(tfCmd)
}
