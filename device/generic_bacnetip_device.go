package device

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hootrhino/rhilex/component/apiserver/service"
	"net"
	"time"

	bacnet "github.com/hootrhino/gobacnet"
	"github.com/hootrhino/gobacnet/apdus"
	"github.com/hootrhino/gobacnet/btypes"
	"github.com/hootrhino/rhilex/component/intercache"

	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)

type bacnetCommonConfig struct {
	Frequency int `json:"frequency" title:"采集间隔，单位毫秒"`
}
type bacnetConfig struct {
	Mode        string `json:"mode" title:"bacnet运行模式"` // IP/MSTP
	LocalPort   int    `json:"localPort" validate:"required"`
	NetworkCidr string `json:"networkCidr" validate:"required"`
	DeviceId    uint32 `json:"deviceId" validate:"required"`
	VendorId    uint32 `json:"vendorId" validate:"required"`
	// NetWorkId   uint16 `json:"netWorkId" validate:"required"`
}

type bacnetDataPoint struct {
	UUID           string            `json:"uuid"`
	Tag            string            `json:"tag" validate:"required" title:"数据Tag"`
	BacnetDeviceId uint32            `json:"bacnetDeviceId" title:"bacnet设备id"`
	ObjectType     btypes.ObjectType `json:"objectType" title:"object类型"`
	ObjectId       uint32            `json:"objectId" title:"object的id"`

	property btypes.PropertyData
}

type BacnetMainConfig struct {
	BacnetConfig bacnetConfig       `json:"bacnetConfig" validate:"required"`
	CommonConfig bacnetCommonConfig `json:"commonConfig" validate:"required"`
}

type BacnetDataPointConfig struct {
	BacnetDeviceId uint32 `json:"bacnetDeviceId"`
	ObjectType     string `json:"objectType"`
	ObjectId       uint32 `json:"objectId"`
}

type GenericBacnetIpDevice struct {
	typex.XStatus
	bacnetClient bacnet.Client
	status       typex.DeviceState
	RuleEngine   typex.Rhilex
	mainConfig   BacnetMainConfig
	// 点位表
	SubDeviceDataPoints []bacnetDataPoint           // 读到子设备的点位
	SelfPropertyData    map[uint32][2]btypes.Object // 自己的点位
	remoteDeviceMap     map[uint32]btypes.Device
}

func NewGenericBacnetIpDevice(e typex.Rhilex) typex.XDevice {
	g := new(GenericBacnetIpDevice)
	g.RuleEngine = e
	g.mainConfig = BacnetMainConfig{
		CommonConfig: bacnetCommonConfig{Frequency: 1000},
		BacnetConfig: bacnetConfig{
			Mode:      "BROADCAST",
			LocalPort: 47808,
			DeviceId:  2580,
			VendorId:  2580,
		},
	}
	g.SubDeviceDataPoints = make([]bacnetDataPoint, 0)
	g.status = typex.DEV_DOWN
	return g
}

func (dev *GenericBacnetIpDevice) Init(devId string, configMap map[string]interface{}) error {
	dev.PointId = devId
	// 先给个空的
	dev.remoteDeviceMap = make(map[uint32]btypes.Device)

	intercache.RegisterSlot(devId)
	err := utils.BindSourceConfig(configMap, &dev.mainConfig)
	if err != nil {
		return err
	}
	dataPoints, err := service.ListDataPointByUuid(devId)
	if err != nil {
		return err
	}
	for _, mDataPoint := range dataPoints {
		config := BacnetDataPointConfig{}
		err = json.Unmarshal([]byte(mDataPoint.Config), &config)
		if err != nil {
			return err
		}
		dataPoint := bacnetDataPoint{
			UUID:           mDataPoint.UUID,
			Tag:            mDataPoint.Tag,
			BacnetDeviceId: config.BacnetDeviceId,
			ObjectType:     getObjectTypeByNumber(config.ObjectType),
			ObjectId:       config.ObjectId,
		}
		// Cache Value
		intercache.SetValue(dev.PointId, mDataPoint.UUID, intercache.CacheValue{
			UUID:          mDataPoint.UUID,
			Status:        0,
			LastFetchTime: uint64(time.Now().UnixMilli()),
			Value:         "",
			ErrMsg:        "Loading",
		})
		dev.SubDeviceDataPoints = append(dev.SubDeviceDataPoints, dataPoint)
	}
	if err != nil {
		glogger.GLogger.Error(err)
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
	PropertyData := map[uint32][2]btypes.Object{}
	// 将nodeConfig对应的配置信息
	for idx, BacnetDataPoint := range dev.SubDeviceDataPoints {
		SubPropertyData := btypes.PropertyData{
			Object: btypes.Object{
				ID: btypes.ObjectID{
					Type:     BacnetDataPoint.ObjectType,
					Instance: btypes.ObjectInstance(BacnetDataPoint.ObjectId),
				},
				Properties: []btypes.Property{
					{
						Type:       btypes.PropPresentValue, // Present value
						ArrayIndex: btypes.ArrayAll,
					},
				},
			},
		}
		dev.SubDeviceDataPoints[idx].property = SubPropertyData
		// 配置自身的点位
		PropertyData[BacnetDataPoint.ObjectId] = apdus.NewAIPropertyWithRequiredFields(BacnetDataPoint.Tag,
			BacnetDataPoint.ObjectId, float32(0.00), "")
	}
	// 广播模式监听
	if dev.mainConfig.BacnetConfig.Mode == "BROADCAST" {
		// 创建一个bacnet ip的本地网络
		IP, IPNet, errParseCIDR := net.ParseCIDR(dev.mainConfig.BacnetConfig.NetworkCidr)
		if errParseCIDR != nil {
			glogger.GLogger.Error(errParseCIDR)
			return errParseCIDR
		}
		MaskSize, _ := IPNet.Mask.Size()
		client, err := bacnet.NewClient(&bacnet.ClientBuilder{
			Ip:           IP.String(),
			SubnetCIDR:   MaskSize,
			Port:         dev.mainConfig.BacnetConfig.LocalPort,
			DeviceId:     dev.mainConfig.BacnetConfig.DeviceId, // RHILEX 自身的ID
			VendorId:     dev.mainConfig.BacnetConfig.VendorId, // RHILEX 自身的厂家
			NetWorkId:    0,                                    // RHILEX 自身的网络号
			PropertyData: PropertyData,                         // 点位表, 需要更新为动态
		})
		if err != nil {
			return err
		}

		dev.bacnetClient = client
		client.SetLogger(glogger.GLogger.Logger)
		go dev.bacnetClient.StartPoll(dev.Ctx)
		go func(ctx context.Context) {
			// 定时刷新device列表 后续可以优化下逻辑
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					/// 迁移配置到前端控制
					devices, err := client.WhoIs(&bacnet.WhoIsOpts{
						Low:             -1,
						High:            -1,
						GlobalBroadcast: true,
					})
					if err != nil {
						glogger.GLogger.Error(err)
						continue
					}
					if len(devices) > 0 {
						deviceMap := make(map[uint32]btypes.Device)
						for i := range devices {
							deviceMap[uint32(devices[i].DeviceID)] = devices[i]
						}
						dev.remoteDeviceMap = nil
						dev.remoteDeviceMap = deviceMap
					}

				case <-ctx.Done():
					return
				}
			}
		}(dev.Ctx)

	}

	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Duration(dev.mainConfig.CommonConfig.Frequency) * time.Millisecond)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			default:
			}

			ReadBacnetValues := dev.ReadProperty()
			for _, ReadBacnetValue := range ReadBacnetValues {
				if bytes, err := json.Marshal(ReadBacnetValue); err != nil {
					glogger.GLogger.Error(err)
				} else {
					glogger.GLogger.Debug(string(bytes))
					dev.RuleEngine.WorkDevice(dev.Details(), string(bytes))
				}
			}
			<-ticker.C
		}
	}(dev.Ctx)

	dev.status = typex.DEV_UP
	return nil
}

