package xmanager

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Test cases for Gateway
func TestGateway(t *testing.T) {
	// Create a logger for testing
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create a new Gateway instance
	gateway := NewGateway(logger)

	// Register mock resource types
	gateway.GetSouthernResourceManager().RegisterFactory("mock", func(uuid string, config map[string]any) (GenericResource, error) {
		return &MockGenericResource{}, nil
	})

	t.Run("LoadAndRetrieveResource", func(t *testing.T) {
		// Mock resource configuration
		resourceType := "mock"
		uuid := "resource1"
		config := map[string]any{"key": "value"}

		// Load resource into the Southern resource manager
		err := gateway.LoadSouthernResource(resourceType, uuid, config)
		assert.NoError(t, err, "Loading resource should not fail")

		// Retrieve the resource
		resource, err := gateway.GetSouthernResource(uuid)
		assert.NoError(t, err, "Retrieving resource should not fail")
		assert.NotNil(t, resource, "Retrieved resource should not be nil")
		assert.Equal(t, RESOURCE_PENDING, resource.Status())
	})

	t.Run("StartResource", func(t *testing.T) {
		ctx := context.Background()
		uuid := "resource1"

		// Start the resource
		err := gateway.GetSouthernResourceManager().StartResource(uuid, ctx)
		assert.NoError(t, err, "Starting resource should not fail")

		// Verify the resource status
		resource, err := gateway.GetSouthernResource(uuid)
		assert.NoError(t, err, "Retrieving resource should not fail")
		assert.Equal(t, RESOURCE_UP, resource.Status())
	})

	t.Run("StopResource", func(t *testing.T) {
		uuid := "resource1"

		// Stop the resource
		err := gateway.RemoveSouthernResource(uuid)
		assert.NoError(t, err, "Stopping resource should not fail")

		// Verify the resource is removed
		_, err = gateway.GetSouthernResource(uuid)
		assert.Error(t, err, "Resource should not exist after being stopped")
	})

	t.Run("ReloadAllManagers", func(t *testing.T) {
		// Reload all managers
		assert.NotPanics(t, func() {
			gateway.ReloadAllManagers()
		}, "Reloading all managers should not panic")
	})

	t.Run("StartMonitoring", func(t *testing.T) {
		// Create a resource that starts in RESOURCE_DOWN state
		gateway.GetSouthernResourceManager().RegisterFactory("mock_down", func(uuid string, config map[string]any) (GenericResource, error) {
			return &MockGenericResource{state: RESOURCE_DOWN}, nil
		})
		err := gateway.LoadSouthernResource("mock_down", "resource2", map[string]any{"key": "value"})
		assert.NoError(t, err, "Loading resource should not fail")
		err1 := gateway.GetSouthernResourceManager().StartResource("resource2", context.Background())
		assert.NoError(t, err1, nil)
		// Start monitoring
		go gateway.StartAllManagers()

		// Wait for the monitoring loop to attempt a reload
		time.Sleep(6 * time.Second)

		resource, err := gateway.GetSouthernResource("resource2")
		assert.NoError(t, err, "Retrieving resource should not fail")
		assert.Equal(t, RESOURCE_UP, resource.Status())
	})

	t.Run("PaginationResources", func(t *testing.T) {
		// Create multiple resources
		for i := 0; i < 10; i++ {
			err := gateway.LoadSouthernResource("mock", fmt.Sprintf("resource%d", i), map[string]any{"key": "value"})
			assert.NoError(t, err, "Loading resource should not fail")
		}

		// Paginate resources
		resources := gateway.GetSouthernResourceManager().PaginationResources(1, 5)
		assert.Len(t, resources, 5, "Pagination should return 5 resources")
	})
}
