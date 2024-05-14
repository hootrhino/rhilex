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

package device

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)

const (
	// 属性
	_tencent_iothub_PropertyTopic   = "$thing/down/property/%v/%v"
	_tencent_iothub_PropertyUpTopic = "$thing/up/property/%v/%v"
	// 动作
	_tencent_iothub_ActionTopic   = "$thing/down/action/%v/%v"
	_tencent_iothub_ActionUpTopic = "$thing/up/action/%v/%v"
	// 子设备拓扑
	_tencent_iothub_operation_up   = "$gateway/operation/%s/%s"        //数据上行 Topic（用于发布）
	_tencent_iothub_operation_down = "$gateway/operation/result/%s/%s" //数据下行 Topic（用于订阅）
)

type TencentIoTGatewayConfig struct {
	// "tcp://$.iotcloud.tencentdevices.com:1883"
	ServerEndpoint string `json:"serverEndpoint" validate:"required"` //服务地址
	Mode           string `json:"mode" validate:"required"`           //模式: DEVICE|GATEWAY
	ProductId      string `json:"productId" validate:"required"`      //产品名
	DeviceName     string `json:"deviceName" validate:"required"`     //设备名
	DevicePsk      string `json:"devicePsk" validate:"required"`      //秘钥
	ClientId       string `json:"clientId" validate:"required"`       //客户端ID
}
type TencentIoTGatewayMainConfig struct {
	CommonConfig TencentIoTGatewayConfig `json:"commonConfig" validate:"required"` // 通用配置
}

// 腾讯云物联网平台网关
type TencentIoTGateway struct {
	typex.XStatus
	status            typex.DeviceState
	mainConfig        TencentIoTGatewayMainConfig
	client            mqtt.Client
	mqttClientId      string
	mqttUsername      string
	mqttPassword      string
	propertyUpTopic   string
	propertyDownTopic string
	actionUpTopic     string
	actionDownTopic   string
	topologyTopicUp   string
	topologyTopicDown string
}

func NewTencentIoTGateway(e typex.Rhilex) typex.XDevice {
	hd := new(TencentIoTGateway)
	hd.RuleEngine = e
	hd.mainConfig = TencentIoTGatewayMainConfig{
		CommonConfig: TencentIoTGatewayConfig{
			Mode:           "DEVICE",
			ServerEndpoint: "tcp://%s.iotcloud.tencentdevices.com:1883",
			ProductId:      "",
			DeviceName:     "",
			DevicePsk:      "",
			ClientId:       "",
		},
	}
	return hd
}

//  初始化
func (hd *TencentIoTGateway) Init(devId string, configMap map[string]interface{}) error {
	hd.PointId = devId
	if err := utils.BindSourceConfig(configMap, &hd.mainConfig); err != nil {
		glogger.GLogger.Error(err)
		return err
	}
	Info, err := GenerateTencentIotMQTTAuthInfo(hd.mainConfig.CommonConfig.ProductId,
		hd.mainConfig.CommonConfig.DeviceName, hd.mainConfig.CommonConfig.DevicePsk)
	if err != nil {
		glogger.GLogger.Error(err)
		return err
	}
	glogger.GLogger.Debug(Info.ToJSONString())
	hd.mqttClientId = Info.ClientID
	hd.mqttUsername = Info.UserName
	hd.mqttPassword = Info.Password
	return nil
}

