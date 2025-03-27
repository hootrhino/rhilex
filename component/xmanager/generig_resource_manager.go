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

import "github.com/hootrhino/rhilex/typex"

var __DefaultGenericResourceManager *GenericResourceManager

/*
*
* 管理器
*
 */
type GenericResourceManager struct {
	RuleEngine      typex.Rhilex
	ResourceManager *GatewayResourceManager
}

// 初始化多媒体运行时
func InitGenericRuntime(Rhilex typex.Rhilex) {
	__DefaultGenericResourceManager = &GenericResourceManager{
		RuleEngine:      Rhilex,
		ResourceManager: NewGatewayResourceManager(),
	}
}

func StopGenericRuntime() {
	if __DefaultGenericResourceManager == nil {
		return
	}
	for _, v := range __DefaultGenericResourceManager.ResourceManager.GetResourceList() {
		v.Stop()
	}
}
