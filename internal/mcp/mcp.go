package mcp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// ServerInterface defines the contract for an MCP server.
type ServerInterface interface {
	Run(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// Config holds the configuration for the MCP client.
type Config struct {
	HttpAddr        string        `json:"http_addr"`
	ServerName      string        `json:"server_name"`
	ServerVersion   string        `json:"server_version"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
	EnableStdioLogs bool          `json:"enable_stdio_logs"`
}

// Client represents an MCP server client that manages service registration and server lifecycle.
type Client struct {
	ServerInterface
	config   Config
	logger   *logrus.Logger
	server   *mcp.Server
	httpSrv  *http.Server
	registry *ServiceRegistry
}

// NewClient creates a new MCP client with the provided configuration.
func NewClient(ctx context.Context, config Config) (*Client, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	logger := logrus.New()
	if config.EnableStdioLogs {
		logger.SetOutput(os.Stderr)
	}

	// Use default values if not provided
	if config.ServerName == "" {
		config.ServerName = "docker-deliver"
	}
	if config.ServerVersion == "" {
		config.ServerVersion = "v1.0.0"
	}
	if config.ShutdownTimeout == 0 {
		config.ShutdownTimeout = 30 * time.Second
	}

	return &Client{
		config:   config,
		logger:   logger,
		registry: GetServiceRegistry(),
	}, nil
}

// validateConfig validates the client configuration.
func validateConfig(_ Config) error {
	// Add validation logic here if needed
	return nil
}

// Run starts the MCP server with the configured transport.
func (c *Client) Run(ctx context.Context) error {
	if err := c.setupServer(); err != nil {
		return fmt.Errorf("failed to setup server: %w", err)
	}

	if c.config.HttpAddr != "" {
		return c.runHTTPServer(ctx)
	}
	return c.runStdioServer(ctx)
}

// setupServer creates and configures the MCP server with registered services.
func (c *Client) setupServer() error {
	c.server = mcp.NewServer(&mcp.Implementation{
		Name:    c.config.ServerName,
		Version: c.config.ServerVersion,
	}, nil)

	// Register all services from the registry
	services := c.registry.GetServices()
	for name, service := range services {
		c.logger.Infof("Registering service: %s", name)
		if err := service.RegisterTool(c.config.HttpAddr, c.server); err != nil {
			return fmt.Errorf("failed to register service %s: %w", name, err)
		}
	}

	c.logger.Infof("Successfully registered %d services", len(services))
	return nil
}

// runHTTPServer starts the MCP server with HTTP transport.
func (c *Client) runHTTPServer(ctx context.Context) error {
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return c.server
	}, nil)

	c.httpSrv = &http.Server{
		Addr:    c.config.HttpAddr,
		Handler: handler,
	}

	c.logger.Infof("MCP handler listening at %s", c.config.HttpAddr)

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := c.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server failed: %w", err)
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		c.logger.Info("Shutting down HTTP server...")
		return c.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

// runStdioServer starts the MCP server with stdio transport.
func (c *Client) runStdioServer(ctx context.Context) error {
	c.logger.Info("Running MCP server with stdio transport")
	transport := mcp.NewStdioTransport()

	if c.config.EnableStdioLogs {
		loggingTransport := mcp.NewLoggingTransport(transport, os.Stderr)
		if err := c.server.Run(ctx, loggingTransport); err != nil {
			return fmt.Errorf("stdio server with logging failed: %w", err)
		}
	} else {
		if err := c.server.Run(ctx, transport); err != nil {
			return fmt.Errorf("stdio server failed: %w", err)
		}
	}
	return nil
}

// Shutdown gracefully shuts down the MCP server.
func (c *Client) Shutdown(ctx context.Context) error {
	if c.httpSrv != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, c.config.ShutdownTimeout)
		defer cancel()

		c.logger.Info("Gracefully shutting down HTTP server...")
		if err := c.httpSrv.Shutdown(shutdownCtx); err != nil {
			c.logger.Errorf("Error during HTTP server shutdown: %v", err)
			return fmt.Errorf("failed to shutdown HTTP server: %w", err)
		}
		c.logger.Info("HTTP server shutdown complete")
	}
	return nil
}
