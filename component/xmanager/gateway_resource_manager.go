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

// gateway_resource_worker.go
package xmanager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// GenericResourceFactory 资源构造函数类型
type GenericResourceFactory func(uuid string, config map[string]any) (GenericResource, error)

// GenericResourceManager 管理所有资源
type GenericResourceManager struct {
	mu        sync.RWMutex
	resources map[string]GenericResource
	factories map[string]GenericResourceFactory
	logger    *logrus.Logger
}

// NewGenericResourceManager 创建新的资源管理器
func NewGenericResourceManager() *GenericResourceManager {
	return &GenericResourceManager{
		resources: make(map[string]GenericResource),
		factories: make(map[string]GenericResourceFactory),
	}
}
func (m *GenericResourceManager) SetLogger(logger *logrus.Logger) {
	m.logger = logger
}

// RegisterFactory 注册资源类型及其构造函数
func (m *GenericResourceManager) RegisterFactory(resourceType string, factory GenericResourceFactory) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.factories[resourceType] = factory
}

// ReloadResource 重新加载资源
func (m *GenericResourceManager) ReloadResource(uuid string) error {
	m.mu.RLock()
	resource, exists := m.resources[uuid]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("resource %s not found", uuid)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check the current state of the resource
	currentState := resource.Status()
	if currentState == RESOURCE_UP {
		m.logger.Infof("Resource %s is already running, skipping reload", uuid)
		return nil
	}

	// Reinitialize the resource
	if err := resource.Init(uuid, resource.Details().GetConfig()); err != nil {
		m.logger.Errorf("Failed to reinitialize resource %s: %v", uuid, err)
		return err
	}

	// Restart the resource
	ctx := context.Background()
	if err := resource.Start(ctx); err != nil {
		m.logger.Errorf("Failed to restart resource %s: %v", uuid, err)
		return err
	}

	m.logger.Infof("Resource %s successfully reloaded", uuid)
	return nil
}

// CreateResource 创建资源
func (m *GenericResourceManager) CreateResource(resourceType, uuid string, config map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if the factory for the resource type exists
	factory, exists := m.factories[resourceType]
	if !exists {
		return fmt.Errorf("resource type '%s' is not registered", resourceType)
	}

	// Validate the configuration
	if config == nil {
		return fmt.Errorf("configuration for resource '%s' cannot be nil", uuid)
	}

	// Create the resource using the factory
	resource, err := factory(uuid, config)
	if err != nil {
		return fmt.Errorf("failed to create resource '%s' of type '%s': %w", uuid, resourceType, err)
	}

	// Add the resource to the manager before initialization
	m.resources[uuid] = resource

	// Initialize the resource
	if err := resource.Init(uuid, config); err != nil {
		return fmt.Errorf("failed to initialize resource '%s': %w", uuid, err)
	}

	return nil
}

// StartResource 启动资源
func (m *GenericResourceManager) StartResource(uuid string, ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	resource, exists := m.resources[uuid]
	if !exists {
		return fmt.Errorf("resource %s not found", uuid)
	}

	return resource.Start(ctx)
}

// StopResource 停止资源
func (m *GenericResourceManager) StopResource(uuid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	resource, exists := m.resources[uuid]
	if !exists {
		return fmt.Errorf("resource %s not found", uuid)
	}

	resource.Stop()
	delete(m.resources, uuid)
	return nil
}

// GetResource 获取资源
func (m *GenericResourceManager) GetResource(uuid string) (GenericResource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	resource, exists := m.resources[uuid]
	if !exists {
		return nil, fmt.Errorf("resource %s not found", uuid)
	}

	return resource, nil
}

// GetResourceList 获取资源列表
func (m *GenericResourceManager) GetResourceList() []GenericResource {
	m.mu.RLock()
	defer m.mu.RUnlock()
	resources := make([]GenericResource, 0, len(m.resources))
	for _, resource := range m.resources {
		resources = append(resources, resource)
	}
	return resources
}

// PaginationResources 分页获取
func (m *GenericResourceManager) PaginationResources(current, size int) []GenericResource {
	m.mu.RLock()
	defer m.mu.RUnlock()
	resources := m.GetResourceList()
	start := (current - 1) * size
	end := start + size
	if start > len(resources) {
		start = len(resources)
	}
	if end > len(resources) {
		end = len(resources)
	}
	return resources[start:end]
}

// GetResourceStatus 获取资源状态
func (m *GenericResourceManager) GetResourceStatus(uuid string) (GenericResourceState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	resource, exists := m.resources[uuid]
	if !exists {
		return RESOURCE_DOWN, fmt.Errorf("resource not found: %s", uuid)
	}
	return resource.Status(), nil
}

// StartMonitoring 开始资源监控
func (m *GenericResourceManager) StartMonitoring() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		// Map to track restart attempts for each resource
		restartAttempts := make(map[string]int)
		const maxRetries = 3
		const backoffDuration = 2 * time.Second

		for range ticker.C {
			select {
			case <-context.Background().Done():
				return
			default:
				m.monitorResourcesWithRestartPolicy(restartAttempts, maxRetries, backoffDuration)
			}
		}
	}()
}

// monitorResourcesWithRestartPolicy 监控所有资源并应用重启策略
func (m *GenericResourceManager) monitorResourcesWithRestartPolicy(restartAttempts map[string]int, maxRetries int, backoffDuration time.Duration) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for uuid, resource := range m.resources {
		status := resource.Status()

		// Handle resource status with restart policy
		switch status {
		case RESOURCE_DOWN:
			m.logger.Warnf("Resource %s is down, attempting to reload", uuid)

			// Check if the resource has exceeded the maximum retries
			if restartAttempts[uuid] >= maxRetries {
				m.logger.Errorf("Resource %s has exceeded maximum restart attempts (%d), skipping further retries", uuid, maxRetries)
				continue
			}

			// Attempt to reload the resource
			if err := m.ReloadResource(uuid); err != nil {
				m.logger.Errorf("Failed to reload resource %s: %v", uuid, err)
				restartAttempts[uuid]++
				time.Sleep(backoffDuration) // Apply backoff before the next retry
			} else {
				m.logger.Infof("Resource %s successfully reloaded", uuid)
				restartAttempts[uuid] = 0 // Reset retry count on success
			}

		case RESOURCE_STOP, RESOURCE_DISABLE:
			m.logger.Warnf("Resource %s is stopped or disabled, skipping reload", uuid)

		case RESOURCE_PENDING:
			m.logger.Debugf("Resource %s is pending, waiting for initialization", uuid)

		default:
			m.logger.Debugf("Resource %s is in state %v, no action required", uuid, status)
		}
	}
}

// StopMonitoring 停止资源监控
func (m *GenericResourceManager) StopMonitoring() {

	m.logger.Infof("Monitoring has been stopped")
}
