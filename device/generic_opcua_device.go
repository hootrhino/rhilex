package device

import (
	"sync"

	"github.com/gopcua/opcua"

	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)

type OpcuaNode struct {
	Tag         string `json:"tag" validate:"required"`
	Description string `json:"description" validate:"required"`
	NodeID      string `json:"nodeId" validate:"required"` // "NodeID" example:"ns=1;s=Test"
	DataType    string `json:"dataType"`
	Value       string `json:"value"`
}

// OpcCommonConfig 定义了OPC通信的通用配置。
type OpcCommonConfig struct {
	// Timeout 指定了通信的超时时间（以毫秒为单位）。
	// validate:"required" 表示这是一个必填字段。
	Timeout *int `json:"timeout" validate:"required"`

	// AutoRequest 指定是否自动发送请求。
	// validate:"required" 表示这是一个必填字段。
	AutoRequest *bool `json:"autoRequest" validate:"required"`
}

// OpcUAConfig 定义了OPC UA特定的配置。
type OpcUAConfig struct {
	// Endpoint 是OPC UA服务器的URL。
	// title:"服务器URL" 提供了字段的说明。
	// example:"opc.tcp://NOAH:53530/OPCUA/SimulationServer" 提供了示例URL。
	Endpoint string `json:"endpoint" title:"服务器URL"`

	// Policy 是消息的安全模式。
	// title:"消息安全模式" 提供了字段的说明。
	// 可选的模式包括：无、Basic128Rsa15、Basic256、Basic256Sha256。
	Policy string `json:"policy" title:"消息安全模式"`

	// Mode 是消息的安全模式。
	// title:"消息安全模式" 提供了字段的说明。
	// 可选的模式包括：无、签名、签名加密。
	Mode string `json:"mode" title:"消息安全模式"`

	// Auth 是认证方式。
	// title:"认证方式" 提供了字段的说明。
	// 可选的模式包括：匿名、用户名。
	Auth string `json:"auth" title:"认证方式"`

	// Username 是用于认证的用户名。
	Username string `json:"username" title:"用户名"`

	// Password 是用于认证的密码。
	Password string `json:"password" title:"密码"`
}

// OpcUAMainConfig 是OPC UA配置的主要结构体，包含通用配置和OPC UA特定配置。
type OpcUAMainConfig struct {
	// OpcCommonConfig 包含OPC通信的通用配置。
	OpcCommonConfig OpcCommonConfig `json:"commonConfig" validate:"required"`

	// OpcUAConfig 包含OPC UA特定的配置。
	OpcUAConfig OpcUAConfig `json:"opcUAConfig" validate:"required"`
}

type genericOpcuaDevice struct {
	typex.XStatus
	status     typex.DeviceState
	RuleEngine typex.Rhilex
	driver     typex.XExternalDriver
	client     *opcua.Client
	mainConfig OpcUAMainConfig
	locker     sync.Mutex
	OpcNodes   []OpcuaNode
	errorCount int
}

func NewGenericOpcuaDevice(e typex.Rhilex) typex.XDevice {
	opc := new(genericOpcuaDevice)
	opc.RuleEngine = e
	opc.locker = sync.Mutex{}
	opc.mainConfig = OpcUAMainConfig{
		OpcCommonConfig: OpcCommonConfig{
			AutoRequest: func() *bool {
				b := false
				return &b
			}(),
			Timeout: func() *int {
				b := 3000
				return &b
			}(),
		},
		OpcUAConfig: OpcUAConfig{
			Endpoint: "opc.tcp://localhost:4840",
			Policy:   "Basic256",
			Mode:     "SignAndEncrypt",
			Auth:     "USERNAME",
			Username: "admin",
			Password: "admin",
		},
	}
	opc.Busy = false
	opc.errorCount = 5
	opc.status = typex.DEV_DOWN
	return opc
}

// 初始化配置文件
func (opcDev *genericOpcuaDevice) Init(devId string, configMap map[string]interface{}) error {
	opcDev.PointId = devId
	if err := utils.BindSourceConfig(configMap, &opcDev.mainConfig); err != nil {
		return err
	}
	return nil
}

func (opcDev *genericOpcuaDevice) Start(cctx typex.CCTX) error {
	opcDev.Ctx = cctx.Ctx
	opcDev.CancelCTX = cctx.CancelCTX

	opcDev.status = typex.DEV_UP
	return nil
}

func (opcDev *genericOpcuaDevice) OnRead(cmd []byte, data []byte) (int, error) {
	return 0, nil
}
func (opcDev *genericOpcuaDevice) OnCtrl(cmd []byte, args []byte) ([]byte, error) {
	return []byte{}, nil
}

func (opcDev *genericOpcuaDevice) OnWrite(cmd []byte, data []byte) (int, error) {
	return 0, nil
}

// 设备当前状态
func (opcDev *genericOpcuaDevice) Status() typex.DeviceState {
	return opcDev.status
}

// 停止设备
func (opcDev *genericOpcuaDevice) Stop() {
	opcDev.status = typex.DEV_DOWN
	opcDev.CancelCTX()
	if opcDev.driver != nil {
		opcDev.client.Close(opcDev.Ctx)
		opcDev.driver.Stop()
	}
}

// 真实设备
func (opcDev *genericOpcuaDevice) Details() *typex.Device {
	return opcDev.RuleEngine.GetDevice(opcDev.PointId)
}

// 状态
func (opcDev *genericOpcuaDevice) SetState(status typex.DeviceState) {
	opcDev.status = status

}

func (opcDev *genericOpcuaDevice) OnDCACall(UUID string, Command string, Args interface{}) typex.DCAResult {
	return typex.DCAResult{}
}
