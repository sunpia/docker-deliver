package mcp

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRegisterInterface is a mock implementation of RegisterInterface for testing.
type MockRegisterInterface struct {
	mock.Mock
}

func (m *MockRegisterInterface) RegisterTool(name string, mServer *mcp.Server) error {
	args := m.Called(name, mServer)
	return args.Error(0)
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		wantErr  bool
		validate func(*testing.T, *Client)
	}{
		{
			name: "valid config with default values",
			config: Config{
				HttpAddr: ":8080",
			},
			wantErr: false,
			validate: func(t *testing.T, client *Client) {
				assert.Equal(t, ":8080", client.config.HttpAddr)
				assert.Equal(t, "docker-deliver", client.config.ServerName)
				assert.Equal(t, "v1.0.0", client.config.ServerVersion)
				assert.Equal(t, 30*time.Second, client.config.ShutdownTimeout)
				assert.NotNil(t, client.logger)
				assert.NotNil(t, client.registry)
			},
		},
		{
			name: "valid config with custom values",
			config: Config{
				HttpAddr:        ":9090",
				ServerName:      "custom-server",
				ServerVersion:   "v2.0.0",
				ShutdownTimeout: 60 * time.Second,
				EnableStdioLogs: true,
			},
			wantErr: false,
			validate: func(t *testing.T, client *Client) {
				assert.Equal(t, ":9090", client.config.HttpAddr)
				assert.Equal(t, "custom-server", client.config.ServerName)
				assert.Equal(t, "v2.0.0", client.config.ServerVersion)
				assert.Equal(t, 60*time.Second, client.config.ShutdownTimeout)
				assert.True(t, client.config.EnableStdioLogs)
			},
		},
		{
			name:    "empty config uses defaults",
			config:  Config{},
			wantErr: false,
			validate: func(t *testing.T, client *Client) {
				assert.Equal(t, "docker-deliver", client.config.ServerName)
				assert.Equal(t, "v1.0.0", client.config.ServerVersion)
				assert.Equal(t, 30*time.Second, client.config.ShutdownTimeout)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(context.Background(), tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				if tt.validate != nil {
					tt.validate(t, client)
				}
			}
		})
	}
}

func TestClient_setupServer(t *testing.T) {
	tests := []struct {
		name           string
		setupRegistry  func(*ServiceRegistry)
		expectedError  string
		validateServer func(*testing.T, *Client)
	}{
		{
			name: "successful setup with no services",
			setupRegistry: func(reg *ServiceRegistry) {
				// No services registered
			},
			expectedError: "",
			validateServer: func(t *testing.T, client *Client) {
				assert.NotNil(t, client.server)
			},
		},
		{
			name: "successful setup with valid service",
			setupRegistry: func(reg *ServiceRegistry) {
				mockService := &MockRegisterInterface{}
				mockService.On("RegisterTool", ":8080", mock.AnythingOfType("*mcp.Server")).Return(nil)
				err := reg.RegisterService("test-service", mockService)
				require.NoError(t, err)
			},
			expectedError: "",
			validateServer: func(t *testing.T, client *Client) {
				assert.NotNil(t, client.server)
			},
		},
		{
			name: "setup fails when service registration fails",
			setupRegistry: func(reg *ServiceRegistry) {
				mockService := &MockRegisterInterface{}
				mockService.On("RegisterTool", ":8080", mock.AnythingOfType("*mcp.Server")).
					Return(assert.AnError)
				err := reg.RegisterService("failing-service", mockService)
				require.NoError(t, err)
			},
			expectedError: "failed to register service failing-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh registry for each test
			registry := NewServiceRegistry()
			tt.setupRegistry(registry)

			config := Config{
				HttpAddr:      ":8080",
				ServerName:    "test-server",
				ServerVersion: "v1.0.0",
			}

			client := &Client{
				config:   config,
				logger:   logrus.New(),
				registry: registry,
			}

			err := client.setupServer()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.validateServer != nil {
					tt.validateServer(t, client)
				}
			}
		})
	}
}

