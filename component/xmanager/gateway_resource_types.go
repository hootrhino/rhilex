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

// gateway_resource_types.go
package xmanager

import (
	"context"
	"fmt"
	"sync"
)

// GatewayResourceState 资源状态类型
type GatewayResourceState int

// to string
func (s GatewayResourceState) String() string {
	switch s {
	case RESOURCE_DOWN:
		return "DOWN"
	case RESOURCE_UP:
		return "UP"
	case RESOURCE_PAUSE:
		return "PAUSE"
	case RESOURCE_STOP:
		return "STOP"
	case RESOURCE_PENDING:
		return "PENDING"
	case RESOURCE_DISABLE:
		return "DISABLE"
	default:
		return "UNKNOWN"
	}
}

const (
	// 故障
	RESOURCE_DOWN GatewayResourceState = 0
	// 启用
	RESOURCE_UP GatewayResourceState = 1
	// 暂停
	RESOURCE_PAUSE GatewayResourceState = 2
	// 停止
	RESOURCE_STOP GatewayResourceState = 3
	// 准备
	RESOURCE_PENDING GatewayResourceState = 4
	// 禁用
	RESOURCE_DISABLE GatewayResourceState = 5
)

// 资源服务
type ResourceServiceRequest struct {
	Name   string // 服务名称
	Method string // 服务方法
	Args   []any  // 服务参数
}

// ResourceServiceReturn 资源服务返回
type ResourceServiceResponse struct {
	Type   string
	Result any
	Error  error
}

// to string
func (s *ResourceServiceResponse) String() string {
	return fmt.Sprintf("ResourceServiceResponse Type: %s, Result: %v, Error: %v", s.Type, s.Result, s.Error)
}

// 资源服务
type ResourceService struct {
	Name        string                  // 服务名称
	Description string                  // 服务描述
	Method      string                  // 服务方法
	Args        []any                   // 服务参数
	Response    ResourceServiceResponse // 服务返回
}

// to string
func (s *ResourceService) String() string {
	return fmt.Sprintf("ResourceService Name: %s, Description: %s, Method: %s, Args: %v, Response: %v",
		s.Name, s.Description, s.Method, s.Args, s.Response)
}

// GatewayResource 多媒体资源工作接口
type GatewayResource interface {
	Init(uuid string, configMap map[string]any) error
	Start(context.Context) error
	Status() GatewayResourceState
	Services() []ResourceService
	OnService(request ResourceServiceRequest) (ResourceServiceResponse, error)
	Details() *GatewayResourceWorker
	Stop()
}

// BaseGatewayResource 提供基础实现，确保状态的线程安全
type BaseGatewayResource struct {
	mu     sync.RWMutex
	state  GatewayResourceState
	config map[string]any
}

func (r *BaseGatewayResource) Init(uuid string, configMap map[string]any) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.config = configMap
	r.state = RESOURCE_PENDING
	return nil
}

func (r *BaseGatewayResource) Start(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.state != RESOURCE_PENDING {
		return fmt.Errorf("cannot start resource in state %s", r.state)
	}
	r.state = RESOURCE_UP
	return nil
}

func (r *BaseGatewayResource) Status() GatewayResourceState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state
}

func (r *BaseGatewayResource) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state = RESOURCE_STOP
}
