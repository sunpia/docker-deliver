package Compose

import (
	"context"
	"os"

	"bufio"
	"encoding/json"
	"io"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"gopkg.in/yaml.v3"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"

	"github.com/sirupsen/logrus"
)

type ComposeConfig struct {
	DockerComposePath []string `json:"docker_compose_path"`
	WorkDir           string   `json:"work_dir"`
	OutputDir         string   `json:"output_dir"`
	Tag               string   `json:"tag"`      // Default tag for images
	LogLevel          string   `json:"loglevel"` // Log level: "debug", "info", "warn", "error"
	Verbose           bool     `json:"verbose"`  // If true, enables verbose output
}

type ComposeInterface interface {
	load(ctx context.Context) error
	sortServices(ctx context.Context) error
	SaveImage(ctx context.Context) error
	SaveComposeFile(ctx context.Context) error
	BuildImage(ctx context.Context) error
}

type ComposeClient struct {
	Config ComposeConfig
	ComposeInterface
	Project        *types.Project
	SortedServices []*types.ServiceConfig // Changed to list of list
	logger         *logrus.Logger
}

func (c *ComposeClient) SaveComposeFile(ctx context.Context) error {
	if c.Project == nil {
		return nil
	}
	outPath := c.Config.OutputDir + "/docker-compose.generated.yaml"
	file, err := os.Create(outPath)
	if err != nil {
		c.logger.Errorf("failed to create compose file: %v", err)
		return err
	}
	defer file.Close()
	data, err := yaml.Marshal(c.Project)
	if err != nil {
		c.logger.Errorf("failed to marshal compose project: %v", err)
		return err
	}

	if _, err := file.Write(data); err != nil {
		c.logger.Errorf("failed to write compose file: %v", err)
		return err
	}
	c.logger.Infof("Saved compose file to %s", outPath)
	return nil
}

func NewComposeClient(ctx context.Context, config ComposeConfig) (*ComposeClient, error) {
	// Set c.logger level based on config.LogLevel
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

	if err := c.sortServices(ctx); err != nil {
		c.logger.Errorf("Error sorting services: %v", err)
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

func (c *ComposeClient) sortServices(ctx context.Context) error {
	type node struct {
		name     string
		deps     []string
		visited  bool
		tempMark bool
		svc      *types.ServiceConfig
	}

	serviceMap := make(map[string]*node)
	for _, svc := range c.Project.Services {
		deps := make([]string, 0)
		if svc.DependsOn != nil {
			for dep := range svc.DependsOn {
				deps = append(deps, dep)
			}
		}
		serviceMap[svc.Name] = &node{name: svc.Name, deps: deps, svc: &svc}
	}

	var sorted []*types.ServiceConfig
	var visit func(n *node) error
	visit = func(n *node) error {
		if n.tempMark {
			c.logger.Errorf("circular dependency detected at %s", n.name)
			return nil
		}
		if !n.visited {
			n.tempMark = true
			for _, dep := range n.deps {
				if depNode, ok := serviceMap[dep]; ok {
					if err := visit(depNode); err != nil {
						return err
					}
				}
			}
			n.visited = true
			n.tempMark = false
			sorted = append(sorted, n.svc)
		}
		return nil
	}

	for _, n := range serviceMap {
		if err := visit(n); err != nil {
			return err
		}
	}

	c.SortedServices = sorted

	return nil
}

func createTarFromDir(dir string) (io.ReadCloser, error) {
	return archive.TarWithOptions(dir, &archive.TarOptions{})
}

func (c *ComposeClient) SaveImage(ctx context.Context) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		c.logger.Errorf("error creating Docker client: %v", err)
		return err
	}
	defer cli.Close()
	images := make([]string, len(c.Project.Services))
	cnt := 0
	for _, svc := range c.Project.Services {
		if svc.Image != "" {
			images[cnt] = svc.Image
		} else {
			c.logger.Warnf("Service %s does not have an image specified.", svc.Name)
		}
		cnt += 1
	}

	if len(images) == 0 {
		c.logger.Error("no images to save")
		return nil
	}

	imageSaveReader, err := cli.ImageSave(ctx, images)
	if err != nil {
		c.logger.Errorf("failed to save images: %v", err)
		return err
	}
	defer imageSaveReader.Close()

	outPath := c.Config.OutputDir + "/images.tar"
	outFile, err := os.Create(outPath)
	if err != nil {
		c.logger.Errorf("failed to create tar file for images: %v", err)
		return err
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, imageSaveReader); err != nil {
		c.logger.Errorf("failed to write image tar: %v", err)
		return err
	}
	c.logger.Infof("Saved images to %s", outPath)
	return nil
}

func (c *ComposeClient) BuildImage(ctx context.Context) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	tags := c.Config.Tag
	if tags == "" {
		tags = "latest"
	}

	for _, svc := range c.SortedServices {
		if svc.Build != nil {
			buildCtx, err := createTarFromDir(svc.Build.Context)
			if err != nil {
				c.logger.Errorf("Error opening build context: %v", err)
				continue
			}
			if svc.Build.Tags == nil || len(svc.Build.Tags) == 0 {
				if svc.Image == "" {
					svc.Build.Tags = []string{svc.Name + ":" + tags}
				} else {
					svc.Build.Tags = []string{svc.Image}
				}
			}

			buildOptions := build.ImageBuildOptions{
				BuildArgs:  svc.Build.Args,
				Dockerfile: svc.Build.Dockerfile,
				Target:     svc.Build.Target,
				Tags:       svc.Build.Tags,
				NoCache:    svc.Build.NoCache,
				Labels:     svc.Build.Labels,
				CacheFrom:  svc.Build.CacheFrom,
			}

			defer buildCtx.Close()
			resp, err := cli.ImageBuild(ctx, buildCtx, buildOptions)
			if err != nil {
				c.logger.Errorf("Failed to build image for %s: %v", svc.Name, err)
				continue
			}
			defer resp.Body.Close()
			c.logger.Infof("Building image for service: %s", svc.Name)
			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				var msg struct {
					Stream      string `json:"stream"`
					ErrorDetail struct {
						Message string `json:"message"`
					} `json:"errorDetail"`
				}
				line := scanner.Text()
				if err := json.Unmarshal([]byte(line), &msg); err != nil {
					c.logger.Errorf("Error: %v", err)
					continue
				}
				c.logger.Info(msg.Stream)
				if msg.ErrorDetail.Message != "" {
					c.logger.Errorf("Error building image for %s: %s", svc.Name, msg.ErrorDetail.Message)
					os.Exit(1)
				}
			}
			svcCopy := c.Project.Services[svc.Name]
			svcCopy.Image = svc.Build.Tags[0]
			svcCopy.Build = nil // Clear build context after building
			c.Project.Services[svc.Name] = svcCopy
		}
	}
	return nil
}
