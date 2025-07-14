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

// ClientConfig holds configuration for ComposeClient.
type ClientConfig struct {
	DockerComposePath []string `json:"docker_compose_path"`
	WorkDir           string   `json:"work_dir"`
	OutputDir         string   `json:"output_dir"`
	Tag               string   `json:"tag"`      // Default tag for images
	LogLevel          string   `json:"loglevel"` // Log level: "debug", "info", "warn", "error"
}

// ClientInterface defines the main Compose actions.
type ClientInterface interface {
	SaveImages(ctx context.Context) error
	SaveComposeFile(ctx context.Context) error
	Build(ctx context.Context) error
}

// Dependencies holds all external dependencies for ComposeClient.
type Dependencies struct {
	OSCreate           func(string) (*os.File, error)
	OSMkdirAll         func(string, os.FileMode) error
	YAMLMarshal        func(interface{}) ([]byte, error)
	NewComposeService  func(*command.DockerCli) api.Service
	ProjectFromOptions func(context.Context, *cli.ProjectOptions) (*types.Project, error)
	NewDockerClient    func() (*client.Client, error)
	NewDockerCli       func(client.APIClient) (*command.DockerCli, error)
}

// DefaultDependencies returns the default production dependencies.
func DefaultDependencies() *Dependencies {
	return &Dependencies{
		OSCreate:    os.Create,
		OSMkdirAll:  os.MkdirAll,
		YAMLMarshal: yaml.Marshal,
		NewComposeService: func(cli *command.DockerCli) api.Service {
			return compose.NewComposeService(cli)
		},
		ProjectFromOptions: cli.ProjectFromOptions,
		NewDockerClient: func() (*client.Client, error) {
			return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		},
		NewDockerCli: func(apiClient client.APIClient) (*command.DockerCli, error) {
			return command.NewDockerCli(command.WithAPIClient(apiClient))
		},
	}
}

// Client implements ComposeInterface and holds project state.
type Client struct {
	ClientInterface // Interface embedding

	Config  ClientConfig
	Project *types.Project
	Logger  *logrus.Logger
	Deps    *Dependencies
}

// NewComposeClient creates and initializes a ComposeClient.
func NewComposeClient(ctx context.Context, config ClientConfig) (*Client, error) {
	return NewComposeClientWithDeps(ctx, config, DefaultDependencies())
}

// NewComposeClientWithDeps creates a ComposeClient with custom dependencies for testing.
func NewComposeClientWithDeps(ctx context.Context, config ClientConfig, deps *Dependencies) (*Client, error) {
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		return nil, err
	}

	c := &Client{
		Config: config,
		Logger: logrus.New(),
		Deps:   deps,
	}
	c.Logger.SetLevel(level)

	if loadErr := c.load(ctx); loadErr != nil {
		c.Logger.Errorf("Error loading compose file: %v", loadErr)
		return nil, loadErr
	}

	if _, statErr := os.Stat(c.Config.OutputDir); os.IsNotExist(statErr) {
		const dirPermissions = 0755
		if mkdirErr := c.Deps.OSMkdirAll(c.Config.OutputDir, dirPermissions); mkdirErr != nil {
			c.Logger.Errorf("Failed to create output directory: %v", mkdirErr)
			return nil, mkdirErr
		}
	}
	return c, nil
}

// load loads the compose project from the provided config.
func (c *Client) load(ctx context.Context) error {
	project, err := c.Deps.ProjectFromOptions(ctx, &cli.ProjectOptions{
		ConfigPaths: c.Config.DockerComposePath,
		WorkingDir:  c.Config.WorkDir,
		Environment: map[string]string{},
	})
	if err != nil {
		return err
	}
	c.Project = project
	return nil
}

