package compose_test

import (
	"context"
	"fmt"
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
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	Compose "github.com/sunpia/docker-deliver/internal/compose"
)

// setupBenchmarkProject creates a mock project for benchmarking.
func setupBenchmarkProject(numServices int) *types.Project {
	services := make(types.Services)

	for i := range numServices {
		serviceName := fmt.Sprintf("service-%d", i)
		services[serviceName] = types.ServiceConfig{
			Name:  serviceName,
			Image: fmt.Sprintf("nginx:latest-%d", i),
			Build: &types.BuildConfig{
				Context: ".",
			},
		}
	}

	return &types.Project{
		Name:     "benchmark-project",
		Services: services,
	}
}

func setupBenchmarkDependencies() *Compose.Dependencies {
	return &Compose.Dependencies{
		OSCreate:    os.Create,
		OSMkdirAll:  os.MkdirAll,
		YAMLMarshal: yaml.Marshal,
		NewComposeService: func(cli *command.DockerCli) api.Service {
			return compose.NewComposeService(cli)
		},
		ProjectFromOptions: func(_ context.Context, _ *cli.ProjectOptions) (*types.Project, error) {
			// Return a mock project for benchmarking
			return setupBenchmarkProject(10), nil
		},
		NewDockerClient: func() (*client.Client, error) {
			return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		},
		NewDockerCli: func(apiClient client.APIClient) (*command.DockerCli, error) {
			return command.NewDockerCli(command.WithAPIClient(apiClient))
		},
	}
}

// BenchmarkNewComposeClient benchmarks client creation.
func BenchmarkNewComposeClient(b *testing.B) {
	tempDir := b.TempDir()
	config := Compose.Config{
		DockerComposePath: []string{"docker-compose.yml"},
		WorkDir:           tempDir,
		OutputDir:         tempDir,
		Tag:               "latest",
		LogLevel:          "info",
	}
	deps := setupBenchmarkDependencies()

	b.ResetTimer()
	for range b.N {
		client, err := Compose.NewComposeClientWithDeps(context.Background(), config, deps)
		if err != nil {
			b.Fatalf("Failed to create client: %v", err)
		}
		_ = client // Use the client to prevent optimization
	}
}

// BenchmarkSaveComposeFile_MultipleServices benchmarks compose file saving with various service counts.
func BenchmarkSaveComposeFile_MultipleServices(b *testing.B) {
	tests := []struct {
		name        string
		numServices int
	}{
		{"Small_5_Services", 5},
		{"Medium_20_Services", 20},
		{"Large_50_Services", 50},
		{"XLarge_100_Services", 100},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			tempDir := b.TempDir()
			deps := setupBenchmarkDependencies()
			project := setupBenchmarkProject(tt.numServices)

			client := &Compose.Client{
				Config: Compose.Config{
					OutputDir: tempDir,
				},
				Project: project,
				Logger:  logrus.New(),
				Deps:    deps,
			}

			b.ResetTimer()
			for range b.N {
				_, err := client.SaveComposeFile(context.Background())
				if err != nil {
					b.Fatalf("Failed to save compose file: %v", err)
				}
			}
		})
	}
}

// BenchmarkBuild_ServiceTagging benchmarks the service image tagging logic.
func BenchmarkBuild_ServiceTagging(b *testing.B) {
	tests := []struct {
		name        string
		numServices int
	}{
		{"Small_5_Services", 5},
		{"Medium_20_Services", 20},
		{"Large_50_Services", 50},
		{"XLarge_100_Services", 100},
		{"XXLarge_500_Services", 500},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			tag := "v1.0.0"

			b.ResetTimer()
			for range b.N {
				b.StopTimer()
				project := setupBenchmarkProject(tt.numServices)
				// Remove images to force tagging
				for name, service := range project.Services {
					service.Image = ""
					project.Services[name] = service
				}
				b.StartTimer()

				// Benchmark the tagging logic
				for _, s := range project.Services {
					if s.Image == "" {
						s.Image = s.Name + ":" + tag
						project.Services[s.Name] = s
					}
				}
			}
		})
	}
}

