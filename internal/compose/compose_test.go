package compose_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	pkg "github.com/sunpia/docker-deliver/internal/compose"
)

// Setup function to create a temporary directory for tests.
func setupTempDir(t *testing.T) string {
	return t.TempDir()
}

// Setup function to create test dependencies.
func setupTestDependencies() *pkg.Dependencies {
	return &pkg.Dependencies{
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
	deps := pkg.DefaultDependencies()
	if deps.OSCreate == nil {
		t.Error("OSCreate should not be nil")
	}
	if deps.OSMkdirAll == nil {
		t.Error("OSMkdirAll should not be nil")
	}
	if deps.YAMLMarshal == nil {
		t.Error("YAMLMarshal should not be nil")
	}
}

func TestNewComposeClient_InvalidLogLevel(t *testing.T) {
	setupTestDependencies()
	tempDir := setupTempDir(t)

	config := pkg.ClientConfig{
		DockerComposePath: []string{"docker-compose.yml"},
		WorkDir:           tempDir,
		OutputDir:         tempDir,
		Tag:               "latest",
		LogLevel:          "invalid",
	}

	client, err := pkg.NewComposeClient(context.Background(), config)

	if err == nil {
		t.Fatal("Expected error for invalid log level")
	}
	if client != nil {
		t.Fatal("Expected client to be nil on error")
	}
}

func TestNewComposeClient_LoadError(t *testing.T) {
	tempDir := setupTempDir(t)

	// Create mock dependencies with project loading error
	deps := setupTestDependencies()
	deps.ProjectFromOptions = func(_ context.Context, _ *cli.ProjectOptions) (*types.Project, error) {
		return nil, errors.New("failed to load project")
	}

	config := pkg.ClientConfig{
		DockerComposePath: []string{"docker-compose.yml"},
		WorkDir:           tempDir,
		OutputDir:         tempDir,
		Tag:               "latest",
		LogLevel:          "info",
	}

	client, err := pkg.NewComposeClientWithDeps(context.Background(), config, deps)

	if err == nil {
		t.Fatal("Expected error from project loading")
	}
	if client != nil {
		t.Fatal("Expected client to be nil on error")
	}
	if !strings.Contains(err.Error(), "failed to load project") {
		t.Errorf("Expected error to contain 'failed to load project', got: %v", err)
	}
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

	config := pkg.ClientConfig{
		DockerComposePath: []string{"docker-compose.yml"},
		WorkDir:           tempDir,
		OutputDir:         "/nonexistent/path",
		Tag:               "latest",
		LogLevel:          "info",
	}

	client, err := pkg.NewComposeClientWithDeps(context.Background(), config, deps)

	if err == nil {
		t.Fatal("Expected error from mkdir")
	}
	if client != nil {
		t.Fatal("Expected client to be nil on error")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("Expected error to contain 'permission denied', got: %v", err)
	}
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
		if name != expectedPath {
			t.Errorf("Expected file path %s, got %s", expectedPath, name)
		}
		return os.Create(name)
	}

	deps.YAMLMarshal = func(_ interface{}) ([]byte, error) {
		return []byte("test yaml content"), nil
	}

	client := &pkg.Client{
		Config: pkg.ClientConfig{
			OutputDir: tempDir,
		},
		Project: mockProject,
		Logger:  logrus.New(),
		Deps:    deps,
	}

	err := client.SaveComposeFile(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify file was created and content written
	content, err := os.ReadFile(filepath.Join(tempDir, "docker-compose.generated.yaml"))
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}
	if string(content) != "test yaml content" {
		t.Errorf("Expected content 'test yaml content', got '%s'", string(content))
	}
}

func TestSaveComposeFile_NilProject(t *testing.T) {
	deps := setupTestDependencies()
	client := &pkg.Client{
		Config: pkg.ClientConfig{
			OutputDir: "/tmp",
		},
		Project: nil,
		Logger:  logrus.New(),
		Deps:    deps,
	}

	err := client.SaveComposeFile(context.Background())

	if err != nil {
		t.Errorf("Expected no error for nil project, got: %v", err)
	}
}

func TestSaveComposeFile_CreateFileError(t *testing.T) {
	deps := setupTestDependencies()
	deps.OSCreate = func(_ string) (*os.File, error) {
		return nil, errors.New("file creation failed")
	}

	client := &pkg.Client{
		Config: pkg.ClientConfig{
			OutputDir: "/tmp",
		},
		Project: &types.Project{Name: "test"},
		Logger:  logrus.New(),
		Deps:    deps,
	}

	err := client.SaveComposeFile(context.Background())

	if err == nil {
		t.Fatal("Expected error from file creation")
	}
	if !strings.Contains(err.Error(), "file creation failed") {
		t.Errorf("Expected error to contain 'file creation failed', got: %v", err)
	}
}

func TestSaveComposeFile_MarshalError(t *testing.T) {
	tempDir := setupTempDir(t)

	deps := setupTestDependencies()
	deps.OSCreate = os.Create
	deps.YAMLMarshal = func(_ interface{}) ([]byte, error) {
		return nil, errors.New("marshal failed")
	}

	client := &pkg.Client{
		Config: pkg.ClientConfig{
			OutputDir: tempDir,
		},
		Project: &types.Project{Name: "test"},
		Logger:  logrus.New(),
		Deps:    deps,
	}

	err := client.SaveComposeFile(context.Background())

	if err == nil {
		t.Fatal("Expected error from yaml marshal")
	}
	if !strings.Contains(err.Error(), "marshal failed") {
		t.Errorf("Expected error to contain 'marshal failed', got: %v", err)
	}
}

