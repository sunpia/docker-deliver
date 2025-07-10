package compose

import (
	"context"
	"io"
	"os"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// ComposeConfig holds configuration for ComposeClient.
type ComposeConfig struct {
	DockerComposePath []string `json:"docker_compose_path"`
	WorkDir           string   `json:"work_dir"`
	OutputDir         string   `json:"output_dir"`
	Tag               string   `json:"tag"`      // Default tag for images
	LogLevel          string   `json:"loglevel"` // Log level: "debug", "info", "warn", "error"
}

// ComposeInterface defines the main Compose actions.
type ComposeInterface interface {
	SaveImages(ctx context.Context) error
	SaveComposeFile(ctx context.Context) error
	Build(ctx context.Context) error
}

// ComposeClient implements ComposeInterface and holds project state.
type ComposeClient struct {
	Config           ComposeConfig
	ComposeInterface // Not strictly necessary, but kept for interface compliance
	Project          *types.Project
	logger           *logrus.Logger
}

// NewComposeClient creates and initializes a ComposeClient.
func NewComposeClient(ctx context.Context, config ComposeConfig) (*ComposeClient, error) {
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}

	c := &ComposeClient{
		Config: config,
		logger: logrus.New(),
	}
	c.logger.SetLevel(level)

	if err := c.load(ctx); err != nil {
		c.logger.Errorf("Error loading compose file: %v", err)
		return nil, err
	}

	if _, err := os.Stat(c.Config.OutputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(c.Config.OutputDir, 0755); err != nil {
			c.logger.Errorf("Failed to create output directory: %v", err)
			return nil, err
		}
	}
	return c, nil
}

// load loads the compose project from the provided config.
func (c *ComposeClient) load(ctx context.Context) error {
	project, err := cli.ProjectFromOptions(ctx, &cli.ProjectOptions{
		ConfigPaths: c.Config.DockerComposePath,
		WorkingDir:  c.Config.WorkDir,
		Environment: map[string]string{},
	})
	if err != nil {
		c.logger.Errorf("Error loading compose file: %v", err)
		os.Exit(1)
	}
	c.Project = project
	return nil
}

// SaveComposeFile writes the current compose project to a YAML file.
func (c *ComposeClient) SaveComposeFile(ctx context.Context) error {
	if c.Project == nil {
		return nil
	}
	outPath := c.Config.OutputDir + "/docker-compose.generated.yaml"
	file, err := os.Create(outPath)
	if err != nil {
		c.logger.Errorf("Failed to create compose file: %v", err)
		return err
	}
	defer file.Close()

	data, err := yaml.Marshal(c.Project)
	if err != nil {
		c.logger.Errorf("Failed to marshal compose project: %v", err)
		return err
	}

	if _, err := file.Write(data); err != nil {
		c.logger.Errorf("Failed to write compose file: %v", err)
		return err
	}
	c.logger.Infof("Saved compose file to %s", outPath)
	return nil
}

// Build builds all services in the compose project.
func (c *ComposeClient) Build(ctx context.Context) error {
	project := c.Project
	if project == nil {
		return nil
	}

	// Ensure all services have an image tag
	for _, s := range project.Services {
		if s.Image == "" {
			s.Image = s.Name + ":" + c.Config.Tag
			project.Services[s.Name] = s
			c.logger.Debugf("Tag Service %s image tag: %s", s.Name, s.Image)
		}
	}

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer dockerClient.Close()

	dockerCli, err := command.NewDockerCli(command.WithAPIClient(dockerClient))
	if err != nil {
		return err
	}

	if err := dockerCli.Initialize(flags.NewClientOptions()); err != nil {
		return err
	}
	backend := compose.NewComposeService(dockerCli)
	if backend == nil {
		return err
	}
	if err := backend.Build(ctx, project, api.BuildOptions{}); err != nil {
		c.logger.Errorf("Failed to build project: %v", err)
		return err
	}

	// Remove build context from services
	for _, s := range project.Services {
		if s.Build != nil {
			s.Build = nil
			project.Services[s.Name] = s
		}
	}

	return nil
}

// SaveImages saves all images from the compose project to a tar archive.
func (c *ComposeClient) SaveImages(ctx context.Context) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		c.logger.Errorf("Error creating Docker client: %v", err)
		return err
	}
	defer cli.Close()

	images := make([]string, 0, len(c.Project.Services))
	for _, svc := range c.Project.Services {
		if svc.Image != "" {
			images = append(images, svc.Image)
		} else {
			c.logger.Warnf("Service %s does not have an image specified.", svc.Name)
		}
	}

	if len(images) == 0 {
		c.logger.Error("No images to save")
		return nil
	}

	imageSaveReader, err := cli.ImageSave(ctx, images)
	if err != nil {
		c.logger.Errorf("Failed to save images: %v", err)
		return err
	}
	defer imageSaveReader.Close()

	outPath := c.Config.OutputDir + "/images.tar"
	outFile, err := os.Create(outPath)
	if err != nil {
		c.logger.Errorf("Failed to create tar file for images: %v", err)
		return err
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, imageSaveReader); err != nil {
		c.logger.Errorf("Failed to write image tar: %v", err)
		return err
	}
	fi, err := outFile.Stat()
	if err != nil {
		c.logger.Warnf("Could not get file size for %s: %v", outPath, err)
	} else {
		sizeGB := float64(fi.Size()) / (1024 * 1024 * 1024)
		c.logger.Infof("Saved images to %s (%.2f GB)", outPath, sizeGB)
	}
	return nil
}
