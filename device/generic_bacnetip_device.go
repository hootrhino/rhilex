package device

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/component/interdb"
	"time"

	"github.com/BeatTime/bacnet"
	"github.com/BeatTime/bacnet/btypes"

	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)

type bacnetConfig struct {
	Type       string `json:"type" title:"bacnet运行模式"`
	Ip         string `json:"ip,omitempty" title:"bacnet设备ip"`
	Port       int    `json:"port,omitempty" title:"bacnet端口，通常是47808"`
	Boardcast  string `json:"boardcast" title:"bacnet广播地址"`
	SubnetCIDR int    `json:"subnetCidr" title:"子网掩码长度"`
	LocalPort  int    `json:"localPort" title:"本地监听端口，填0表示默认47808（有的模拟器必须本地监听47808才能正常交互）"`
	Frequency  int    `json:"frequency" title:"采集间隔，单位毫秒"`
}

type bacnetDataPoint struct {
	//IsMstp   int    `json:"isMstp,omitempty" title:"是否为mstp设备，若是则设备id和子网号必须填写"`
	//Subnet   int    `json:"subnet,omitempty" title:"子网号"`
	UUID           string            `json:"uuid"`
	Tag            string            `json:"tag" validate:"required" title:"数据Tag"`
	BacnetDeviceId int               `json:"bacnetDeviceId,omitempty" title:"bacnet设备id"`
	ObjectType     btypes.ObjectType `json:"objectType,omitempty" title:"object类型"`
	ObjectId       int               `json:"objectId,omitempty" title:"object的id"`

	property btypes.PropertyData
}

type GenericBacnetIpDevice struct {
	typex.XStatus
	status           typex.DeviceState
	RuleEngine       typex.Rhilex
	BacnetConfig     bacnetConfig
	BacnetDataPoints []bacnetDataPoint
	// Bacnet
	bacnetClient    bacnet.Client
	remoteDeviceMap map[int]btypes.Device
}

func NewGenericBacnetIpDevice(e typex.Rhilex) typex.XDevice {
	g := new(GenericBacnetIpDevice)
	g.RuleEngine = e
	g.BacnetConfig = bacnetConfig{}
	g.BacnetDataPoints = make([]bacnetDataPoint, 0)
	g.status = typex.DEV_DOWN
	return g
}

func (dev *GenericBacnetIpDevice) Init(devId string, configMap map[string]interface{}) error {
	dev.PointId = devId
	err := utils.BindSourceConfig(configMap, &dev.BacnetConfig)
	if err != nil {
		return err
	}
	var dataPoints []model.MBacnetDataPoint
	err = interdb.DB().Table("m_bacnet_data_points").
		Where("device_uuid=?", devId).Find(&dataPoints).Error

	points := make([]bacnetDataPoint, len(dataPoints))
	for i := range dataPoints {
		point := dataPoints[i]
		dataPoint := bacnetDataPoint{
			UUID:           point.UUID,
			Tag:            point.Tag,
			BacnetDeviceId: point.BacnetDeviceId,
			ObjectType:     getObjectTypeByNumber(point.ObjectType),
			ObjectId:       point.ObjectId,
		}
		points[i] = dataPoint
	}
	dev.BacnetDataPoints = points
	if err != nil {
		glogger.GLogger.Error("加载bacnet点位出现错误", err)
		return err
	}

	return nil
}

func getObjectTypeByNumber(strType string) btypes.ObjectType {
	switch strType {
	case "AI":
		return btypes.AnalogInput
	case "AO":
		return btypes.AnalogOutput
	case "AV":
		return btypes.AnalogValue
	case "BI":
		return btypes.BinaryInput
	case "BO":
		return btypes.BinaryOutput
	case "BV":
		return btypes.BinaryValue
	case "MI":
		return btypes.MultiStateInput
	case "MO":
		return btypes.MultiStateOutput
	case "MV":
		return btypes.MultiStateValue
	}
	return btypes.AnalogInput
}

