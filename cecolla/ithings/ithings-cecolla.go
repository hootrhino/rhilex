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

package ithings

import (
	"context"

	"github.com/hootrhino/rhilex/component/xmanager"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
)

// IthingsResourceConfig Ithings资源配置结构体
type IthingsResourceConfig struct {
}

// IthingsResource Ithings资源实现
type IthingsResource struct {
	state  xmanager.GatewayResourceState
	uuid   string
	config IthingsResourceConfig
	rhilex typex.Rhilex
}

// NewIthingsResource 创建新的Ithings资源
func NewIthingsResource(rhilex typex.Rhilex) (xmanager.GatewayResource, error) {
	return &IthingsResource{
		state:  xmanager.MEDIA_PENDING,
		config: IthingsResourceConfig{},
		rhilex: rhilex,
	}, nil
}

// Init 初始化Ithings资源
func (r *IthingsResource) Init(uuid string, configMap map[string]any) error {
	r.uuid = uuid
	err := xmanager.MapToConfig(configMap, &r.config)
	if err != nil {
		return err
	}
	r.state = xmanager.MEDIA_PENDING
	glogger.GLogger.Infof("Ithings resource %s initialized with config: %+v", uuid, r.config)
	return nil
}

// Start 启动Ithings资源
func (r *IthingsResource) Start(ctx context.Context) error {
	r.state = xmanager.MEDIA_UP
	glogger.GLogger.Infof("Ithings resource %s started", r.uuid)
	return nil
}

// Status 获取Ithings资源状态
func (r *IthingsResource) Status() xmanager.GatewayResourceState {
	glogger.GLogger.Infof("Ithings resource %s status: %s", r.uuid, r.state)
	return r.state
}

// Services 获取Ithings资源服务
func (r *IthingsResource) Services() []xmanager.ResourceService {
	services := []xmanager.ResourceService{}
	return services
}

// OnService 处理Ithings资源服务请求
func (r *IthingsResource) OnService(request xmanager.ResourceServiceRequest) (xmanager.ResourceServiceResponse, error) {
	glogger.GLogger.Debugf("Ithings resource %s received service request: %+v", r.uuid, request)
	return xmanager.ResourceServiceResponse{
		Type:   "string",
		Result: "ok",
		Error:  nil,
	}, nil
}

// Details 获取Ithings资源详情
func (r *IthingsResource) Details() *xmanager.GatewayResourceWorker {
	return nil
}

// Stop 停止Ithings资源
func (r *IthingsResource) Stop() {
	r.state = xmanager.MEDIA_DOWN
	glogger.GLogger.Infof("Ithings resource %s stopped", r.uuid)
}
