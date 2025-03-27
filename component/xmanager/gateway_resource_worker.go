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
	"fmt"
)

// GatewayResourceWorker 用于记录流媒体的元信息
type GatewayResourceWorker struct {
	Worker      GatewayResource // 实际的实现接口
	UUID        string          // 资源唯一标识
	Name        string          // 资源名称
	Type        string          // 资源类型
	Config      map[string]any  // 资源配置
	Description string          // 资源描述
}

// to string
func (g *GatewayResourceWorker) String() string {
	return fmt.Sprintf("UUID: %s, Name: %s, Type: %s, Description: %s", g.UUID, g.Name, g.Type, g.Description)
}

// GetConfig 获取配置
func (g *GatewayResourceWorker) GetConfig() map[string]any {
	return g.Config
}

// Check Config
func (g *GatewayResourceWorker) CheckConfig(config any) error {
	if g.Config == nil {
		return fmt.Errorf("config is nil")
	}
	if g.Config["uuid"] == nil {
		return fmt.Errorf("config uuid is nil")
	}
	if g.Config["name"] == nil {
		return fmt.Errorf("config name is nil")
	}
	if g.Config["type"] == nil {
		return fmt.Errorf("config type is nil")
	}
	if g.Config["description"] == nil {
		return fmt.Errorf("config description is nil")
	}
	if g.Config["config"] == nil {
		return fmt.Errorf("config config is nil")
	}

	if config == nil {
		return fmt.Errorf("config is nil")
	}
	err := MapToConfig(g.Config, config)
	if err != nil {
		return err
	}
	return nil
}