func (dev *GenericBacnetIpDevice) OnRead(cmd []byte, data []byte) (int, error) {
	return 0, nil
}

type ReadBacnetValue struct {
	Tag              string      `json:"tag"`
	DeviceId         uint32      `json:"deviceId"`
	PropertyType     string      `json:"propertyType"`
	PropertyInstance uint32      `json:"propertyInstance"`
	Value            interface{} `json:"value"`
}

/*
*
* 局域网广播
*
 */
func (dev *GenericBacnetIpDevice) ReadProperty() []ReadBacnetValue {
	ReadBacnetValues := []ReadBacnetValue{}
	for _, SubDeviceDataPoint := range dev.SubDeviceDataPoints {
		var bacnetDeviceId uint32
		if dev.mainConfig.BacnetConfig.Mode == "SINGLE" {
			bacnetDeviceId = 1
		} else {
			bacnetDeviceId = SubDeviceDataPoint.BacnetDeviceId
		}
		if device, ok := dev.remoteDeviceMap[bacnetDeviceId]; ok {
			property, err := dev.bacnetClient.ReadProperty(device, SubDeviceDataPoint.property)
			if err != nil {
				glogger.GLogger.Errorf("bacnet Client Read Property failed. tag = %v, err=%v", SubDeviceDataPoint.Tag, err)
				intercache.SetValue(dev.PointId, SubDeviceDataPoint.UUID, intercache.CacheValue{
					UUID:          SubDeviceDataPoint.UUID,
					Status:        0,
					LastFetchTime: uint64(time.Now().UnixMilli()),
					Value:         "",
					ErrMsg:        err.Error(),
				})
				dev.bacnetClient.GetBacnetIPServer().
					UpdateAIPropertyValue(uint32(SubDeviceDataPoint.ObjectId), float32(0))
				continue
			}
			ReadBacnetValue := ReadBacnetValue{
				Tag:              SubDeviceDataPoint.Tag,
				DeviceId:         bacnetDeviceId,
				PropertyType:     property.Object.ID.Type.String(),
				PropertyInstance: uint32(property.Object.ID.Instance),
			}
			if len(property.Object.Properties) > 0 {
				ReadBacnetValue.Value = property.Object.Properties[0].Data
			} else {
				ReadBacnetValue.Value = uint32(0)
			}
			ReadBacnetValues = append(ReadBacnetValues, ReadBacnetValue)
			dev.bacnetClient.GetBacnetIPServer().
				UpdateAIPropertyValue(uint32(SubDeviceDataPoint.ObjectId), ReadBacnetValue.Value)

			intercache.SetValue(dev.PointId, SubDeviceDataPoint.UUID, intercache.CacheValue{
				UUID:          SubDeviceDataPoint.UUID,
				Status:        1,
				LastFetchTime: uint64(time.Now().UnixMilli()),
				Value:         fmt.Sprintf("%v", ReadBacnetValue.Value),
				ErrMsg:        "",
			})
		}
	}
	return ReadBacnetValues
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
	dev.status = typex.DEV_DOWN
	if dev.CancelCTX != nil {
		dev.CancelCTX()
	}
	if dev.bacnetClient != nil {
		dev.bacnetClient.ClientClose(false)
		dev.bacnetClient.Close()
	}
	intercache.UnRegisterSlot(dev.PointId)
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