func (dev *GenericBacnetIpDevice) Start(cctx typex.CCTX) error {
	dev.CancelCTX = cctx.CancelCTX
	dev.Ctx = cctx.Ctx

	// 将nodeConfig对应的配置信息
	for idx, v := range dev.BacnetDataPoints {
		tmp := btypes.PropertyData{
			Object: btypes.Object{
				ID: btypes.ObjectID{
					Type:     v.ObjectType,
					Instance: btypes.ObjectInstance(v.ObjectId),
				},
				Properties: []btypes.Property{
					{
						Type:       btypes.PropPresentValue, // Present value
						ArrayIndex: btypes.ArrayAll,
					},
				},
			},
		}
		dev.BacnetDataPoints[idx].property = tmp
	}

	if dev.BacnetConfig.Type == "BOARDCAST" {
		// 创建一个bacnetip的本地网络
		client, err := bacnet.NewClient(&bacnet.ClientBuilder{
			Ip:         dev.BacnetConfig.Boardcast,
			Port:       dev.BacnetConfig.LocalPort,
			SubnetCIDR: dev.BacnetConfig.SubnetCIDR, // 随便填一个，主要为了能够创建Client
		})
		if err != nil {
			return err
		}

		dev.bacnetClient = client
		client.SetLogger(glogger.GLogger.Logger)
		go dev.bacnetClient.ClientRun()
		devices, err := client.WhoIs(&bacnet.WhoIsOpts{
			Low:             -1,
			High:            -1,
			GlobalBroadcast: true,
		})
		if err != nil {
			glogger.GLogger.Error("查找bacnet设备失败", err)
			return err
		}
		// TODO 后续需要改成定时更新设备列表
		dev.remoteDeviceMap = make(map[int]btypes.Device)
		for i := range devices {
			dev.remoteDeviceMap[devices[i].DeviceID] = devices[i]
		}
	}

	if dev.BacnetConfig.Type == "SINGLE" {
		// 创建一个bacnetip的本地网络
		client, err := bacnet.NewClient(&bacnet.ClientBuilder{
			Ip:         "0.0.0.0",
			Port:       dev.BacnetConfig.LocalPort,
			SubnetCIDR: 10, // 随便填一个，主要为了能够创建Client
		})
		if err != nil {
			return err
		}

		dev.bacnetClient = client
		client.SetLogger(glogger.GLogger.Logger)
		go dev.bacnetClient.ClientRun()

		mac := make([]byte, 6)
		fmt.Sscanf(dev.BacnetConfig.Ip, "%d.%d.%d.%d", &mac[0], &mac[1], &mac[2], &mac[3])
		port := uint16(dev.BacnetConfig.Port)
		mac[4] = byte(port >> 8)
		mac[5] = byte(port & 0x00FF)

		dev.remoteDeviceMap = make(map[int]btypes.Device)
		dev.remoteDeviceMap[1] = btypes.Device{
			Addr: btypes.Address{
				MacLen: 6,
				Mac:    mac,
			},
		}
	}

	go func(ctx context.Context) {
		interval := dev.BacnetConfig.Frequency
		if interval == 0 {
			interval = 3000
		}
		ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
		for {
			select {
			case <-ctx.Done():
				{
					ticker.Stop()
					return
				}
			default:
				{
				}
			}

			read, err2 := dev.read()
			if err2 != nil {
				glogger.GLogger.Error(err2)
			} else {
				dev.RuleEngine.WorkDevice(dev.Details(), string(read))
			}
			<-ticker.C
		}
	}(dev.Ctx)

	dev.status = typex.DEV_UP
	return nil
}

func (dev *GenericBacnetIpDevice) OnRead(cmd []byte, data []byte) (int, error) {
	read, err := dev.read()
	if err != nil {
		return 0, err
	}
	len := copy(data, read)
	return len, nil
}

func (dev *GenericBacnetIpDevice) read() ([]byte, error) {
	retMap := map[string]string{}
	for _, v := range dev.BacnetDataPoints {
		if device, ok := dev.remoteDeviceMap[v.BacnetDeviceId]; ok {
			property, err := dev.bacnetClient.ReadProperty(device, v.property)
			if err != nil {
				glogger.GLogger.Errorf("read failed. tag = %v, err=%v", v.Tag, err)
				continue
			}
			value := fmt.Sprintf("%v", property.Object.Properties[0].Data)
			retMap[v.Tag] = value
		}
	}
	bytes, _ := json.Marshal(retMap)
	glogger.GLogger.Debugf("%v", retMap)
	return bytes, nil
}

func (dev *GenericBacnetIpDevice) OnWrite(cmd []byte, data []byte) (int, error) {
	//TODO implement me
	return 0, errors.New("not Support")
}

func (dev *GenericBacnetIpDevice) OnCtrl(cmd []byte, args []byte) ([]byte, error) {
	return nil, errors.New("not Support")
}

func (dev *GenericBacnetIpDevice) Status() typex.DeviceState {
	return dev.status
}

func (dev *GenericBacnetIpDevice) Stop() {
	dev.CancelCTX()
	if dev.bacnetClient != nil {
		dev.bacnetClient.Close()
	}
}

func (dev *GenericBacnetIpDevice) Details() *typex.Device {
	return dev.RuleEngine.GetDevice(dev.PointId)
}

func (dev *GenericBacnetIpDevice) SetState(state typex.DeviceState) {
	dev.status = state
}

func (dev *GenericBacnetIpDevice) OnDCACall(UUID string, Command string, Args interface{}) typex.DCAResult {
	return typex.DCAResult{}
}
