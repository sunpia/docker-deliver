package mcp

import (
	"context"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

type Interface interface {
	Run(ctx context.Context) error
}

type Config struct {
	HttpAddr string
}

type Client struct {
	Interface // Interface embedding
	Config    Config
	Logger    *logrus.Logger
}

func NewClient(ctx context.Context, config Config) (*Client, error) {
	return &Client{
		Config: config,
		Logger: logrus.New(),
	}, nil
}

func (c *Client) Run(ctx context.Context) error {
	// Create a server with a single tool.
	server := mcp.NewServer(&mcp.Implementation{Name: "docker-deliver", Version: "v1.0.0"}, nil)
	for n, service := range globalRegistry.services {
		c.Logger.Infof("Registering service: %s", n)
		if err := service.RegisterTool(c.Config.HttpAddr, server); err != nil {
			return err
		}
	}

	if c.Config.HttpAddr != "" {
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
			return server
		}, nil)
		c.Logger.Infof("MCP handler listening at %s", c.Config.HttpAddr)
		http.ListenAndServe(c.Config.HttpAddr, handler)
	} else {
		c.Logger.Info("Running MCP server with stdio transport")
		t := mcp.NewLoggingTransport(mcp.NewStdioTransport(), os.Stderr)

		if err := server.Run(context.Background(), t); err != nil {
			c.Logger.Errorf("Server failed: %v", err)
		}
	}
	return nil
}
