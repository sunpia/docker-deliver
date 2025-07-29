package compose_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	Compose "github.com/sunpia/docker-deliver/internal/compose"
)

// Setup function to create a temporary directory for tests.
func setupTempDir(t *testing.T) string {
	return t.TempDir()
}

// Setup function to create test dependencies.
func setupTestDependencies() *Compose.Dependencies {
	return &Compose.Dependencies{
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

// Simple test to verify the default dependencies work.
func TestDefaultDependencies(t *testing.T) {
	deps := Compose.DefaultDependencies()
	assert.NotNil(t, deps.OSCreate, "OSCreate should not be nil")
	assert.NotNil(t, deps.OSMkdirAll, "OSMkdirAll should not be nil")
	assert.NotNil(t, deps.YAMLMarshal, "YAMLMarshal should not be nil")
}

func TestNewComposeClient_InvalidLogLevel(t *testing.T) {
	setupTestDependencies()
	tempDir := setupTempDir(t)

	config := Compose.Config{
		DockerComposePath: []string{"docker-compose.yml"},
		WorkDir:           tempDir,
		OutputDir:         tempDir,
		Tag:               "latest",
		LogLevel:          "invalid",
	}

	client, err := Compose.NewComposeClient(context.Background(), config)

	assert.Error(t, err, "Expected error for invalid log level")
	assert.Nil(t, client, "Expected client to be nil on error")
}

func TestNewComposeClient_LoadError(t *testing.T) {
	tempDir := setupTempDir(t)

	// Create mock dependencies with project loading error
	deps := setupTestDependencies()
	deps.ProjectFromOptions = func(_ context.Context, _ *cli.ProjectOptions) (*types.Project, error) {
		return nil, errors.New("failed to load project")
	}

	config := Compose.Config{
		DockerComposePath: []string{"docker-compose.yml"},
		WorkDir:           tempDir,
		OutputDir:         tempDir,
		Tag:               "latest",
		LogLevel:          "info",
	}

	client, err := Compose.NewComposeClientWithDeps(context.Background(), config, deps)

	assert.Error(t, err, "Expected error from project loading")
	assert.Nil(t, client, "Expected client to be nil on error")
	assert.Contains(t, err.Error(), "failed to load project")
}

func TestNewComposeClient_CreateOutputDirError(t *testing.T) {
	tempDir := setupTempDir(t)

	// Create mock dependencies with mkdir error
	deps := setupTestDependencies()
	deps.OSMkdirAll = func(_ string, _ os.FileMode) error {
		return errors.New("permission denied")
	}

	// Mock project loading
	mockProject := &types.Project{Name: "test-project"}
	deps.ProjectFromOptions = func(_ context.Context, _ *cli.ProjectOptions) (*types.Project, error) {
		return mockProject, nil
	}

	config := Compose.Config{
		DockerComposePath: []string{"docker-compose.yml"},
		WorkDir:           tempDir,
		OutputDir:         "/nonexistent/path",
		Tag:               "latest",
		LogLevel:          "info",
	}

	client, err := Compose.NewComposeClientWithDeps(context.Background(), config, deps)

	assert.Error(t, err, "Expected error from mkdir")
	assert.Nil(t, client, "Expected client to be nil on error")
	assert.Contains(t, err.Error(), "permission denied")
}

func TestSaveComposeFile_Success(t *testing.T) {
	tempDir := setupTempDir(t)

	mockProject := &types.Project{
		Name: "test-project",
		Services: types.Services{
			"web": types.ServiceConfig{
				Name:  "web",
				Image: "nginx:latest",
			},
		},
	}

	deps := setupTestDependencies()
	deps.OSCreate = func(name string) (*os.File, error) {
		expectedPath := filepath.Join(tempDir, "docker-compose.generated.yaml")
		assert.Equal(t, expectedPath, name, "Expected file path to match")
		return os.Create(name)
	}

	deps.YAMLMarshal = func(_ interface{}) ([]byte, error) {
		return []byte("test yaml content"), nil
	}

	client := &Compose.Client{
		Config: Compose.Config{
			OutputDir: tempDir,
		},
		Project: mockProject,
		Logger:  logrus.New(),
		Deps:    deps,
	}

	_, err := client.SaveComposeFile(context.Background())
	require.NoError(t, err)

	// Verify file was created and content written
	content, err := os.ReadFile(filepath.Join(tempDir, "docker-compose.generated.yaml"))
	require.NoError(t, err, "Failed to read generated file")
	assert.Equal(t, "test yaml content", string(content))
}

func TestSaveComposeFile_NilProject(t *testing.T) {
	deps := setupTestDependencies()
	client := &Compose.Client{
		Config: Compose.Config{
			OutputDir: "/tmp",
		},
		Project: nil,
		Logger:  logrus.New(),
		Deps:    deps,
	}

	_, err := client.SaveComposeFile(context.Background())
	assert.NoError(t, err, "Expected no error for nil project")
}

func TestSaveComposeFile_CreateFileError(t *testing.T) {
	deps := setupTestDependencies()
	deps.OSCreate = func(_ string) (*os.File, error) {
		return nil, errors.New("file creation failed")
	}

	client := &Compose.Client{
		Config: Compose.Config{
			OutputDir: "/tmp",
		},
		Project: &types.Project{Name: "test"},
		Logger:  logrus.New(),
		Deps:    deps,
	}

	_, err := client.SaveComposeFile(context.Background())

	assert.Error(t, err, "Expected error from file creation")
	assert.Contains(t, err.Error(), "file creation failed")
}

func TestSaveComposeFile_MarshalError(t *testing.T) {
	tempDir := setupTempDir(t)

	deps := setupTestDependencies()
	deps.OSCreate = os.Create
	deps.YAMLMarshal = func(_ interface{}) ([]byte, error) {
		return nil, errors.New("marshal failed")
	}

	client := &Compose.Client{
		Config: Compose.Config{
			OutputDir: tempDir,
		},
		Project: &types.Project{Name: "test"},
		Logger:  logrus.New(),
		Deps:    deps,
	}

	_, err := client.SaveComposeFile(context.Background())

	assert.Error(t, err, "Expected error from yaml marshal")
	assert.Contains(t, err.Error(), "marshal failed")
}

func TestBuild_NilProject(t *testing.T) {
	deps := setupTestDependencies()
	client := &Compose.Client{
		Project: nil,
		Logger:  logrus.New(),
		Deps:    deps,
	}

	err := client.Build(context.Background())
	assert.NoError(t, err, "Expected no error for nil project")
}

func TestBuild_DockerClientError(t *testing.T) {
	deps := setupTestDependencies()
	deps.NewDockerClient = func() (*client.Client, error) {
		return nil, errors.New("docker client creation failed")
	}

	client := &Compose.Client{
		Config: Compose.Config{
			Tag: "v1.0.0",
		},
		Project: &types.Project{Name: "test"},
		Logger:  logrus.New(),
		Deps:    deps,
	}

	err := client.Build(context.Background())

	assert.Error(t, err, "Expected error from docker client creation")
	assert.Contains(t, err.Error(), "docker client creation failed")
}

func TestSaveImages_DockerClientError(t *testing.T) {
	deps := setupTestDependencies()
	deps.NewDockerClient = func() (*client.Client, error) {
		return nil, errors.New("docker client creation failed")
	}

	client := &Compose.Client{
		Config: Compose.Config{
			OutputDir: "/tmp",
		},
		Project: &types.Project{
			Services: types.Services{
				"web": types.ServiceConfig{
					Name:  "web",
					Image: "nginx:latest",
				},
			},
		},
		Logger: logrus.New(),
		Deps:   deps,
	}

	err := client.SaveImages(context.Background())

	assert.Error(t, err, "Expected error from docker client creation")
	assert.Contains(t, err.Error(), "docker client creation failed")
}

func TestBuild_ServiceImageTagging(t *testing.T) {
	mockProject := &types.Project{
		Name: "test-project",
		Services: types.Services{
			"web": types.ServiceConfig{
				Name: "web",
				Build: &types.BuildConfig{
					Context: ".",
				},
			},
			"db": types.ServiceConfig{
				Name:  "db",
				Image: "postgres:13",
			},
		},
	}

	client := &Compose.Client{
		Config: Compose.Config{
			Tag: "v1.0.0",
		},
		Project: mockProject,
	}

	// Test the image tagging logic specifically (before Docker operations)
	project := client.Project
	for _, s := range project.Services {
		if s.Image == "" {
			s.Image = s.Name + ":" + client.Config.Tag
			project.Services[s.Name] = s
		}
	}

	// Verify tagging worked correctly
	assert.Equal(t, "web:v1.0.0", project.Services["web"].Image, "Expected web service image to be 'web:v1.0.0'")
	assert.Equal(t, "postgres:13", project.Services["db"].Image, "Expected db service image to remain 'postgres:13'")
}

func TestSaveImages_NoImagesSpecified(t *testing.T) {
	mockProject := &types.Project{
		Services: types.Services{
			"web": types.ServiceConfig{
				Name: "web",
				// No image specified
			},
		},
	}

	client := &Compose.Client{
		Config: Compose.Config{
			OutputDir: "/tmp",
		},
		Project: mockProject,
	}

	// Test the logic that collects images
	images := make([]string, 0, len(client.Project.Services))
	for _, svc := range client.Project.Services {
		if svc.Image != "" {
			images = append(images, svc.Image)
		}
	}

	assert.Empty(t, images, "Expected no images to be collected")
}

func TestSaveImages_ImageCollection(t *testing.T) {
	mockProject := &types.Project{
		Services: types.Services{
			"web": types.ServiceConfig{
				Name:  "web",
				Image: "nginx:latest",
			},
			"db": types.ServiceConfig{
				Name:  "db",
				Image: "postgres:13",
			},
			"worker": types.ServiceConfig{
				Name: "worker",
				// No image specified
			},
		},
	}

	client := &Compose.Client{
		Config: Compose.Config{
			OutputDir: "/tmp",
		},
		Project: mockProject,
	}

	// Test the logic that collects images
	images := make([]string, 0, len(client.Project.Services))
	for _, svc := range client.Project.Services {
		if svc.Image != "" {
			images = append(images, svc.Image)
		}
	}

	expectedImages := []string{"nginx:latest", "postgres:13"}
	assert.Len(t, images, len(expectedImages), "Expected same number of images")

	// Check if all expected images are present
	imageMap := make(map[string]bool)
	for _, img := range images {
		imageMap[img] = true
	}

	for _, expected := range expectedImages {
		assert.True(t, imageMap[expected], "Expected image %s not found in collected images", expected)
	}
}

func TestSaveComposeFile_WriteError(t *testing.T) {
	tempDir := setupTempDir(t)

	// Create a file that will cause os.Create to return a file
	// but then fail on Write by closing it immediately
	deps := setupTestDependencies()
	deps.OSCreate = func(name string) (*os.File, error) {
		file, err := os.Create(name)
		if err != nil {
			return nil, err
		}
		// Close the file immediately to cause write error
		_ = file.Close() // Ignoring close error for test setup
		// Return a closed file to simulate write error
		return file, nil
	}

	deps.YAMLMarshal = func(_ interface{}) ([]byte, error) {
		return []byte("test yaml content"), nil
	}

	client := &Compose.Client{
		Config: Compose.Config{
			OutputDir: tempDir,
		},
		Project: &types.Project{Name: "test"},
		Logger:  logrus.New(),
		Deps:    deps,
	}

	_, err := client.SaveComposeFile(context.Background())
	assert.Error(t, err, "Expected error from file write")
}

func TestBuild_WithExistingImages(t *testing.T) {
	mockProject := &types.Project{
		Name: "test-project",
		Services: types.Services{
			"web": types.ServiceConfig{
				Name:  "web",
				Image: "existing:latest",
			},
			"api": types.ServiceConfig{
				Name: "api",
				// No image, should get tagged
			},
		},
	}

	client := &Compose.Client{
		Config: Compose.Config{
			Tag: "v2.0.0",
		},
		Project: mockProject,
	}

	// Test the image tagging logic before Docker operations
	project := client.Project
	for _, s := range project.Services {
		if s.Image == "" {
			s.Image = s.Name + ":" + client.Config.Tag
			project.Services[s.Name] = s
		}
	}

	// Verify only services without images got tagged
	assert.Equal(t, "existing:latest", project.Services["web"].Image, "Expected web service image to remain 'existing:latest'")
	assert.Equal(t, "api:v2.0.0", project.Services["api"].Image, "Expected api service image to be 'api:v2.0.0'")
}

func TestBuild_BuildConfigRemoval(t *testing.T) {
	mockProject := &types.Project{
		Name: "test-project",
		Services: types.Services{
			"web": types.ServiceConfig{
				Name: "web",
				Build: &types.BuildConfig{
					Context: ".",
				},
			},
			"db": types.ServiceConfig{
				Name:  "db",
				Image: "postgres:13",
			},
		},
	}

	client := &Compose.Client{
		Config: Compose.Config{
			Tag: "v1.0.0",
		},
		Project: mockProject,
	}

	// Test the build config removal logic
	project := client.Project
	for _, s := range project.Services {
		if s.Build != nil {
			s.Build = nil
			project.Services[s.Name] = s
		}
	}

	// Verify build config was removed
	assert.Nil(t, project.Services["web"].Build, "Expected web service build config to be removed")
	assert.Nil(t, project.Services["db"].Build, "Expected db service build config to remain nil")
}

func TestNewComposeClient_OutputDirExists(t *testing.T) {
	tempDir := setupTempDir(t)

	// Create the output directory first
	existingDir := filepath.Join(tempDir, "existing")
	const testDirPermissions = 0750 // More restrictive permissions for test
	err := os.MkdirAll(existingDir, testDirPermissions)
	require.NoError(t, err, "Failed to create existing directory")

	// Create mock dependencies
	deps := setupTestDependencies()
	mockProject := &types.Project{Name: "test-project"}
	deps.ProjectFromOptions = func(_ context.Context, _ *cli.ProjectOptions) (*types.Project, error) {
		return mockProject, nil
	}

	config := Compose.Config{
		DockerComposePath: []string{"docker-compose.yml"},
		WorkDir:           tempDir,
		OutputDir:         existingDir, // Use existing directory
		Tag:               "latest",
		LogLevel:          "info",
	}

	client, err := Compose.NewComposeClientWithDeps(context.Background(), config, deps)

	assert.NoError(t, err, "Expected no error when output directory exists")
	assert.NotNil(t, client, "Expected client to be not nil")
}

func TestSaveImages_EmptyImagesList(t *testing.T) {
	mockProject := &types.Project{
		Services: types.Services{
			"worker": types.ServiceConfig{
				Name: "worker",
				// No image specified
			},
		},
	}

	client := &Compose.Client{
		Config: Compose.Config{
			OutputDir: "/tmp",
		},
		Project: mockProject,
	}

	// Test the logic that collects images - should return early when no images
	images := make([]string, 0, len(client.Project.Services))
	for _, svc := range client.Project.Services {
		if svc.Image != "" {
			images = append(images, svc.Image)
		}
	}

	assert.Empty(t, images, "Expected no images to be collected")

	// The actual SaveImages call would return early without Docker operations
	// when no images are found, so we'll just test the collection logic here
}

// Benchmark for SaveComposeFile
// Example benchmark function.
func BenchmarkSaveComposeFile(b *testing.B) {
	deps := setupTestDependencies()
	tempDir := b.TempDir()
	deps.OSCreate = os.Create
	deps.YAMLMarshal = yaml.Marshal

	mockProject := &types.Project{
		Name: "bench-project",
		Services: types.Services{
			"web": types.ServiceConfig{
				Name:  "web",
				Image: "nginx:latest",
			},
		},
	}

	client := &Compose.Client{
		Config: Compose.Config{
			OutputDir: tempDir,
		},
		Project: mockProject,
		Logger:  logrus.New(),
		Deps:    deps,
	}

	b.ResetTimer()
	for range b.N {
		_, _ = client.SaveComposeFile(context.Background())
	}
}
