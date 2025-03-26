// Copyright (C) 2023 wwhai
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
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package cecolla

import (
	"fmt"

	"github.com/hootrhino/rhilex/cecolla/ithings"
	"github.com/hootrhino/rhilex/component/intercache"
	"github.com/hootrhino/rhilex/component/xmanager"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
)

var __DefaultCecollaResourceManager *CecollaResourceManager

/*
*
* 管理器
*
 */
type CecollaResourceManager struct {
	RuleEngine             typex.Rhilex
	CecollaResourceManager *xmanager.GatewayResourceManager
}

// 初始化多媒体运行时
func InitCecollaRuntime(Rhilex typex.Rhilex) {
	__DefaultCecollaResourceManager = &CecollaResourceManager{
		RuleEngine:             Rhilex,
		CecollaResourceManager: xmanager.NewGatewayResourceManager(Rhilex),
	}

	__DefaultCecollaResourceManager.CecollaResourceManager.RegisterType("ITHINGS_IOTHUB", ithings.NewIthingsResource)
	__DefaultCecollaResourceManager.CecollaResourceManager.StartMonitoring()

	intercache.RegisterSlot("__CecollaBinding")
	glogger.GLogger.Infof("CecollaResourceManager is initialized")
}

// 停止所有资源
func StopCecollaRuntime() {
	if __DefaultCecollaResourceManager == nil {
		return
	}
	for _, resource := range __DefaultCecollaResourceManager.CecollaResourceManager.GetResourceList() {
		if resource.Worker != nil {
			glogger.GLogger.Infof("Stop resource: %s", resource.UUID)
			resource.Worker.Stop()
		}
	}
	intercache.UnRegisterSlot("__CecollaBinding")
	glogger.GLogger.Infof("CecollaResourceManager is stopped")
}

// 加载多媒体资源
func LoadCecollaResource(uuid string, name string, resourceType string,
	configMap map[string]any, description string) error {
	if __DefaultCecollaResourceManager == nil {
		return fmt.Errorf("CecollaResourceManager is not initialized")
	}
	return __DefaultCecollaResourceManager.CecollaResourceManager.LoadResource(uuid, name,
		resourceType, configMap, description)
}

// 重启多媒体资源
func RestartCecollaResource(uuid string) error {
	if __DefaultCecollaResourceManager == nil {
		return fmt.Errorf("CecollaResourceManager is not initialized")
	}
	return __DefaultCecollaResourceManager.CecollaResourceManager.ReloadResource(uuid)
}

// 停止指定的多媒体资源
func StopCecollaResource(uuid string) error {
	if __DefaultCecollaResourceManager == nil {
		return fmt.Errorf("CecollaResourceManager is not initialized")
	}
	return __DefaultCecollaResourceManager.CecollaResourceManager.StopResource(uuid)
}

// 获取多媒体资源列表
func GetCecollaResourceList() []*xmanager.GatewayResourceWorker {
	if __DefaultCecollaResourceManager == nil {
		return nil
	}
	return __DefaultCecollaResourceManager.CecollaResourceManager.GetResourceList()
}

// 获取多媒体资源详情
func GetCecollaResourceDetails(uuid string) (*xmanager.GatewayResourceWorker, error) {
	if __DefaultCecollaResourceManager == nil {
		return nil, fmt.Errorf("CecollaResourceManager is not initialized")
	}
	return __DefaultCecollaResourceManager.CecollaResourceManager.GetResource(uuid)
}

// 获取多媒体资源状态
func GetCecollaResourceStatus(uuid string) (xmanager.GatewayResourceState, error) {
	if __DefaultCecollaResourceManager == nil {
		return xmanager.MEDIA_DOWN, fmt.Errorf("CecollaResourceManager is not initialized")
	}
	return __DefaultCecollaResourceManager.CecollaResourceManager.GetResourceStatus(uuid)
}

// 开始监控多媒体资源
func StartCecollaResourceMonitoring() {
	if __DefaultCecollaResourceManager == nil {
		return
	}
	__DefaultCecollaResourceManager.CecollaResourceManager.StartMonitoring()
}

// 注册多媒体资源类型
func RegisterCecollaResourceType(resourceType string,
	factory func(rhilex typex.Rhilex) (xmanager.GatewayResource, error)) {
	if __DefaultCecollaResourceManager == nil {
		return
	}
	__DefaultCecollaResourceManager.CecollaResourceManager.RegisterType(resourceType, factory)
}
