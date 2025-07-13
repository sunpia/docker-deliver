package main

import (
	"github.com/spf13/cobra"
	"github.com/sunpia/docker-deliver/cmd/commands"
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "docker-deliver",
		Short: "Docker-deliver is a deployment tool",
	}
	rootCmd.AddCommand(commands.NewSaveCmd())
	return rootCmd
}

func Execute() {
	cobra.CheckErr(newRootCmd().Execute())
}
