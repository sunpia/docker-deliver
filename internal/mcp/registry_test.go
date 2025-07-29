package mcp

import (
	"fmt"
	"sync"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockService implements RegisterInterface for testing.
type MockService struct {
	mock.Mock
}

func (m *MockService) RegisterTool(name string, mServer *mcp.Server) error {
	args := m.Called(name, mServer)
	return args.Error(0)
}

func TestNewServiceRegistry(t *testing.T) {
	registry := NewServiceRegistry()

	assert.NotNil(t, registry)
	assert.NotNil(t, registry.services)
	assert.Equal(t, 0, len(registry.services))
}

func TestGetServiceRegistry(t *testing.T) {
	// Reset the singleton for testing
	once = sync.Once{}
	globalRegistry = nil

	registry1 := GetServiceRegistry()
	registry2 := GetServiceRegistry()

	// Should return the same instance (singleton)
	assert.Same(t, registry1, registry2)
	assert.NotNil(t, registry1)
}

func TestServiceRegistry_RegisterService(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		service     RegisterInterface
		wantErr     bool
		errorMsg    string
	}{
		{
			name:        "successful registration",
			serviceName: "test-service",
			service:     &MockService{},
			wantErr:     false,
		},
		{
			name:        "empty service name",
			serviceName: "",
			service:     &MockService{},
			wantErr:     true,
			errorMsg:    "service name cannot be empty",
		},
		{
			name:        "nil service",
			serviceName: "test-service",
			service:     nil,
			wantErr:     true,
			errorMsg:    "service cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewServiceRegistry()

			err := registry.RegisterService(tt.serviceName, tt.service)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				// Verify the service was actually registered
				service, exists := registry.GetService(tt.serviceName)
				assert.True(t, exists)
				assert.Same(t, tt.service, service)
			}
		})
	}
}