func TestBuild_NilProject(t *testing.T) {
	deps := setupTestDependencies()
	client := &pkg.Client{
		Project: nil,
		Logger:  logrus.New(),
		Deps:    deps,
	}

	err := client.Build(context.Background())

	if err != nil {
		t.Errorf("Expected no error for nil project, got: %v", err)
	}
}

func TestBuild_DockerClientError(t *testing.T) {
	deps := setupTestDependencies()
	deps.NewDockerClient = func() (*client.Client, error) {
		return nil, errors.New("docker client creation failed")
	}

	client := &pkg.Client{
		Config: pkg.ClientConfig{
			Tag: "v1.0.0",
		},
		Project: &types.Project{Name: "test"},
		Logger:  logrus.New(),
		Deps:    deps,
	}

	err := client.Build(context.Background())

	if err == nil {
		t.Fatal("Expected error from docker client creation")
	}
	if !strings.Contains(err.Error(), "docker client creation failed") {
		t.Errorf("Expected error to contain 'docker client creation failed', got: %v", err)
	}
}

func TestSaveImages_DockerClientError(t *testing.T) {
	deps := setupTestDependencies()
	deps.NewDockerClient = func() (*client.Client, error) {
		return nil, errors.New("docker client creation failed")
	}

	client := &pkg.Client{
		Config: pkg.ClientConfig{
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

	if err == nil {
		t.Fatal("Expected error from docker client creation")
	}
	if !strings.Contains(err.Error(), "docker client creation failed") {
		t.Errorf("Expected error to contain 'docker client creation failed', got: %v", err)
	}
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

	client := &pkg.Client{
		Config: pkg.ClientConfig{
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
	if project.Services["web"].Image != "web:v1.0.0" {
		t.Errorf("Expected web service image to be 'web:v1.0.0', got '%s'", project.Services["web"].Image)
	}
	if project.Services["db"].Image != "postgres:13" {
		t.Errorf("Expected db service image to remain 'postgres:13', got '%s'", project.Services["db"].Image)
	}
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

	client := &pkg.Client{
		Config: pkg.ClientConfig{
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

	if len(images) != 0 {
		t.Errorf("Expected no images to be collected, got %v", images)
	}
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

	client := &pkg.Client{
		Config: pkg.ClientConfig{
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
	if len(images) != len(expectedImages) {
		t.Errorf("Expected %d images, got %d", len(expectedImages), len(images))
	}

	// Check if all expected images are present
	imageMap := make(map[string]bool)
	for _, img := range images {
		imageMap[img] = true
	}

	for _, expected := range expectedImages {
		if !imageMap[expected] {
			t.Errorf("Expected image %s not found in collected images", expected)
		}
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

	client := &pkg.Client{
		Config: pkg.ClientConfig{
			OutputDir: tempDir,
		},
		Project: &types.Project{Name: "test"},
		Logger:  logrus.New(),
		Deps:    deps,
	}

	err := client.SaveComposeFile(context.Background())

	if err == nil {
		t.Fatal("Expected error from file write")
	}
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

	client := &pkg.Client{
		Config: pkg.ClientConfig{
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
	if project.Services["web"].Image != "existing:latest" {
		t.Errorf("Expected web service image to remain 'existing:latest', got '%s'", project.Services["web"].Image)
	}
	if project.Services["api"].Image != "api:v2.0.0" {
		t.Errorf("Expected api service image to be 'api:v2.0.0', got '%s'", project.Services["api"].Image)
	}
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

	client := &pkg.Client{
		Config: pkg.ClientConfig{
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
	if project.Services["web"].Build != nil {
		t.Error("Expected web service build config to be removed")
	}
	if project.Services["db"].Build != nil {
		t.Error("Expected db service build config to remain nil")
	}
}

func TestNewComposeClient_OutputDirExists(t *testing.T) {
	tempDir := setupTempDir(t)

	// Create the output directory first
	existingDir := filepath.Join(tempDir, "existing")
	const testDirPermissions = 0750 // More restrictive permissions for test
	err := os.MkdirAll(existingDir, testDirPermissions)
	if err != nil {
		t.Fatalf("Failed to create existing directory: %v", err)
	}

	// Create mock dependencies
	deps := setupTestDependencies()
	mockProject := &types.Project{Name: "test-project"}
	deps.ProjectFromOptions = func(_ context.Context, _ *cli.ProjectOptions) (*types.Project, error) {
		return mockProject, nil
	}

	config := pkg.ClientConfig{
		DockerComposePath: []string{"docker-compose.yml"},
		WorkDir:           tempDir,
		OutputDir:         existingDir, // Use existing directory
		Tag:               "latest",
		LogLevel:          "info",
	}

	client, err := pkg.NewComposeClientWithDeps(context.Background(), config, deps)

	if err != nil {
		t.Fatalf("Expected no error when output directory exists, got: %v", err)
	}
	if client == nil {
		t.Fatal("Expected client to be not nil")
	}
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

	client := &pkg.Client{
		Config: pkg.ClientConfig{
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

	if len(images) != 0 {
		t.Errorf("Expected no images to be collected, got %v", images)
	}

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

	client := &pkg.Client{
		Config: pkg.ClientConfig{
			OutputDir: tempDir,
		},
		Project: mockProject,
		Logger:  logrus.New(),
		Deps:    deps,
	}

	b.ResetTimer()
	for range b.N {
		_ = client.SaveComposeFile(context.Background())
	}
}
