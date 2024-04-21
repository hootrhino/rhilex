package typex

import "context"

// InEndType
type InEndType string

func (i InEndType) String() string {
	return string(i)
}

const (
	MQTT            InEndType = "MQTT"
	HTTP            InEndType = "HTTP"
	COAP            InEndType = "COAP"
	GRPC            InEndType = "GRPC"
	NATS_SERVER     InEndType = "NATS_SERVER"
	RHILEX_UDP      InEndType = "RHILEX_UDP"
	GENERIC_IOT_HUB InEndType = "GENERIC_IOT_HUB"
	INTERNAL_EVENT  InEndType = "INTERNAL_EVENT" // 内部消息
	GENERIC_MQTT    InEndType = "GENERIC_MQTT"   // 通用MQTT
)

// XStatus for source status
type XStatus struct {
	PointId    string             // Input: Source; Output: Target
	Enable     bool               // 是否开启
	Ctx        context.Context    // context
	CancelCTX  context.CancelFunc // cancel
	RuleEngine Rhilex             // rhilex
	Busy       bool               // 是否处于忙碌状态, 防止请求拥挤
}

// XSource 接口代表了一个终端资源，例如实际的MQTT客户端。
// 它定义了与资源交互所需的一系列方法，包括测试资源可用性、初始化、启动、数据传输等。
type XSource interface {
	// Test方法用于测试资源是否可用。
	// inEndId是资源的标识符。
	// 返回测试结果，如果资源可用则返回true，否则返回false。
	Test(inEndId string) bool

	// Init方法用于初始化资源，传递资源配置信息。
	// inEndId是资源的标识符，configMap是资源配置的映射。
	// 返回初始化是否成功的错误信息。
	Init(inEndId string, configMap map[string]interface{}) error

	// Start方法用于启动资源。
	// CCTX是上下文，具体作用取决于资源的实现。
	// 返回启动是否成功的错误信息。
	Start(CCTX CCTX) error

	// DataModels方法用于获取资源支持的数据模型列表。
	// 这些模型对应于云平台的物模型。
	DataModels() []XDataModel

	// Status方法用于获取资源的当前状态。
	Status() SourceState

	// Details方法用于获取资源绑定的详细信息。
	Details() *InEnd

	// Stop方法用于停止资源并释放相关资源。
	Stop()

	// DownStream方法用于处理下行数据，即从云平台发送到本地资源的数据。
	// 接收一个字节切片作为数据。
	// 返回实际处理的数据长度和错误信息。
	DownStream([]byte) (int, error)

	// UpStream方法用于处理上行数据，即从本地资源发送到云平台的数据。
	// 接收一个字节切片作为数据。
	// 返回实际处理的数据长度和错误信息。
	UpStream([]byte) (int, error)
}
