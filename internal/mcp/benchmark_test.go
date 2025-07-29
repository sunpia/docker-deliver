package mcp

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

// BenchmarkServiceRegistry_RegisterService benchmarks service registration.
func BenchmarkServiceRegistry_RegisterService(b *testing.B) {
	registry := NewServiceRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service := &MockService{}
		service.On("RegisterTool", mock.Anything, mock.Anything).Return(nil)

		err := registry.RegisterService(fmt.Sprintf("service-%d", i), service)
		if err != nil {
			b.Fatalf("Failed to register service: %v", err)
		}
	}
}

// BenchmarkServiceRegistry_GetService benchmarks service retrieval.
func BenchmarkServiceRegistry_GetService(b *testing.B) {
	registry := NewServiceRegistry()

	// Pre-populate with services
	const numServices = 1000
	for i := 0; i < numServices; i++ {
		service := &MockService{}
		service.On("RegisterTool", mock.Anything, mock.Anything).Return(nil)
		err := registry.RegisterService(fmt.Sprintf("service-%d", i), service)
		if err != nil {
			b.Fatalf("Failed to register service: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		serviceName := fmt.Sprintf("service-%d", i%numServices)
		_, exists := registry.GetService(serviceName)
		if !exists {
			b.Fatalf("Service %s not found", serviceName)
		}
	}
}

// BenchmarkServiceRegistry_GetServices benchmarks getting all services.
func BenchmarkServiceRegistry_GetServices(b *testing.B) {
	registry := NewServiceRegistry()

	// Pre-populate with services
	const numServices = 100
	for i := 0; i < numServices; i++ {
		service := &MockService{}
		service.On("RegisterTool", mock.Anything, mock.Anything).Return(nil)
		err := registry.RegisterService(fmt.Sprintf("service-%d", i), service)
		if err != nil {
			b.Fatalf("Failed to register service: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		services := registry.GetServices()
		if len(services) != numServices {
			b.Fatalf("Expected %d services, got %d", numServices, len(services))
		}
	}
}

// BenchmarkServiceRegistry_ConcurrentReadWrite benchmarks concurrent read/write operations.
func BenchmarkServiceRegistry_ConcurrentReadWrite(b *testing.B) {
	registry := NewServiceRegistry()

	// Pre-populate with some services
	const initialServices = 100
	for i := 0; i < initialServices; i++ {
		service := &MockService{}
		service.On("RegisterTool", mock.Anything, mock.Anything).Return(nil)
		err := registry.RegisterService(fmt.Sprintf("initial-service-%d", i), service)
		if err != nil {
			b.Fatalf("Failed to register service: %v", err)
		}
	}

	b.ResetTimer()

	var wg sync.WaitGroup

	// Start readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < b.N/10; j++ {
				serviceName := fmt.Sprintf("initial-service-%d", j%initialServices)
				registry.GetService(serviceName)
				registry.Count()
				registry.ListServiceNames()
			}
		}(i)
	}

	// Start writers
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < b.N/50; j++ {
				service := &MockService{}
				service.On("RegisterTool", mock.Anything, mock.Anything).Return(nil)
				serviceName := fmt.Sprintf("writer-%d-service-%d", id, j)
				registry.RegisterService(serviceName, service)
			}
		}(i)
	}

	wg.Wait()
}

// BenchmarkClient_SetupServer benchmarks server setup with multiple services.
func BenchmarkClient_SetupServer(b *testing.B) {
	const numServices = 50

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		// Create a fresh registry for each iteration
		registry := NewServiceRegistry()

		// Register services
		for j := 0; j < numServices; j++ {
			service := &MockService{}
			service.On("RegisterTool", ":8080", mock.AnythingOfType("*mcp.Server")).Return(nil)
			err := registry.RegisterService(fmt.Sprintf("service-%d", j), service)
			if err != nil {
				b.Fatalf("Failed to register service: %v", err)
			}
		}

		config := Config{
			HttpAddr:      ":8080",
			ServerName:    "benchmark-server",
			ServerVersion: "v1.0.0",
		}

		client := &Client{
			config:   config,
			logger:   &logrus.Logger{},
			registry: registry,
		}

		b.StartTimer()

		// Benchmark the server setup
		err := client.setupServer()
		if err != nil {
			b.Fatalf("Failed to setup server: %v", err)
		}

		b.StopTimer()

		// Verify all mocks were called
		for j := 0; j < numServices; j++ {
			serviceName := fmt.Sprintf("service-%d", j)
			service, exists := registry.GetService(serviceName)
			if !exists {
				b.Fatalf("Service %s not found", serviceName)
			}
			service.(*MockService).AssertExpectations(b)
		}
	}
}

// BenchmarkNewClient benchmarks client creation.
func BenchmarkNewClient(b *testing.B) {
	config := Config{
		HttpAddr:      ":8080",
		ServerName:    "benchmark-server",
		ServerVersion: "v1.0.0",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client, err := NewClient(context.Background(), config)
		if err != nil {
			b.Fatalf("Failed to create client: %v", err)
		}
		_ = client // Use the client to prevent optimization
	}
}

// Example benchmark output expectations:
// BenchmarkServiceRegistry_RegisterService-8      1000000    1000 ns/op     200 B/op    3 allocs/op
// BenchmarkServiceRegistry_GetService-8           5000000     300 ns/op       0 B/op    0 allocs/op
// BenchmarkServiceRegistry_GetServices-8           100000   10000 ns/op    5000 B/op   50 allocs/op
// BenchmarkServiceRegistry_ConcurrentReadWrite-8   500000    2000 ns/op     500 B/op   10 allocs/op
// BenchmarkClient_SetupServer-8                     10000  100000 ns/op   10000 B/op  200 allocs/op
// BenchmarkNewClient-8                            1000000    1000 ns/op     500 B/op   10 allocs/op
