// Copyright (C) 2025 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package rhilex

import (
	"context"
	"sync"
	"testing"
	"time"

	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Mock implementation of GenericResource
type MockGenericResource struct {
	mu     sync.RWMutex
	state  GenericResourceState
	config map[string]any
	uuid   string
}

func (r *MockGenericResource) Init(uuid string, configMap map[string]any) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.uuid = uuid
	r.config = configMap
	r.state = RESOURCE_PENDING
	return nil
}

func (r *MockGenericResource) Start(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state = RESOURCE_UP
	return nil
}

func (r *MockGenericResource) Status() GenericResourceState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state
}

func (r *MockGenericResource) Services() []ResourceService {
	return nil
}

func (r *MockGenericResource) OnService(request ResourceServiceRequest) (ResourceServiceResponse, error) {
	return ResourceServiceResponse{}, nil
}

func (r *MockGenericResource) Worker() *GenericResourceWorker {
	return &GenericResourceWorker{
		UUID:        r.uuid,
		Name:        "mock-resource",
		Type:        "mock",
		Worker:      r,
		Config:      r.config,
		Description: "Mock resource for testing",
	}
}
func (r *MockGenericResource) Topology() *LocalTopology {
	return &LocalTopology{}
}
func (r *MockGenericResource) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state = RESOURCE_STOP
}

// Test cases
func TestGenericResourceManager(t *testing.T) {
	manager := NewGenericResourceManager(nil)
	manager.SetLogger(logrus.New())

	// Register a mock factory
	manager.RegisterFactory("mock", func(uuid string, config map[string]any) (GenericResource, error) {
		return &MockGenericResource{}, nil
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
		manager.RegisterFactory("mock_down", func(uuid string, config map[string]any) (GenericResource, error) {
			return &MockGenericResource{state: RESOURCE_DOWN}, nil
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
		for _, v := range resources {
			t.Log(v.Worker())
		}
		assert.Len(t, resources, 5)
	})
	time.Sleep(10 * time.Second)
}
