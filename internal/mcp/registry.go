package mcp

import (
	"fmt"
	"sync"

	m "github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterInterface defines the contract for services that can register MCP tools.
type RegisterInterface interface {
	RegisterTool(name string, mServer *m.Server) error
}

// ServiceRegistry manages a collection of MCP services with thread-safe operations.
type ServiceRegistry struct {
	mu       sync.RWMutex
	services map[string]RegisterInterface
}

var (
	globalRegistry *ServiceRegistry
	once           sync.Once
)

// GetServiceRegistry returns the global service registry instance (singleton).
func GetServiceRegistry() *ServiceRegistry {
	once.Do(func() {
		globalRegistry = NewServiceRegistry()
	})
	return globalRegistry
}

// NewServiceRegistry creates a new service registry.
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]RegisterInterface),
	}
}

// RegisterService registers a service with the given name.
// Returns an error if a service with the same name is already registered.
func (r *ServiceRegistry) RegisterService(name string, service RegisterInterface) error {
	if name == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.services[name]; exists {
		return fmt.Errorf("service %s already registered", name)
	}

	r.services[name] = service
	return nil
}

// UnregisterService removes a service from the registry.
// Returns an error if the service is not found.
func (r *ServiceRegistry) UnregisterService(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.services[name]; !exists {
		return fmt.Errorf("service %s not found", name)
	}

	delete(r.services, name)
	return nil
}

// GetService retrieves a service by name.
// Returns the service and a boolean indicating if it was found.
func (r *ServiceRegistry) GetService(name string) (RegisterInterface, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[name]
	return service, exists
}

// GetServices returns a copy of all registered services.
// This prevents external modification of the internal map.
func (r *ServiceRegistry) GetServices() map[string]RegisterInterface {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Create a copy to prevent external modification
	services := make(map[string]RegisterInterface, len(r.services))
	for name, service := range r.services {
		services[name] = service
	}
	return services
}

// ListServiceNames returns a slice of all registered service names.
func (r *ServiceRegistry) ListServiceNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered services.
func (r *ServiceRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.services)
}

// Clear removes all registered services.
func (r *ServiceRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services = make(map[string]RegisterInterface)
}

// RegisterService is a convenience function that registers a service with the global registry.
func RegisterService(name string, service RegisterInterface) error {
	return GetServiceRegistry().RegisterService(name, service)
}
