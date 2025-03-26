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

package multimedia

import (
	"context"

	"github.com/hootrhino/rhilex/component/xmanager"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
)

// RTSPResourceConfig RTSP资源配置结构体
type RTSPResourceConfig struct {
	StreamUrl  string `validate:"required"`
	EnablePush bool
	PushUrl    string
	EnableAi   bool
	AiModel    string
}

// RTSPResource RTSP资源实现
type RTSPResource struct {
	state  xmanager.GatewayResourceState
	uuid   string
	config RTSPResourceConfig
	rhilex typex.Rhilex
}

// NewRTSPResource 创建新的RTSP资源
func NewRTSPResource(rhilex typex.Rhilex) (xmanager.GatewayResource, error) {
	return &RTSPResource{
		state:  xmanager.MEDIA_PENDING,
		config: RTSPResourceConfig{},
		rhilex: rhilex,
	}, nil
}

// Init 初始化RTSP资源
func (r *RTSPResource) Init(uuid string, configMap map[string]any) error {
	r.uuid = uuid
	err := xmanager.MapToConfig(configMap, &r.config)
	if err != nil {
		return err
	}
	r.state = xmanager.MEDIA_PENDING
	glogger.GLogger.Infof("RTSP resource %s initialized with config: %+v", uuid, r.config)
	return nil
}

// Start 启动RTSP资源
func (r *RTSPResource) Start(ctx context.Context) error {
	r.state = xmanager.MEDIA_UP
	glogger.GLogger.Infof("RTSP resource %s started, pulling stream from %s", r.uuid, r.config.StreamUrl)
	return nil
}

// Status 获取RTSP资源状态
func (r *RTSPResource) Status() xmanager.GatewayResourceState {
	glogger.GLogger.Infof("RTSP resource %s status: %s", r.uuid, r.state)
	return r.state
}

// Services 获取RTSP资源服务
func (r *RTSPResource) Services() []xmanager.ResourceService {
	services := []xmanager.ResourceService{
		{
			Name:   "rtsp",
			Method: "start",
			Args: []xmanager.ResourceServiceArg{
				{
					UUID: r.uuid,
					Args: []any{r.config.StreamUrl},
				},
			},
			Description: "启动RTSP资源",
		},
	}
	return services
}

// OnService 处理RTSP资源服务请求
func (r *RTSPResource) OnService(request xmanager.ResourceServiceRequest) (xmanager.ResourceServiceResponse, error) {
	glogger.GLogger.Debugf("RTSP resource %s received service request: %+v", r.uuid, request)
	if request.Name == "rtsp" && request.Method == "start" {
		if len(request.Args) == 1 {
			streamUrl, ok := request.Args[0].Args[0].(string)
			if ok {
				r.config.StreamUrl = streamUrl
				r.state = xmanager.MEDIA_UP
				glogger.GLogger.Warningf("RTSP resource %s started, pulling stream from %s", r.uuid, streamUrl)
			}
		}
	}
	return xmanager.ResourceServiceResponse{
		Type:   "string",
		Result: "ok",
		Error:  nil,
	}, nil
}

// Details 获取RTSP资源详情
func (r *RTSPResource) Details() *xmanager.GatewayResourceWorker {
	return nil
}

// Stop 停止RTSP资源
func (r *RTSPResource) Stop() {
	r.state = xmanager.MEDIA_DOWN
	glogger.GLogger.Infof("RTSP resource %s stopped", r.uuid)
}