// BenchmarkBuild_BuildConfigRemoval benchmarks removing build configs.
func BenchmarkBuild_BuildConfigRemoval(b *testing.B) {
	tests := []struct {
		name        string
		numServices int
	}{
		{"Small_5_Services", 5},
		{"Medium_20_Services", 20},
		{"Large_50_Services", 50},
		{"XLarge_100_Services", 100},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for range b.N {
				b.StopTimer()
				project := setupBenchmarkProject(tt.numServices)
				b.StartTimer()

				// Benchmark the build config removal logic
				for _, s := range project.Services {
					if s.Build != nil {
						s.Build = nil
						project.Services[s.Name] = s
					}
				}
			}
		})
	}
}

// BenchmarkImageCollection benchmarks image collection from services.
func BenchmarkImageCollection(b *testing.B) {
	tests := []struct {
		name        string
		numServices int
	}{
		{"Small_5_Services", 5},
		{"Medium_20_Services", 20},
		{"Large_50_Services", 50},
		{"XLarge_100_Services", 100},
		{"XXLarge_500_Services", 500},
		{"XXXLarge_1000_Services", 1000},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			project := setupBenchmarkProject(tt.numServices)

			b.ResetTimer()
			for range b.N {
				// Benchmark the image collection logic
				images := make([]string, 0, len(project.Services))
				for _, svc := range project.Services {
					if svc.Image != "" {
						images = append(images, svc.Image)
					}
				}
				_ = images // Use the result to prevent optimization
			}
		})
	}
}

// BenchmarkYAMLMarshal benchmarks YAML marshaling of projects.
func BenchmarkYAMLMarshal(b *testing.B) {
	tests := []struct {
		name        string
		numServices int
	}{
		{"Small_5_Services", 5},
		{"Medium_20_Services", 20},
		{"Large_50_Services", 50},
		{"XLarge_100_Services", 100},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			project := setupBenchmarkProject(tt.numServices)

			b.ResetTimer()
			for range b.N {
				data, err := yaml.Marshal(project)
				if err != nil {
					b.Fatalf("Failed to marshal project: %v", err)
				}
				_ = data // Use the result to prevent optimization
			}
		})
	}
}

// BenchmarkLoadProject benchmarks project loading.
func BenchmarkLoadProject(b *testing.B) {
	tempDir := b.TempDir()

	// Create a realistic docker-compose.yml file
	composeContent := `version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "80:80"
  api:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - db
  db:
    image: postgres:13
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
`

	composeFile := filepath.Join(tempDir, "docker-compose.yml")
	const filePermissions = 0600
	err := os.WriteFile(composeFile, []byte(composeContent), filePermissions)
	require.NoError(b, err)

	config := Compose.Config{
		DockerComposePath: []string{composeFile},
		WorkDir:           tempDir,
		OutputDir:         tempDir,
		Tag:               "latest",
		LogLevel:          "error", // Reduce logging overhead
	}

	b.ResetTimer()
	for range b.N {
		_, clientErr := Compose.NewComposeClient(context.Background(), config)
		if clientErr != nil {
			b.Fatalf("Failed to load project: %v", clientErr)
		}
	}
}

// BenchmarkProjectOperations benchmarks combined project operations.
func BenchmarkProjectOperations(b *testing.B) {
	tests := []struct {
		name        string
		numServices int
	}{
		{"Small_5_Services", 5},
		{"Medium_20_Services", 20},
		{"Large_50_Services", 50},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			benchmarkProjectOpsWithServices(b, tt.numServices)
		})
	}
}

// benchmarkProjectOpsWithServices is a helper function to reduce cognitive complexity.
func benchmarkProjectOpsWithServices(b *testing.B, numServices int) {
	tempDir := b.TempDir()
	deps := setupBenchmarkDependencies()
	deps.ProjectFromOptions = func(_ context.Context, _ *cli.ProjectOptions) (*types.Project, error) {
		return setupBenchmarkProject(numServices), nil
	}

	config := Compose.Config{
		DockerComposePath: []string{"docker-compose.yml"},
		WorkDir:           tempDir,
		OutputDir:         tempDir,
		Tag:               "v1.0.0",
		LogLevel:          "error",
	}

	b.ResetTimer()
	for range b.N {
		client, err := Compose.NewComposeClientWithDeps(context.Background(), config, deps)
		if err != nil {
			b.Fatalf("Failed to create client: %v", err)
		}

		// Benchmark the core operations (without Docker calls)
		_, err = client.SaveComposeFile(context.Background())
		if err != nil {
			b.Fatalf("Failed to save compose file: %v", err)
		}

		// Benchmark the tagging logic
		tagServices(client)
	}
}

