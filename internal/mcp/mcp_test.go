package mcp_test

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
	mcp_internal "github.com/sunpia/docker-deliver/internal/mcp"
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
		config   mcp_internal.Config
		wantErr  bool
		validate func(*testing.T, *mcp_internal.Client)
	}{
		{
			name: "valid config with default values",
			config: mcp_internal.Config{
				HTTPAddr: ":8080",
			},
			wantErr: false,
			validate: func(t *testing.T, client *mcp_internal.Client) {
				assert.Equal(t, ":8080", client.Config.HTTPAddr)
				assert.Equal(t, "docker-deliver", client.Config.ServerName)
				assert.Equal(t, "v1.0.0", client.Config.ServerVersion)
				assert.Equal(t, 30*time.Second, client.Config.ShutdownTimeout)
				assert.NotNil(t, client.Logger)
				assert.NotNil(t, client.Registry)
			},
		},
		{
			name: "valid config with custom values",
			config: mcp_internal.Config{
				HTTPAddr:        ":9090",
				ServerName:      "custom-server",
				ServerVersion:   "v2.0.0",
				ShutdownTimeout: 60 * time.Second,
				EnableStdioLogs: true,
			},
			wantErr: false,
			validate: func(t *testing.T, client *mcp_internal.Client) {
				assert.Equal(t, ":9090", client.Config.HTTPAddr)
				assert.Equal(t, "custom-server", client.Config.ServerName)
				assert.Equal(t, "v2.0.0", client.Config.ServerVersion)
				assert.Equal(t, 60*time.Second, client.Config.ShutdownTimeout)
				assert.True(t, client.Config.EnableStdioLogs)
			},
		},
		{
			name:    "empty config uses defaults",
			config:  mcp_internal.Config{},
			wantErr: false,
			validate: func(t *testing.T, client *mcp_internal.Client) {
				assert.Equal(t, "docker-deliver", client.Config.ServerName)
				assert.Equal(t, "v1.0.0", client.Config.ServerVersion)
				assert.Equal(t, 30*time.Second, client.Config.ShutdownTimeout)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := mcp_internal.NewClient(context.Background(), tt.config)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
				if tt.validate != nil {
					tt.validate(t, client)
				}
			}
		})
	}
}

func TestClient_Shutdown(t *testing.T) {
	t.Run("shutdown without HTTP server", func(t *testing.T) {
		client := &mcp_internal.Client{
			Config: mcp_internal.Config{ShutdownTimeout: 1 * time.Second},
			Logger: logrus.New(),
		}

		err := client.Shutdown(context.Background())
		assert.NoError(t, err)
	})

	t.Run("shutdown with HTTP server", func(t *testing.T) {
		// Create a test HTTP server
		const readHeaderTimeout = 10 * time.Second
		server := &http.Server{
			Addr:              ":0", // Use any available port
			ReadHeaderTimeout: readHeaderTimeout,
		}

		client := &mcp_internal.Client{
			Config:  mcp_internal.Config{ShutdownTimeout: 1 * time.Second},
			Logger:  logrus.New(),
			HTTPSrv: server,
		}

		err := client.Shutdown(context.Background())
		assert.NoError(t, err)
	})

	t.Run("shutdown timeout", func(t *testing.T) {
		// Create a server that won't shutdown gracefully
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(_ http.ResponseWriter, _ *http.Request) {
			// This handler will block to simulate a server that doesn't shutdown gracefully
			time.Sleep(2 * time.Second)
		})

		const readHeaderTimeout = 10 * time.Second
		server := &http.Server{
			Addr:              ":0",
			Handler:           mux,
			ReadHeaderTimeout: readHeaderTimeout,
		}

		client := &mcp_internal.Client{
			Config:  mcp_internal.Config{ShutdownTimeout: 10 * time.Millisecond}, // Very short timeout
			Logger:  logrus.New(),
			HTTPSrv: server,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := client.Shutdown(ctx)
		// Shutdown should complete without error even with timeout
		// The actual behavior depends on the HTTP server implementation
		assert.NoError(t, err)
	})
}
