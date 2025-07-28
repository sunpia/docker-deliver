package commands

import (
	"github.com/spf13/cobra"
	mcp "github.com/sunpia/docker-deliver/internal/mcp"
)

func NewMCPCmd() *cobra.Command {
	var (
		httpAddr string
	)
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start a mcp server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			config := mcp.Config{
				HttpAddr: httpAddr,
			}
			ctx := cmd.Context()

			client, err := mcp.NewClient(ctx, config)
			if err != nil {
				return err
			}
			if runErr := client.Run(ctx); runErr != nil {
				return runErr
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&httpAddr, "http", "H", "", "HTTP address")
	_ = cmd.MarkFlagRequired("file") // Error handling: ignoring error for required flag

	return cmd
}