// tagServices applies tags to services without images.
func tagServices(client *Compose.Client) {
	project := client.Project
	for _, s := range project.Services {
		if s.Image == "" {
			s.Image = s.Name + ":" + client.Config.Tag
			project.Services[s.Name] = s
		}
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns.
func BenchmarkMemoryAllocation(b *testing.B) {
	tests := []struct {
		name        string
		numServices int
	}{
		{"Small_5_Services", 5},
		{"Medium_20_Services", 20},
		{"Large_50_Services", 50},
		{"XLarge_100_Services", 100},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()

			b.ResetTimer()
			for range b.N {
				project := setupBenchmarkProject(tt.numServices)

				// Simulate typical operations
				images := make([]string, 0, len(project.Services))
				for _, svc := range project.Services {
					if svc.Image != "" {
						images = append(images, svc.Image)
					}
				}

				// Marshal to YAML
				data, err := yaml.Marshal(project)
				if err != nil {
					b.Fatalf("Failed to marshal: %v", err)
				}

				_ = images
				_ = data
			}
		})
	}
}

// BenchmarkConcurrentOperations benchmarks concurrent operations.
func BenchmarkConcurrentOperations(b *testing.B) {
	tempDir := b.TempDir()
	deps := setupBenchmarkDependencies()

	config := Compose.Config{
		DockerComposePath: []string{"docker-compose.yml"},
		WorkDir:           tempDir,
		OutputDir:         tempDir,
		Tag:               "latest",
		LogLevel:          "error",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client, err := Compose.NewComposeClientWithDeps(context.Background(), config, deps)
			if err != nil {
				b.Fatalf("Failed to create client: %v", err)
			}

			_, err = client.SaveComposeFile(context.Background())
			if err != nil {
				b.Fatalf("Failed to save compose file: %v", err)
			}
		}
	})
}

// BenchmarkDeliverProject benchmarks the full MCP tool function.
func BenchmarkDeliverProject(b *testing.B) {
	tempDir := b.TempDir()

	// Create a realistic compose file
	composeContent := `version: '3.8'
services:
  web:
    image: nginx:latest
  api:
    build: .
  db:
    image: postgres:13
`

	composeFile := filepath.Join(tempDir, "docker-compose.yml")
	const filePermissions = 0600
	err := os.WriteFile(composeFile, []byte(composeContent), filePermissions)
	require.NoError(b, err)

	config := Compose.Config{
		DockerComposePath: []string{composeFile},
		WorkDir:           tempDir,
		OutputDir:         tempDir,
		Tag:               "v1.0.0",
		LogLevel:          "error",
	}

	b.ResetTimer()
	for range b.N {
		client, clientErr := Compose.NewComposeClient(context.Background(), config)
		if clientErr != nil {
			b.Fatalf("Failed to create client: %v", clientErr)
		}

		// Benchmark only the non-Docker operations
		_, saveErr := client.SaveComposeFile(context.Background())
		if saveErr != nil {
			b.Fatalf("Failed in deliver project: %v", saveErr)
		}
	}
}

// Example expected benchmark results:
// BenchmarkNewComposeClient-8               1000      1000000 ns/op    50000 B/op     100 allocs/op
// BenchmarkSaveComposeFile/Small_5_Services-8    10000     100000 ns/op     5000 B/op      50 allocs/op
// BenchmarkSaveComposeFile/Medium_20_Services-8   5000     200000 ns/op    10000 B/op     100 allocs/op
// BenchmarkSaveComposeFile/Large_50_Services-8    2000     500000 ns/op    25000 B/op     250 allocs/op
// BenchmarkBuild_ServiceTagging/Small_5_Services-8   1000000   1000 ns/op      500 B/op      10 allocs/op
// BenchmarkImageCollection/Large_50_Services-8       500000   3000 ns/op     2000 B/op      20 allocs/op
// BenchmarkYAMLMarshal/Medium_20_Services-8          10000   100000 ns/op    20000 B/op     200 allocs/op