func TestServiceRegistry_RegisterService_DuplicateName(t *testing.T) {
	registry := NewServiceRegistry()
	service1 := &MockService{}
	service2 := &MockService{}

	// Register first service
	err := registry.RegisterService("duplicate", service1)
	require.NoError(t, err)

	// Try to register second service with same name
	err = registry.RegisterService("duplicate", service2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service duplicate already registered")

	// Verify original service is still there
	service, exists := registry.GetService("duplicate")
	assert.True(t, exists)
	assert.Same(t, service1, service)
}

func TestServiceRegistry_UnregisterService(t *testing.T) {
	registry := NewServiceRegistry()
	service := &MockService{}

	// Register a service
	err := registry.RegisterService("test-service", service)
	require.NoError(t, err)

	// Unregister the service
	err = registry.UnregisterService("test-service")
	assert.NoError(t, err)

	// Verify service is no longer there
	_, exists := registry.GetService("test-service")
	assert.False(t, exists)
}

func TestServiceRegistry_UnregisterService_NotFound(t *testing.T) {
	registry := NewServiceRegistry()

	err := registry.UnregisterService("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service non-existent not found")
}

func TestServiceRegistry_GetService(t *testing.T) {
	registry := NewServiceRegistry()
	service := &MockService{}

	// Test getting non-existent service
	result, exists := registry.GetService("non-existent")
	assert.False(t, exists)
	assert.Nil(t, result)

	// Register and test getting existing service
	err := registry.RegisterService("test-service", service)
	require.NoError(t, err)

	result, exists = registry.GetService("test-service")
	assert.True(t, exists)
	assert.Same(t, service, result)
}

func TestServiceRegistry_GetServices(t *testing.T) {
	registry := NewServiceRegistry()
	service1 := &MockService{}
	service2 := &MockService{}

	// Test empty registry
	services := registry.GetServices()
	assert.NotNil(t, services)
	assert.Equal(t, 0, len(services))

	// Register services
	err := registry.RegisterService("service1", service1)
	require.NoError(t, err)
	err = registry.RegisterService("service2", service2)
	require.NoError(t, err)

	// Get all services
	services = registry.GetServices()
	assert.Equal(t, 2, len(services))
	assert.Same(t, service1, services["service1"])
	assert.Same(t, service2, services["service2"])

	// Verify it returns a copy (external modification doesn't affect internal state)
	services["service3"] = &MockService{}
	internalServices := registry.GetServices()
	assert.Equal(t, 2, len(internalServices))
	_, exists := internalServices["service3"]
	assert.False(t, exists)
}

func TestServiceRegistry_ListServiceNames(t *testing.T) {
	registry := NewServiceRegistry()

	// Test empty registry
	names := registry.ListServiceNames()
	assert.NotNil(t, names)
	assert.Equal(t, 0, len(names))

	// Register services
	err := registry.RegisterService("service1", &MockService{})
	require.NoError(t, err)
	err = registry.RegisterService("service2", &MockService{})
	require.NoError(t, err)

	// Get names
	names = registry.ListServiceNames()
	assert.Equal(t, 2, len(names))
	assert.Contains(t, names, "service1")
	assert.Contains(t, names, "service2")
}

func TestServiceRegistry_Count(t *testing.T) {
	registry := NewServiceRegistry()

	// Test empty registry
	assert.Equal(t, 0, registry.Count())

	// Register services
	err := registry.RegisterService("service1", &MockService{})
	require.NoError(t, err)
	assert.Equal(t, 1, registry.Count())

	err = registry.RegisterService("service2", &MockService{})
	require.NoError(t, err)
	assert.Equal(t, 2, registry.Count())

	// Unregister a service
	err = registry.UnregisterService("service1")
	require.NoError(t, err)
	assert.Equal(t, 1, registry.Count())
}

func TestServiceRegistry_Clear(t *testing.T) {
	registry := NewServiceRegistry()

	// Register services
	err := registry.RegisterService("service1", &MockService{})
	require.NoError(t, err)
	err = registry.RegisterService("service2", &MockService{})
	require.NoError(t, err)

	assert.Equal(t, 2, registry.Count())

	// Clear registry
	registry.Clear()
	assert.Equal(t, 0, registry.Count())

	// Verify all services are gone
	_, exists := registry.GetService("service1")
	assert.False(t, exists)
	_, exists = registry.GetService("service2")
	assert.False(t, exists)
}

func TestGlobalRegisterService(t *testing.T) {
	// Reset the global registry for testing
	once = sync.Once{}
	globalRegistry = nil

	service := &MockService{}
	err := RegisterService("global-test", service)
	assert.NoError(t, err)

	// Verify it was registered in the global registry
	registry := GetServiceRegistry()
	retrievedService, exists := registry.GetService("global-test")
	assert.True(t, exists)
	assert.Same(t, service, retrievedService)
}

// TestServiceRegistry_ConcurrentAccess tests thread safety of the registry.
func TestServiceRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewServiceRegistry()
	const numGoroutines = 100
	const numOperations = 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperations)

	// Start multiple goroutines performing concurrent operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				serviceName := fmt.Sprintf("service-%d-%d", id, j)
				service := &MockService{}

				// Register service
				if err := registry.RegisterService(serviceName, service); err != nil {
					errors <- err
					return
				}

				// Get service
				if _, exists := registry.GetService(serviceName); !exists {
					errors <- fmt.Errorf("service %s not found after registration", serviceName)
					return
				}

				// List services (read operation)
				registry.ListServiceNames()
				registry.Count()
				registry.GetServices()

				// Unregister service
				if err := registry.UnregisterService(serviceName); err != nil {
					errors <- err
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("Concurrent operation failed: %v", err)
	}

	// Verify registry is empty after all operations
	assert.Equal(t, 0, registry.Count())
}

// TestServiceRegistry_RealScenario tests a more realistic scenario.
func TestServiceRegistry_RealScenario(t *testing.T) {
	registry := NewServiceRegistry()

	// Create mock services that simulate real behavior
	composeService := &MockService{}
	composeService.On("RegisterTool", mock.AnythingOfType("string"), mock.AnythingOfType("*mcp.Server")).
		Return(nil)

	dockerService := &MockService{}
	dockerService.On("RegisterTool", mock.AnythingOfType("string"), mock.AnythingOfType("*mcp.Server")).
		Return(nil)

	// Register services
	err := registry.RegisterService("compose", composeService)
	require.NoError(t, err)

	err = registry.RegisterService("docker", dockerService)
	require.NoError(t, err)

	// Simulate server setup process
	services := registry.GetServices()
	assert.Equal(t, 2, len(services))

	// Simulate registering tools with an MCP server
	mcpServer := &mcp.Server{} // This would normally be properly initialized
	for name, service := range services {
		err := service.RegisterTool("test-addr", mcpServer)
		assert.NoError(t, err)
		t.Logf("Registered service: %s", name)
	}

	// Verify all mocks were called as expected
	composeService.AssertExpectations(t)
	dockerService.AssertExpectations(t)
}
