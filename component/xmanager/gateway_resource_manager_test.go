package xmanager

import (
	"context"
	"sync"
	"testing"
	"time"

	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Mock implementation of GatewayResource
type MockGatewayResource struct {
	mu       sync.RWMutex
	state    GatewayResourceState
	config   map[string]any
	initErr  error
	startErr error
}

func (r *MockGatewayResource) Init(uuid string, configMap map[string]any) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.config = configMap
	if r.initErr != nil {
		return r.initErr
	}
	r.state = RESOURCE_PENDING
	return nil
}

func (r *MockGatewayResource) Start(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.startErr != nil {
		return r.startErr
	}
	r.state = RESOURCE_UP
	return nil
}

func (r *MockGatewayResource) Status() GatewayResourceState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state
}

func (r *MockGatewayResource) Services() []ResourceService {
	return nil
}

func (r *MockGatewayResource) OnService(request ResourceServiceRequest) (ResourceServiceResponse, error) {
	return ResourceServiceResponse{}, nil
}

func (r *MockGatewayResource) Details() *GatewayResourceWorker {
	return &GatewayResourceWorker{
		Config: r.config,
	}
}

func (r *MockGatewayResource) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state = RESOURCE_STOP
}

// Test cases
func TestGatewayResourceManager(t *testing.T) {
	manager := NewGatewayResourceManager()
	manager.SetLogger(logrus.New())

	// Register a mock factory
	manager.RegisterFactory("mock", func(uuid string, config map[string]any) (GatewayResource, error) {
		return &MockGatewayResource{}, nil
	})

	t.Run("CreateResource", func(t *testing.T) {
		err := manager.CreateResource("mock", "resource1", map[string]any{"key": "value"})
		assert.NoError(t, err)

		resource, err := manager.GetResource("resource1")
		assert.NoError(t, err)
		assert.NotNil(t, resource)
		assert.Equal(t, RESOURCE_PENDING, resource.Status())
	})

	t.Run("StartResource", func(t *testing.T) {
		ctx := context.Background()
		err := manager.StartResource("resource1", ctx)
		assert.NoError(t, err)

		resource, err := manager.GetResource("resource1")
		assert.NoError(t, err)
		assert.Equal(t, RESOURCE_UP, resource.Status())
	})

	t.Run("StopResource", func(t *testing.T) {
		err := manager.StopResource("resource1")
		assert.NoError(t, err)

		_, err = manager.GetResource("resource1")
		assert.Error(t, err)
	})

	t.Run("ReloadResource", func(t *testing.T) {
		// Create and start a resource
		err := manager.CreateResource("mock", "resource2", map[string]any{"key": "value"})
		assert.NoError(t, err)

		ctx := context.Background()
		err = manager.StartResource("resource2", ctx)
		assert.NoError(t, err)

		// Reload the resource
		err = manager.ReloadResource("resource2")
		assert.NoError(t, err)

		resource, err := manager.GetResource("resource2")
		assert.NoError(t, err)
		assert.Equal(t, RESOURCE_UP, resource.Status())
	})

	t.Run("StartMonitoring", func(t *testing.T) {
		// Create a resource that starts in RESOURCE_DOWN state
		manager.RegisterFactory("mock_down", func(uuid string, config map[string]any) (GatewayResource, error) {
			return &MockGatewayResource{state: RESOURCE_DOWN}, nil
		})
		err := manager.CreateResource("mock_down", "resource3", map[string]any{"key": "value"})
		assert.NoError(t, err)

		// Start monitoring
		go manager.StartMonitoring()

		// Wait for the monitoring loop to attempt a reload
		time.Sleep(6 * time.Second)

		resource, err := manager.GetResource("resource3")
		assert.NoError(t, err)
		assert.Equal(t, RESOURCE_UP, resource.Status())
	})

	t.Run("PaginationResources", func(t *testing.T) {
		// Create multiple resources
		for i := 0; i < 10; i++ {
			err := manager.CreateResource("mock", "resource"+fmt.Sprint(i), map[string]any{"key": "value"})
			assert.NoError(t, err)
		}

		// Paginate resources
		resources := manager.PaginationResources(1, 5)
		assert.Len(t, resources, 5)
	})
}
