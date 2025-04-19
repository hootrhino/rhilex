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

package xmanager

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock implementation of GenericResource for testing
type MockResource struct {
	uuid     string
	config   map[string]any
	state    GenericResourceState
	services []ResourceService
}

func (m *MockResource) Init(uuid string, configMap map[string]any) error {
	m.uuid = uuid
	m.config = configMap
	m.state = RESOURCE_PENDING
	return nil
}

func (m *MockResource) Start(ctx context.Context) error {
	m.state = RESOURCE_UP
	return nil
}

func (m *MockResource) Status() GenericResourceState {
	return m.state
}

func (m *MockResource) Services() []ResourceService {
	return m.services
}
func (m *MockResource) OnService(request ResourceServiceRequest) (ResourceServiceResponse, error) {
	// Handle other services
	for _, service := range m.services {
		if service.Name == request.Name {
			return service.Response, nil
		}
	}
	return ResourceServiceResponse{}, fmt.Errorf("service not found")
}
func (m *MockResource) Topology() *LocalTopology {
	// Mock implementation of topology
	topology := NewLocalTopology("local_topology", "Local Topology")
	topology.AddDevice(Device{
		ID:              "device1",
		Type:            "device1",
		Protocol:        "MODBUSTCP",
		IP:              "127.0.0.1",
		Port:            502,
		Name:            "device1",
		Status:          "device1",
		Location:        "device1",
		Model:           "device1",
		Manufacturer:    "device1",
		SerialNumber:    "device1",
		FirmwareVersion: "device1",
		SoftwareVersion: "device1",
		Properties:      map[string]any{"key": "value"},
		DataPoints: []DataPoint{
			{
				Name:        "data_point1",
				Description: "Data Point 1",
				Type:        "int",
				Values:      []any{1, 2, 3},
				Unit:        "unit1",
				Properties:  map[string]any{"key": "value"},
			},
		},
		LastSeen:    "device1",
		LastUpdated: "device1",
	})
	return topology
}
func (m *MockResource) Worker() *GenericResourceWorker {
	return &GenericResourceWorker{
		Worker: m,
		UUID:   m.uuid,
		Config: m.config,
	}
}

func (m *MockResource) Stop() {
	m.state = RESOURCE_STOP
}

func TestTopology(t *testing.T) {
	// Create mock resources
	resource1 := &MockResource{
		services: []ResourceService{
			{
				Name:        "Service1",
				Description: "Test Service 1",
				Method:      "Method1",
				Response: ResourceServiceResponse{
					Type:   "Success",
					Result: "Result1",
				},
			},
		},
	}

	resource2 := &MockResource{
		services: []ResourceService{
			{
				Name:        "Service2",
				Description: "Test Service 2",
				Method:      "Method2",
				Response: ResourceServiceResponse{
					Type:   "Success",
					Result: "Result2",
				},
			},
		},
	}

	// Initialize resources
	err := resource1.Init("resource1", map[string]any{"key": "value1"})
	assert.NoError(t, err)
	err = resource2.Init("resource2", map[string]any{"key": "value2"})
	assert.NoError(t, err)

	// Start resources
	err = resource1.Start(context.Background())
	assert.NoError(t, err)
	err = resource2.Start(context.Background())
	assert.NoError(t, err)

	// Check resource statuses
	assert.Equal(t, RESOURCE_UP, resource1.Status())
	assert.Equal(t, RESOURCE_UP, resource2.Status())

	// Invoke services
	response, err := resource1.OnService(ResourceServiceRequest{Name: "Service1"})
	assert.NoError(t, err)
	assert.Equal(t, "Result1", response.Result)

	response, err = resource2.OnService(ResourceServiceRequest{Name: "Service2"})
	assert.NoError(t, err)
	assert.Equal(t, "Result2", response.Result)
	//
	// Check topology
	topology := resource1.Topology()
	t.Log(topology)
	// Stop resources
	resource1.Stop()
	resource2.Stop()
	assert.Equal(t, RESOURCE_STOP, resource1.Status())
	assert.Equal(t, RESOURCE_STOP, resource2.Status())
}
