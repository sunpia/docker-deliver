package root

import (
	"github.com/spf13/cobra"
	"github.com/sunpia/docker-deliver/cmd/commands"
)

var rootCmd = &cobra.Command{
	Use:   "docker-deliver",
	Short: "Docker-deliver is a deployment tool",
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.AddCommand(commands.NewSaveCmd())
}