// 启动
func (hd *TencentIoTGateway) Start(cctx typex.CCTX) error {
	hd.Ctx = cctx.Ctx
	hd.CancelCTX = cctx.CancelCTX
	// 属性
	hd.propertyDownTopic = fmt.Sprintf(_tencent_iothub_PropertyTopic,
		hd.mainConfig.CommonConfig.ProductId, hd.mainConfig.CommonConfig.DeviceName)
	hd.propertyUpTopic = fmt.Sprintf(_tencent_iothub_PropertyUpTopic,
		hd.mainConfig.CommonConfig.ProductId, hd.mainConfig.CommonConfig.DeviceName)
	hd.actionDownTopic = fmt.Sprintf(_tencent_iothub_ActionTopic,
		hd.mainConfig.CommonConfig.ProductId, hd.mainConfig.CommonConfig.DeviceName)
	hd.actionUpTopic = fmt.Sprintf(_tencent_iothub_ActionUpTopic,
		hd.mainConfig.CommonConfig.ProductId, hd.mainConfig.CommonConfig.DeviceName)
	hd.topologyTopicUp = fmt.Sprintf(_tencent_iothub_operation_up,
		hd.mainConfig.CommonConfig.ProductId, hd.mainConfig.CommonConfig.DeviceName)
	hd.topologyTopicDown = fmt.Sprintf(_tencent_iothub_operation_down,
		hd.mainConfig.CommonConfig.ProductId, hd.mainConfig.CommonConfig.DeviceName)

	var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
		glogger.GLogger.Infof("IOTHUB Connected Success")
		// 属性下发
		if err := hd.subscribe(hd.propertyDownTopic); err != nil {
			glogger.GLogger.Error(err)
		}
		// 动作下发
		if err := hd.subscribe(hd.actionDownTopic); err != nil {
			glogger.GLogger.Error(err)
		}
		// 网关模式
		//    数据上行 Topic（用于发布）：$gateway/operation/${productid}/${devicename}
		//    数据下行 Topic（用于订阅）：$gateway/operation/result/${productid}/${devicename}
		if hd.mainConfig.CommonConfig.Mode == "GATEWAY" {
			if err := hd.subscribe(hd.topologyTopicDown); err != nil {
				glogger.GLogger.Error(err)
			} else {
				// Get Topology: {"type": "describe_sub_devices"}
				glogger.GLogger.Info("Connect iothub with Gateway Mode")
				hd.client.Publish(hd.topologyTopicUp, 1, false, `{"type": "describe_sub_devices"}`)
			}
		}
	}

	var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
		glogger.GLogger.Warnf("IOTHUB Disconnect: %v, %v try to reconnect", err, hd.status)
	}

	opts := mqtt.NewClientOptions()
	// iotcloud.tencentdevices.com
	// endPoint := "tcp://%s.iotcloud.tencentdevices.com:1883"
	// hd.mainConfig.CommonConfig.ServerEndpoint = fmt.Sprintf(endPoint, hd.mainConfig.CommonConfig.ProductId)
	endPoint := "tcp://192.168.1.170:1883"
	opts.AddBroker(endPoint)
	opts.SetClientID(hd.mqttClientId)
	opts.SetUsername(hd.mqttUsername)
	opts.SetPassword(hd.mqttPassword)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	opts.SetCleanSession(true)
	opts.SetPingTimeout(30 * time.Second)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetConnectTimeout(5 * time.Second)
	opts.SetAutoReconnect(false)
	opts.SetMaxReconnectInterval(0)
	hd.client = mqtt.NewClient(opts)
	if token := hd.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	hd.status = typex.DEV_UP
	return nil
}

// 设备当前状态
func (hd *TencentIoTGateway) Status() typex.DeviceState {
	if hd.client != nil {
		if hd.client.IsConnectionOpen() {
			return typex.DEV_UP
		}
	}
	return hd.status
}

// 停止设备
func (hd *TencentIoTGateway) Stop() {
	hd.status = typex.DEV_DOWN
	if hd.CancelCTX != nil {
		hd.CancelCTX()
	}
	if hd.client != nil {
		hd.client.Disconnect(50)
	}
}

// 真实设备
func (hd *TencentIoTGateway) Details() *typex.Device {
	return hd.RuleEngine.GetDevice(hd.PointId)
}

// 状态
func (hd *TencentIoTGateway) SetState(status typex.DeviceState) {
	hd.status = status

}

func (hd *TencentIoTGateway) OnDCACall(UUID string, Command string, Args interface{}) typex.DCAResult {
	return typex.DCAResult{}
}

func (hd *TencentIoTGateway) OnCtrl(cmd []byte, args []byte) ([]byte, error) {
	return []byte{}, nil
}

func (hd *TencentIoTGateway) OnRead(cmd []byte, data []byte) (int, error) {

	return 0, nil
}

func (hd *TencentIoTGateway) OnWrite(cmd []byte, b []byte) (int, error) {
	return 0, nil
}

func (hd *TencentIoTGateway) subscribe(topic string) error {
	token := hd.client.Subscribe(topic, 1, func(c mqtt.Client, msg mqtt.Message) {
		glogger.GLogger.Debug(topic, msg)
	})
	if token.Error() != nil {
		return token.Error()
	} else {
		return nil
	}
}

// TencentIotHubMQTTAuthInfo 腾讯云 Iot Hub MQTT 认证信息
type TencentIotHubMQTTAuthInfo struct {
	ClientID string `json:"clientid"`
	UserName string `json:"username"`
	Password string `json:"password"`
}

// GenerateTencentIotMQTTAuthInfo 生成腾讯云iot的mqtt密码
// https://cloud.tencent.com/document/product/634/32546
// productID 产品号
// deviceName 设备名
// psk 设备对称密钥
func GenerateTencentIotMQTTAuthInfo(productID, deviceName, psk string) (TencentIotHubMQTTAuthInfo, error) {
	var ai TencentIotHubMQTTAuthInfo
	var err error
	connid := randConnID(5)
	expiry := time.Now().Add(365 * 24 * time.Hour).Unix()
	ai.ClientID = productID + deviceName
	ai.UserName = fmt.Sprintf("%v;%v;%v;%v", ai.ClientID, 12010126, connid, expiry)
	key, err := base64.StdEncoding.DecodeString(psk)
	if err != nil {
		return ai, err
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(ai.UserName))
	token := hex.EncodeToString(mac.Sum(nil))
	ai.Password = token + ";hmacsha256"
	return ai, err
}

// ToJSONString 将TencentIotHubMQTTAuthInfo转换为json string
func (ai TencentIotHubMQTTAuthInfo) ToJSONString() string {
	data, _ := json.Marshal(ai)
	return string(data)
}

// randConnID 随机产生一个长度为n的connid
func randConnID(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	s := make([]byte, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
