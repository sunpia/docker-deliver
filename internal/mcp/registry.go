package mcp

import (
	"fmt"
	"sync"

	m "github.com/modelcontextprotocol/go-sdk/mcp"
)

type RegisterInterface interface {
	RegisterTool(name string, mServer *m.Server) error
}

type ServiceRegistry struct {
	mu       sync.RWMutex
	services map[string]RegisterInterface
}

var globalRegistry = NewServiceRegistry()

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]RegisterInterface),
	}
}

func (r *ServiceRegistry) RegisterService(name string, server RegisterInterface) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.services[name]; exists {
		return fmt.Errorf("service %s already registered", name)
	}

	r.services[name] = server
	return nil
}

// Global function
func RegisterService(name string, server RegisterInterface) error {
	return globalRegistry.RegisterService(name, server)
}
