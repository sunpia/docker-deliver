package mcp_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"

	mcp_internal "github.com/sunpia/docker-deliver/internal/mcp"
)

// BenchmarkServiceRegistry_RegisterService benchmarks service registration.
func BenchmarkServiceRegistry_RegisterService(b *testing.B) {
	registry := mcp_internal.NewServiceRegistry()

	b.ResetTimer()
	for i := range b.N {
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
	registry := mcp_internal.NewServiceRegistry()

	// Pre-populate with services
	const numServices = 1000
	for i := range numServices {
		service := &MockService{}
		service.On("RegisterTool", mock.Anything, mock.Anything).Return(nil)
		err := registry.RegisterService(fmt.Sprintf("service-%d", i), service)
		if err != nil {
			b.Fatalf("Failed to register service: %v", err)
		}
	}

	b.ResetTimer()
	for i := range b.N {
		serviceName := fmt.Sprintf("service-%d", i%numServices)
		_, exists := registry.GetService(serviceName)
		if !exists {
			b.Fatalf("Service %s not found", serviceName)
		}
	}
}

// BenchmarkServiceRegistry_GetServices benchmarks getting all services.
func BenchmarkServiceRegistry_GetServices(b *testing.B) {
	registry := mcp_internal.NewServiceRegistry()

	// Pre-populate with services
	const numServices = 100
	for i := range numServices {
		service := &MockService{}
		service.On("RegisterTool", mock.Anything, mock.Anything).Return(nil)
		err := registry.RegisterService(fmt.Sprintf("service-%d", i), service)
		if err != nil {
			b.Fatalf("Failed to register service: %v", err)
		}
	}

	b.ResetTimer()
	for range b.N {
		services := registry.GetServices()
		if len(services) != numServices {
			b.Fatalf("Expected %d services, got %d", numServices, len(services))
		}
	}
}

// BenchmarkServiceRegistry_ConcurrentReadWrite benchmarks concurrent read/write operations.
func BenchmarkServiceRegistry_ConcurrentReadWrite(b *testing.B) {
	registry := mcp_internal.NewServiceRegistry()

	// Pre-populate with some services
	const initialServices = 100
	for i := range initialServices {
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
	for i := range 10 {
		wg.Add(1)
		go func(_ int) {
			defer wg.Done()
			for j := range b.N / 10 {
				serviceName := fmt.Sprintf("initial-service-%d", j%initialServices)
				registry.GetService(serviceName)
				registry.Count()
				registry.ListServiceNames()
			}
		}(i)
	}

	// Start writers
	for i := range 2 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range b.N / 50 {
				service := &MockService{}
				service.On("RegisterTool", mock.Anything, mock.Anything).Return(nil)
				serviceName := fmt.Sprintf("writer-%d-service-%d", id, j)
				err := registry.RegisterService(serviceName, service)
				if err != nil {
					// Use b.Error instead of b.Fatalf to avoid calling Fatalf from a goroutine
					b.Errorf("Failed to register service: %v", err)
					return
				}
			}
		}(i)
	}

	wg.Wait()
}

// BenchmarkNewClient benchmarks client creation.
func BenchmarkNewClient(b *testing.B) {
	config := mcp_internal.Config{
		HTTPAddr:      ":8080",
		ServerName:    "benchmark-server",
		ServerVersion: "v1.0.0",
	}

	b.ResetTimer()
	for range b.N {
		client, err := mcp_internal.NewClient(context.Background(), config)
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
