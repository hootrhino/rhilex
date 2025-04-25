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
package rhilex

import (
	"context"
	"fmt"
)

// GenericResourceState 资源状态类型
type GenericResourceState int

// To String
func (s GenericResourceState) String() string {
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
	RESOURCE_DOWN GenericResourceState = 0
	// 启用
	RESOURCE_UP GenericResourceState = 1
	// 暂停
	RESOURCE_PAUSE GenericResourceState = 2
	// 停止
	RESOURCE_STOP GenericResourceState = 3
	// 准备
	RESOURCE_PENDING GenericResourceState = 4
	// 禁用
	RESOURCE_DISABLE GenericResourceState = 5
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
	Name        string
	Description string
	Method      string
	Args        []any
	Response    ResourceServiceResponse
}

func (s *ResourceService) String() string {
	return fmt.Sprintf("ResourceService Name: %s, Description: %s, Method: %s, Args: %v, Response: %v",
		s.Name, s.Description, s.Method, s.Args, s.Response)
}

// GenericResource 多媒体资源工作接口
type GenericResource interface {
	Init(uuid string, configMap map[string]any) error
	Start(context.Context) error
	Status() GenericResourceState
	Services() []ResourceService
	Topology() *LocalTopology
	OnService(request ResourceServiceRequest) (ResourceServiceResponse, error)
	Worker() *GenericResourceWorker
	Stop()
}

// GenericResourceWorker
type GenericResourceWorker struct {
	Worker      GenericResource // 实际的实现接口
	UUID        string          // 资源唯一标识
	Name        string          // 资源名称
	Type        string          // 资源类型
	Config      map[string]any  // 资源配置
	Description string          // 资源描述
}

// to string
func (g *GenericResourceWorker) String() string {
	return fmt.Sprintf("UUID: %s, Name: %s, Type: %s, Description: %s", g.UUID, g.Name, g.Type, g.Description)
}

// GetConfig 获取配置
func (g *GenericResourceWorker) GetConfig() map[string]any {
	return g.Config
}
