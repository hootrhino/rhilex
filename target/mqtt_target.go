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

package target

import (
	"errors"
	"fmt"

	"github.com/hootrhino/rhilex/common"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

//
/*
*
* 单向的MQTT客户端，不支持subscribe，订阅了不生效
*
 */
type mqttOutEndTarget struct {
	typex.XStatus
	client     mqtt.Client
	mainConfig common.MqttConfig
	status     typex.SourceState
}

func NewMqttTarget(e typex.Rhilex) typex.XTarget {
	m := new(mqttOutEndTarget)
	m.RuleEngine = e
	m.mainConfig = common.MqttConfig{
		Host: "127.0.0.1",
		Port: 1883,
	}
	m.status = typex.SOURCE_DOWN
	return m
}

func (mq *mqttOutEndTarget) Init(outEndId string, configMap map[string]interface{}) error {
	mq.PointId = outEndId
	if err := utils.BindSourceConfig(configMap, &mq.mainConfig); err != nil {
		return err
	}
	return nil
}
func (mq *mqttOutEndTarget) Start(cctx typex.CCTX) error {
	mq.Ctx = cctx.Ctx
	mq.CancelCTX = cctx.CancelCTX
	//
	//
	var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
		glogger.GLogger.Infof("Mqtt OutEnd Connected Success")
	}

	var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
		glogger.GLogger.Warn("Mqtt Connect lost:", err)
	}
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%v", mq.mainConfig.Host, mq.mainConfig.Port))
	opts.SetClientID(mq.mainConfig.ClientId)
	opts.SetUsername(mq.mainConfig.Username)
	opts.SetPassword(mq.mainConfig.Password)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	opts.CleanSession = true
	opts.SetAutoReconnect(false)    //不需要自动重连, 交给RHILEX管理
	opts.SetMaxReconnectInterval(0) // 不需要自动重连, 交给RHILEX管理
	mq.client = mqtt.NewClient(opts)
	token := mq.client.Connect()
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	mq.status = typex.SOURCE_UP
	return nil
}

func (mq *mqttOutEndTarget) DataModels() []typex.XDataModel {
	return []typex.XDataModel{}
}

func (mq *mqttOutEndTarget) Stop() {
	mq.status = typex.SOURCE_DOWN
	if mq.CancelCTX != nil {
		mq.CancelCTX()
	}
	if mq.client != nil {
		mq.client.Disconnect(0)
	}
}

func (mq *mqttOutEndTarget) Status() typex.SourceState {
	if mq.client != nil {
		if mq.client.IsConnected() {
			return typex.SOURCE_UP
		}
		return typex.SOURCE_DOWN
	}
	return mq.status
}

func (mq *mqttOutEndTarget) Details() *typex.OutEnd {
	return mq.RuleEngine.GetOutEnd(mq.PointId)
}

func (mq *mqttOutEndTarget) To(data interface{}) (interface{}, error) {
	if mq.client != nil {
		switch s := data.(type) {
		case string:
			// glogger.GLogger.Debug("mqtt Target publish:", mq.mainConfig.PubTopic, 1, false, data)
			token := mq.client.Publish(mq.mainConfig.PubTopic, 1, false, s)
			return nil, token.Error()
		default:
			return nil, errors.New("invalid mqtt data type")
		}
	}
	return nil, errors.New("mqtt client is nil")
}
