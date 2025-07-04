// Copyright (C) 2024 wwhai
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

package typex

import "context"

// InEndType
type InEndType string

func (i InEndType) String() string {
	return string(i)
}

const (
	CUSTOM_PROTOCOL_SERVER InEndType = "CUSTOM_PROTOCOL_SERVER" // CUSTOM_PROTOCOL_SERVER
	COAP_SERVER            InEndType = "COAP_SERVER"            // COAP_SERVER
	UDP_SERVER             InEndType = "UDP_SERVER"             // UDP_SERVER
	TCP_SERVER             InEndType = "TCP_SERVER"             // TCP_SERVER
	HTTP_SERVER            InEndType = "HTTP_SERVER"            // HTTP_SERVER
	GRPC_SERVER            InEndType = "GRPC_SERVER"            // GRPC_SERVER
	GENERIC_MQTT_SERVER    InEndType = "GENERIC_MQTT_SERVER"    // Mqtt Server
	INTERNAL_EVENT         InEndType = "INTERNAL_EVENT"         // 内部消息
	COMTC_EVENT_FORWARDER  InEndType = "COMTC_EVENT_FORWARDER"  // 外设通信模块事件
)

// XStatus for source status
type XStatus struct {
	PointId     string             // Input: Source; Output: Target
	Ctx         context.Context    // context
	CancelCTX   context.CancelFunc // cancel
	RuleEngine  Rhilex             // rhilex
	Busy        bool               // 是否处于忙碌状态, 防止请求拥挤
	SourceState SourceState
}

// XSource 接口代表了一个终端资源，例如实际的MQTT客户端。
// 它定义了与资源交互所需的一系列方法，包括测试资源可用性、初始化、启动、数据传输等。
type XSource interface {
	// Init方法用于初始化资源，传递资源配置信息。
	// inEndId是资源的标识符，configMap是资源配置的映射。
	// 返回初始化是否成功的错误信息。
	Init(inEndId string, configMap map[string]any) error
	// Start方法用于启动资源。
	// CCTX是上下文，具体作用取决于资源的实现。
	// 返回启动是否成功的错误信息。
	Start(CCTX CCTX) error

	// Status方法用于获取资源的当前状态。
	Status() SourceState

	// Details方法用于获取资源绑定的详细信息。
	Details() *InEnd

	// Stop方法用于停止资源并释放相关资源。
	Stop()
}
