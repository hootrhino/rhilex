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

	"github.com/hootrhino/rhilex/component/orderedmap"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
)

// GatewayResourceManager 通用资源管理器
type GatewayResourceManager struct {
	resources *orderedmap.OrderedMap[string, *GatewayResourceWorker]
	types     map[string]func(rhilex typex.Rhilex) (GatewayResource, error)
	mu        sync.RWMutex
	rhilex    typex.Rhilex
}

// NewGatewayResourceManager 创建新的资源管理器
func NewGatewayResourceManager(rhilex typex.Rhilex) *GatewayResourceManager {
	return &GatewayResourceManager{
		resources: orderedmap.NewOrderedMap[string, *GatewayResourceWorker](),
		types:     make(map[string]func(rhilex typex.Rhilex) (GatewayResource, error)),
		rhilex:    rhilex,
	}
}

// RegisterType 注册资源类型和其对应的 worker 实现
func (m *GatewayResourceManager) RegisterType(resourceType string,
	factory func(rhilex typex.Rhilex) (GatewayResource, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.types[resourceType] = factory
}

// LoadResource 加载资源
func (m *GatewayResourceManager) LoadResource(uuid string, name string, resourceType string,
	configMap map[string]any, description string) error {
	m.mu.RLock()
	factory, exists := m.types[resourceType]
	m.mu.RUnlock()
	if !exists {
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}
	for _, resource := range m.resources.Values() {
		if resource.Name == name {
			return fmt.Errorf("resource name already exists: %s", name)
		}
		if resource.UUID == uuid {
			return fmt.Errorf("resource uuid already exists: %s", uuid)
		}
	}
	worker, err := factory(m.rhilex)
	if err != nil {
		return err
	}
	grw := &GatewayResourceWorker{
		Worker:      worker,
		UUID:        uuid,
		Name:        name,
		Type:        resourceType,
		Config:      configMap,
		Description: description,
	}
	m.resources.Set(uuid, grw)

	err = worker.Init(uuid, configMap)
	if err != nil {
		return err
	}
	err = worker.Start(context.Background())
	if err != nil {
		return err
	}

	return nil
}

// ReloadResource 重启资源
func (m *GatewayResourceManager) ReloadResource(uuid string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	worker, exists := m.resources.Get(uuid)
	if !exists {
		return fmt.Errorf("resource not found: %s", uuid)
	}
	worker.Worker.Stop()
	return m.LoadResource(uuid, worker.Name, worker.Type, worker.Config, worker.Description)
}

// StopResource 停止资源
func (m *GatewayResourceManager) StopResource(uuid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	worker, exists := m.resources.Get(uuid)
	if !exists {
		return fmt.Errorf("resource not found: %s", uuid)
	}
	worker.Worker.Stop()
	m.resources.Delete(uuid)
	return nil
}

// GetResourceList 获取资源列表
func (m *GatewayResourceManager) GetResourceList() []*GatewayResourceWorker {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.resources.Values()
}

// PaginationResources 分页获取
func (m *GatewayResourceManager) PaginationResources(current, size int) []*GatewayResourceWorker {
	m.mu.RLock()
	defer m.mu.RUnlock()
	workers := m.resources.Values()
	start := (current - 1) * size
	end := start + size
	if start > len(workers) {
		start = len(workers)
	}
	if end > len(workers) {
		end = len(workers)
	}
	return workers[start:end]
}

// GetResourceDetails 获取资源详情
func (m *GatewayResourceManager) GetResource(uuid string) (*GatewayResourceWorker, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	worker, exists := m.resources.Get(uuid)
	if !exists {
		return nil, fmt.Errorf("resource not found: %s", uuid)
	}
	return worker, nil
}

// GetResourceStatus 获取资源状态
func (m *GatewayResourceManager) GetResourceStatus(uuid string) (GatewayResourceState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	worker, exists := m.resources.Get(uuid)
	if !exists {
		return MEDIA_DOWN, fmt.Errorf("resource not found: %s", uuid)
	}
	return worker.Worker.Status(), nil
}

// StartMonitoring 开始资源监控
func (m *GatewayResourceManager) StartMonitoring() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			select {
			case <-context.Background().Done():
				return
			default:
			}
			for _, uuid := range m.resources.Keys() {
				worker, exists := m.resources.Get(uuid)
				if exists {
					status := worker.Worker.Status()
					switch status {
					case MEDIA_DOWN:
						glogger.GLogger.Warnf("Resource %s is down, reloading:", worker.UUID)
						m.ReloadResource(worker.UUID)
					case MEDIA_STOP, MEDIA_DISABLE:
						glogger.GLogger.Warnf("Resource %s is stopped, stopping:", worker.UUID)
						m.StopResource(worker.UUID)
					case MEDIA_PENDING:
						glogger.GLogger.Debugf("Resource %s is pending, starting:", worker.UUID)
						continue
					}
				}

			}

		}
	}()
}
func (m *GatewayResourceManager) StopMonitoring() {

}
