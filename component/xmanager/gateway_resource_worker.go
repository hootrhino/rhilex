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

// GenericResourceWorker 用于记录流媒体的元信息
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