// SaveComposeFile writes the current compose project to a YAML file.
func (c *Client) SaveComposeFile(_ context.Context) error {
	if c.Project == nil {
		return nil
	}
	outPath := c.Config.OutputDir + "/docker-compose.generated.yaml"
	file, err := c.Deps.OSCreate(outPath)
	if err != nil {
		c.Logger.Errorf("Failed to create compose file: %v", err)
		return err
	}
	defer file.Close()

	data, err := c.Deps.YAMLMarshal(c.Project)
	if err != nil {
		c.Logger.Errorf("Failed to marshal compose project: %v", err)
		return err
	}

	if _, writeErr := file.Write(data); writeErr != nil {
		c.Logger.Errorf("Failed to write compose file: %v", writeErr)
		return writeErr
	}
	c.Logger.Infof("Saved compose file to %s", outPath)
	return nil
}

// Build builds all services in the compose project.
func (c *Client) Build(ctx context.Context) error {
	project := c.Project
	if project == nil {
		return nil
	}

	for _, s := range project.Services {
		if s.Image == "" {
			s.Image = s.Name + ":" + c.Config.Tag
			project.Services[s.Name] = s
			c.Logger.Debugf("Tag Service %s image tag: %s", s.Name, s.Image)
		}
	}

	dockerClient, err := c.Deps.NewDockerClient()
	if err != nil {
		return err
	}
	defer dockerClient.Close()

	dockerCli, err := c.Deps.NewDockerCli(dockerClient)
	if err != nil {
		return err
	}

	if os.Getenv("OS") == "Windows_NT" {
		c.Logger.Debug("Configuring Docker environment for Windows desktop-linux context")
		_ = os.Setenv("DOCKER_HOST", "npipe:////./pipe/dockerDesktopLinuxEngine")
	}

	if initErr := dockerCli.Initialize(flags.NewClientOptions()); initErr != nil {
		return initErr
	}

	backend := c.Deps.NewComposeService(dockerCli)
	if backend == nil {
		return err
	}
	if buildErr := backend.Build(ctx, project, api.BuildOptions{}); buildErr != nil {
		c.Logger.Errorf("Failed to build project: %v", buildErr)
		return buildErr
	}

	for _, s := range project.Services {
		if s.Build != nil {
			s.Build = nil
			project.Services[s.Name] = s
		}
	}

	return nil
}

// SaveImages saves all images from the compose project to a tar archive.
func (c *Client) SaveImages(ctx context.Context) error {
	cli, err := c.Deps.NewDockerClient()
	if err != nil {
		c.Logger.Errorf("Error creating Docker client: %v", err)
		return err
	}
	defer cli.Close()

	images := make([]string, 0, len(c.Project.Services))
	for _, svc := range c.Project.Services {
		if svc.Image != "" {
			images = append(images, svc.Image)
		} else {
			c.Logger.Warnf("Service %s does not have an image specified.", svc.Name)
		}
	}

	if len(images) == 0 {
		c.Logger.Error("No images to save")
		return nil
	}

	imageSaveReader, err := cli.ImageSave(ctx, images)
	if err != nil {
		c.Logger.Errorf("Failed to save images: %v", err)
		return err
	}
	defer imageSaveReader.Close()

	outPath := c.Config.OutputDir + "/images.tar"
	outFile, err := os.Create(outPath)
	if err != nil {
		c.Logger.Errorf("Failed to create tar file for images: %v", err)
		return err
	}
	defer outFile.Close()

	if _, copyErr := io.Copy(outFile, imageSaveReader); copyErr != nil {
		c.Logger.Errorf("Failed to write image tar: %v", copyErr)
		return copyErr
	}
	fi, err := outFile.Stat()
	if err != nil {
		c.Logger.Warnf("Could not get file size for %s: %v", outPath, err)
	} else {
		const bytesToGB = 1024 * 1024 * 1024
		sizeGB := float64(fi.Size()) / bytesToGB
		c.Logger.Infof("Saved images to %s (%.2f GB)", outPath, sizeGB)
	}
	return nil
}
