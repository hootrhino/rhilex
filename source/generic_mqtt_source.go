package source

import (
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/hootrhino/rhilex/common"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)

type MqttMessage struct {
	Topic   string `json:"topic"`
	Payload []byte `json:"payload"`
}

type genericMqttSource struct {
	typex.XStatus
	client     mqtt.Client
	mainConfig common.GenericMqttConfig
	status     typex.SourceState
}

func NewGenericMqttSource(e typex.Rhilex) typex.XSource {
	src := new(genericMqttSource)
	src.mainConfig = common.GenericMqttConfig{
		Port:     1883,
		ClientId: fmt.Sprintf("rhilex_mqtt_source_%d", (time.Now().Second())),
	}
	src.RuleEngine = e
	src.status = typex.SOURCE_DOWN
	return src
}

func (tc *genericMqttSource) Init(inEndId string, configMap map[string]interface{}) error {
	tc.PointId = inEndId
	if err := utils.BindSourceConfig(configMap, &tc.mainConfig); err != nil {
		return err
	}
	return nil
}

func (tc *genericMqttSource) Start(cctx typex.CCTX) error {
	tc.Ctx = cctx.Ctx
	tc.CancelCTX = cctx.CancelCTX
	// connect to mqtt
	err := tc.connectToMqtt()
	if err != nil {
		return err
	}
	// subscribe topic
	err = tc.subscribe()
	if err != nil {
		return err
	}
	tc.status = typex.SOURCE_UP
	return nil
}

func (tc *genericMqttSource) connectToMqtt() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%v", tc.mainConfig.Host, tc.mainConfig.Port))
	opts.SetClientID(tc.mainConfig.ClientId)
	opts.SetUsername(tc.mainConfig.Username)
	opts.SetPassword(tc.mainConfig.Password)
	opts.OnConnect = func(client mqtt.Client) {
		glogger.GLogger.Infof("GenericMqtt Connected. inEndId=%v host=%v port=%v clientId=%v",
			tc.PointId, tc.mainConfig.Host, tc.mainConfig.Port, tc.mainConfig.ClientId)
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		glogger.GLogger.Warnf("GenericMqtt Disconnect. err=%v status=%v", err, tc.status)
	}
	opts.SetCleanSession(true)
	opts.SetPingTimeout(30 * time.Second)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetConnectTimeout(5 * time.Second)
	opts.SetAutoReconnect(false)
	opts.SetMaxReconnectInterval(0)

	// sync operation
	tc.client = mqtt.NewClient(opts)
	if token := tc.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (tc *genericMqttSource) subscribe() error {
	filters := make(map[string]byte)
	for _, topic := range tc.mainConfig.SubTopics {
		filters[topic] = byte(tc.mainConfig.Qos)
	}
	multiple := tc.client.SubscribeMultiple(filters, tc.onMessage)

	multiple.Wait()
	<-multiple.Done()
	if multiple.Error() != nil {
		return multiple.Error()
	}
	return nil
}

func (tc *genericMqttSource) onMessage(client mqtt.Client, message mqtt.Message) {
	mqttMessage := MqttMessage{
		Topic:   message.Topic(),
		Payload: message.Payload(),
	}
	msg, err := json.Marshal(mqttMessage)
	if err != nil {
		glogger.GLogger.Error("handle message failed", err)
		return
	}
	work, err := tc.RuleEngine.WorkInEnd(tc.RuleEngine.GetInEnd(tc.PointId), string(msg))
	if !work {
		glogger.GLogger.Error(err)
	}
}

func (tc *genericMqttSource) Status() typex.SourceState {
	if tc.client != nil {
		if tc.client.IsConnectionOpen() {
			return typex.SOURCE_UP
		}
	}
	return tc.status
}

func (tc *genericMqttSource) Details() *typex.InEnd {
	return tc.RuleEngine.GetInEnd(tc.PointId)
}

func (tc *genericMqttSource) Stop() {
	if tc.CancelCTX != nil {
		tc.CancelCTX()
	}
	if tc.client != nil {
		tc.client.Disconnect(1000)
	}
	tc.status = typex.SOURCE_DOWN
}
