package cmd

import (
	"github.com/Ouest-France/gogci/command"
	"github.com/spf13/cobra"
)

// tfInitCmd represents the "tf init" command
var tfInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Launch terraform init",
	RunE: func(cmd *cobra.Command, args []string) error {

		_, _, _, err := command.Run("terraform", []string{"init"})
		return err
	},
}

func init() {
	tfCmd.AddCommand(tfInitCmd)
}
