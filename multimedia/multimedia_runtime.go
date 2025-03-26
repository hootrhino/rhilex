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
package multimedia

import (
	"fmt"

	"github.com/hootrhino/rhilex/component/intercache"
	"github.com/hootrhino/rhilex/component/xmanager"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
)

var __DefaultMultimediaResourceManager *MultimediaResourceManager

/*
*
* 管理器
*
 */
type MultimediaResourceManager struct {
	RuleEngine                typex.Rhilex
	MultimediaResourceManager *xmanager.GatewayResourceManager
}

// 初始化多媒体运行时
func InitMultimediaRuntime(Rhilex typex.Rhilex) {
	__DefaultMultimediaResourceManager = &MultimediaResourceManager{
		RuleEngine:                Rhilex,
		MultimediaResourceManager: xmanager.NewGatewayResourceManager(Rhilex),
	}

	__DefaultMultimediaResourceManager.MultimediaResourceManager.RegisterType("RTSP_STREAM", NewRTSPResource)
	__DefaultMultimediaResourceManager.MultimediaResourceManager.StartMonitoring()

	intercache.RegisterSlot("__MultimediaBinding")
	glogger.GLogger.Infof("MultimediaResourceManager is initialized")
}

// 停止所有资源
func StopMultimediaRuntime() {
	if __DefaultMultimediaResourceManager == nil {
		return
	}
	for _, resource := range __DefaultMultimediaResourceManager.MultimediaResourceManager.GetResourceList() {
		if resource.Worker != nil {
			glogger.GLogger.Infof("Stop resource: %s", resource.UUID)
			resource.Worker.Stop()
		}
	}
	intercache.UnRegisterSlot("__MultimediaBinding")
	glogger.GLogger.Infof("MultimediaResourceManager is stopped")
}

// 加载多媒体资源
func LoadMultimediaResource(uuid string, name string, resourceType string,
	configMap map[string]any, description string) error {
	if __DefaultMultimediaResourceManager == nil {
		return fmt.Errorf("MultimediaResourceManager is not initialized")
	}
	return __DefaultMultimediaResourceManager.MultimediaResourceManager.LoadResource(uuid, name,
		resourceType, configMap, description)
}

// 重启多媒体资源
func RestartMultimediaResource(uuid string) error {
	if __DefaultMultimediaResourceManager == nil {
		return fmt.Errorf("MultimediaResourceManager is not initialized")
	}
	return __DefaultMultimediaResourceManager.MultimediaResourceManager.ReloadResource(uuid)
}

// 停止指定的多媒体资源
func StopMultimediaResource(uuid string) error {
	if __DefaultMultimediaResourceManager == nil {
		return fmt.Errorf("MultimediaResourceManager is not initialized")
	}
	return __DefaultMultimediaResourceManager.MultimediaResourceManager.StopResource(uuid)
}

// 获取多媒体资源列表
func GetMultimediaResourceList() []*xmanager.GatewayResourceWorker {
	if __DefaultMultimediaResourceManager == nil {
		return nil
	}
	return __DefaultMultimediaResourceManager.MultimediaResourceManager.GetResourceList()
}

// 获取多媒体资源详情
func GetMultimediaResourceDetails(uuid string) (*xmanager.GatewayResourceWorker, error) {
	if __DefaultMultimediaResourceManager == nil {
		return nil, fmt.Errorf("MultimediaResourceManager is not initialized")
	}
	return __DefaultMultimediaResourceManager.MultimediaResourceManager.GetResource(uuid)
}

// 获取多媒体资源状态
func GetMultimediaResourceStatus(uuid string) (xmanager.GatewayResourceState, error) {
	if __DefaultMultimediaResourceManager == nil {
		return xmanager.MEDIA_DOWN, fmt.Errorf("MultimediaResourceManager is not initialized")
	}
	return __DefaultMultimediaResourceManager.MultimediaResourceManager.GetResourceStatus(uuid)
}

// 开始监控多媒体资源
func StartMultimediaResourceMonitoring() {
	if __DefaultMultimediaResourceManager == nil {
		return
	}
	__DefaultMultimediaResourceManager.MultimediaResourceManager.StartMonitoring()
}

// 注册多媒体资源类型
func RegisterMultimediaResourceType(resourceType string,
	factory func(rhilex typex.Rhilex) (xmanager.GatewayResource, error)) {
	if __DefaultMultimediaResourceManager == nil {
		return
	}
	__DefaultMultimediaResourceManager.MultimediaResourceManager.RegisterType(resourceType, factory)
}
