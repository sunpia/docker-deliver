package commands

import (
	"github.com/spf13/cobra"
	Compose "github.com/sunpia/docker-deliver/internal/compose"
)

func NewSaveCmd() *cobra.Command {
	var (
		tag               string
		logLevel          string
		outputDir         string
		dockerComposePath []string
		workDir           string
	)

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save docker compose project",
		RunE: func(cmd *cobra.Command, _ []string) error {
			config := Compose.ComposeConfig{
				DockerComposePath: dockerComposePath,
				WorkDir:           workDir,
				OutputDir:         outputDir,
				Tag:               tag,
				LogLevel:          logLevel,
			}
			ctx := cmd.Context()

			client, err := Compose.NewComposeClient(ctx, config)
			if err != nil {
				return err
			}
			if buildErr := client.Build(ctx); buildErr != nil {
				return buildErr
			}
			if saveErr := client.SaveImages(ctx); saveErr != nil {
				return saveErr
			}
			if composeErr := client.SaveComposeFile(ctx); composeErr != nil {
				return composeErr
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory (required)")
	cmd.Flags().StringSliceVarP(&dockerComposePath, "file", "f", nil, "Path to docker-compose file (required)")
	cmd.Flags().StringVarP(&workDir, "workdir", "w", "", "Working directory (optional)")
	cmd.Flags().StringVarP(&tag, "tag", "t", "latest", "Default tag for images (optional)")
	cmd.Flags().StringVarP(&logLevel, "loglevel", "l", "info", "Log level: debug, info, warn, error (optional)")
	_ = cmd.MarkFlagRequired("file") // Error handling: ignoring error for required flag

	return cmd
}
