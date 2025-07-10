package commands

import (
	"github.com/spf13/cobra"
	Compose "github.com/sunpia/docker-deliver/internal/compose"
)

func NewSaveCmd() *cobra.Command {
	var tag, logLevel string
	var verbose bool
	var outputDir string
	var dockerComposePath string
	var workDir string

	saveCmd := &cobra.Command{
		Use:   "save",
		Short: "Save docker compose project",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := Compose.ComposeConfig{
				DockerComposePath: []string{dockerComposePath},
				WorkDir:           workDir,
				OutputDir:         outputDir,
				Tag:               tag,
				LogLevel:          logLevel,
				Verbose:           verbose,
			}
			ctx := cmd.Context()

			client, err := Compose.NewComposeClient(ctx, config)
			if err != nil {
				return err
			}
			err = client.BuildImage(ctx)
			if err != nil {
				return err
			}
			err = client.SaveImage(ctx)
			if err != nil {
				return err
			}
			if err := client.SaveComposeFile(ctx); err != nil {
				return err
			}

			return nil
		},
	}

	saveCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory (required)")
	saveCmd.Flags().StringVarP(&dockerComposePath, "file", "f", "", "Path to docker-compose file (required)")
	saveCmd.Flags().StringVarP(&workDir, "workdir", "w", "", "Working directory (optional)")
	saveCmd.Flags().StringVar(&tag, "tag", "", "Default tag for images (optional)")
	saveCmd.Flags().StringVar(&logLevel, "loglevel", "info", "Log level: debug, info, warn, error (optional)")
	saveCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose output (optional)")
	saveCmd.MarkFlagRequired("file")

	return saveCmd
}