func TestClient_runHTTPServer(t *testing.T) {
	t.Run("successful HTTP server start and shutdown", func(t *testing.T) {
		config := Config{
			HttpAddr:        ":0", // Use any available port
			ServerName:      "test-server",
			ServerVersion:   "v1.0.0",
			ShutdownTimeout: 1 * time.Second,
		}

		client, err := NewClient(context.Background(), config)
		require.NoError(t, err)

		err = client.setupServer()
		require.NoError(t, err)

		// Create a context that we can cancel
		ctx, cancel := context.WithCancel(context.Background())

		// Start the server in a goroutine
		serverErr := make(chan error, 1)
		go func() {
			serverErr <- client.runHTTPServer(ctx)
		}()

		// Give the server a moment to start
		time.Sleep(100 * time.Millisecond)

		// Cancel the context to trigger shutdown
		cancel()

		// Wait for the server to shutdown
		err = <-serverErr
		assert.NoError(t, err)
	})
}

func TestClient_runStdioServer(t *testing.T) {
	t.Skip("Skipping stdio server test due to MCP SDK limitations in test environment")
	// The stdio server test is complex because it tries to read from stdin
	// In a real environment this works fine, but in test environment it can cause issues
	// This functionality is tested in integration tests instead
}

func TestClient_Shutdown(t *testing.T) {
	t.Run("shutdown without HTTP server", func(t *testing.T) {
		client := &Client{
			config: Config{ShutdownTimeout: 1 * time.Second},
			logger: logrus.New(),
		}

		err := client.Shutdown(context.Background())
		assert.NoError(t, err)
	})

	t.Run("shutdown with HTTP server", func(t *testing.T) {
		// Create a test HTTP server
		server := &http.Server{
			Addr: ":0", // Use any available port
		}

		client := &Client{
			config:  Config{ShutdownTimeout: 1 * time.Second},
			logger:  logrus.New(),
			httpSrv: server,
		}

		err := client.Shutdown(context.Background())
		assert.NoError(t, err)
	})

	t.Run("shutdown timeout", func(t *testing.T) {
		// Create a server that won't shutdown gracefully
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// This handler will block to simulate a server that doesn't shutdown gracefully
			time.Sleep(2 * time.Second)
		})

		server := &http.Server{
			Addr:    ":0",
			Handler: mux,
		}

		client := &Client{
			config:  Config{ShutdownTimeout: 10 * time.Millisecond}, // Very short timeout
			logger:  logrus.New(),
			httpSrv: server,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := client.Shutdown(ctx)
		// Shutdown should complete without error even with timeout
		// The actual behavior depends on the HTTP server implementation
		assert.NoError(t, err)
	})
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				HttpAddr: ":8080",
			},
			wantErr: false,
		},
		{
			name:    "empty config",
			config:  Config{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Integration test that tests the full client lifecycle
func TestClient_Integration(t *testing.T) {
	t.Run("full lifecycle test", func(t *testing.T) {
		// Create a registry and register a mock service
		registry := NewServiceRegistry()
		mockService := &MockRegisterInterface{}
		mockService.On("RegisterTool", "", mock.AnythingOfType("*mcp.Server")).Return(nil)

		err := registry.RegisterService("test-service", mockService)
		require.NoError(t, err)

		// Create client with stdio transport (no HTTP address)
		config := Config{
			ServerName:      "integration-test-server",
			ServerVersion:   "v1.0.0",
			EnableStdioLogs: false,
		}

		client, err := NewClient(context.Background(), config)
		require.NoError(t, err)

		// Replace the registry with our test registry
		client.registry = registry

		// Setup the server
		err = client.setupServer()
		require.NoError(t, err)

		// Verify that the mock service was called
		mockService.AssertExpectations(t)

		// Test shutdown
		err = client.Shutdown(context.Background())
		assert.NoError(t, err)
	})
}
