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

package device

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	golog "log"
	"sort"
	"strconv"

	"time"

	"github.com/hootrhino/rhilex/common"
	"github.com/hootrhino/rhilex/component/hwportmanager"
	intercache "github.com/hootrhino/rhilex/component/intercache"
	"github.com/hootrhino/rhilex/component/interdb"

	modbus "github.com/hootrhino/gomodbus"
	core "github.com/hootrhino/rhilex/config"
	modbus_device "github.com/hootrhino/rhilex/device/modbus"

	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)

// 这是个通用Modbus采集器, 主要用来在通用场景下采集数据，因此需要配合规则引擎来使用
//
// Modbus 采集到的数据如下, LUA 脚本可做解析, 示例脚本可参照 generic_modbus_parse.lua
//
//	{
//	    "d1":{
//	        "tag":"d1",
//	        "function":3,
//	        "slaverId":1,
//	        "address":0,
//	        "quantity":2,
//	        "value":"..."
//	    },
//	    "d2":{
//	        "tag":"d2",
//	        "function":3,
//	        "slaverId":2,
//	        "address":0,
//	        "quantity":2,
//	        "value":"..."
//	    }
//	}
type ModbusMasterCommonConfig struct {
	Mode           string `json:"mode" validate:"required"`
	AutoRequest    *bool  `json:"autoRequest" validate:"required"`
	BatchRequest   *bool  `json:"batchRequest" validate:"required"` // 批量采集
	EnableOptimize *bool  `json:"enableOptimize" validate:"required"`
	MaxRegNum      uint16 `json:"maxRegNum" validate:"required"`
}
type ModbusMasterConfig struct {
	CommonConfig ModbusMasterCommonConfig `json:"commonConfig" validate:"required"`
	HostConfig   common.HostConfig        `json:"hostConfig"`
	PortUuid     string                   `json:"portUuid"`
}

type ModbusMasterGroupedTag struct {
	Function  int    `json:"function"`
	SlaverId  byte   `json:"slaverId"`
	Address   uint16 `json:"address"`
	Frequency int64  `json:"frequency"`
	Quantity  uint16 `json:"quantity"`
	Registers map[string]*common.RegisterRW
}

func (g *ModbusMasterGroupedTag) String() string {
	tagIds := make([]string, 0, len(g.Registers))
	for k := range g.Registers {
		tagIds = append(tagIds, k)
	}
	str := fmt.Sprintf("func=%v slaveId=%v address=%v quantity=%v frequency=%v tagIds=%v",
		g.Function, g.SlaverId, g.Address, g.Quantity, g.Frequency, tagIds)
	return str
}

type GenericModbusMaster struct {
	typex.XStatus
	status     typex.DeviceState
	RuleEngine typex.Rhilex
	//
	rtuHandler *modbus.RTUClientHandler
	tcpHandler *modbus.TCPClientHandler
	Client     modbus.Client
	//
	mainConfig     ModbusMasterConfig
	retryTimes     int
	hwPortConfig   hwportmanager.UartConfig
	Registers      map[string]*common.RegisterRW
	RegisterGroups []*ModbusMasterGroupedTag
}

/*
*
* 温湿度传感器
*
 */
func NewGenericModbusMaster(e typex.Rhilex) typex.XDevice {
	mdev := new(GenericModbusMaster)
	mdev.RuleEngine = e
	mdev.mainConfig = ModbusMasterConfig{
		CommonConfig: ModbusMasterCommonConfig{
			EnableOptimize: func() *bool {
				b := false
				return &b
			}(),
			AutoRequest: func() *bool {
				b := false
				return &b
			}(),
			BatchRequest: func() *bool {
				b := false
				return &b
			}(),
			MaxRegNum: 32,
		},
		PortUuid:   "/dev/ttyS0",
		HostConfig: common.HostConfig{Host: "127.0.0.1", Port: 502, Timeout: 3000},
	}
	mdev.Registers = map[string]*common.RegisterRW{}
	mdev.Busy = false
	mdev.status = typex.DEV_DOWN
	return mdev
}

//  初始化
func (mdev *GenericModbusMaster) Init(devId string, configMap map[string]interface{}) error {
	mdev.PointId = devId
	mdev.retryTimes = 0
	intercache.RegisterSlot(mdev.PointId)
	if err := utils.BindSourceConfig(configMap, &mdev.mainConfig); err != nil {
		return err
	}
	if !utils.SContains([]string{"UART", "TCP"}, mdev.mainConfig.CommonConfig.Mode) {
		return errors.New("unsupported mode, only can be one of 'TCP' or 'UART'")
	}
	// 合并数据库里面的点位表
	var ModbusPointList []modbus_device.ModbusPoint
	modbusPointLoadErr := interdb.DB().Table("m_modbus_data_points").
		Where("device_uuid=?", devId).Find(&ModbusPointList).Error
	if modbusPointLoadErr != nil {
		return modbusPointLoadErr
	}
	for _, ModbusPoint := range ModbusPointList {
		// 频率不能太快
		if ModbusPoint.Frequency < 50 {
			return errors.New("'frequency' must grate than 50 millisecond")
		}
		mdev.Registers[ModbusPoint.UUID] = &common.RegisterRW{
			UUID:      ModbusPoint.UUID,
			Tag:       ModbusPoint.Tag,
			Alias:     ModbusPoint.Alias,
			Function:  ModbusPoint.Function,
			SlaverId:  ModbusPoint.SlaverId,
			Address:   ModbusPoint.Address,
			Quantity:  ModbusPoint.Quantity,
			Frequency: ModbusPoint.Frequency,
			DataType:  ModbusPoint.DataType,
			DataOrder: ModbusPoint.DataOrder,
			Weight:    ModbusPoint.Weight,
		}
		LastFetchTime := uint64(time.Now().UnixMilli())
		intercache.SetValue(mdev.PointId, ModbusPoint.UUID, intercache.CacheValue{
			UUID:          ModbusPoint.UUID,
			Status:        0,
			LastFetchTime: LastFetchTime,
			Value:         "",
			ErrMsg:        "Device Loading",
		})
	}

	// 开启优化
	if *mdev.mainConfig.CommonConfig.EnableOptimize {
		rws := make([]*common.RegisterRW, len(mdev.Registers))
		idx := 0
		for _, val := range mdev.Registers {
			rws[idx] = val
			idx++
		}
		mdev.RegisterGroups = mdev.groupTags(rws)
		for i, v := range mdev.RegisterGroups {
			glogger.GLogger.Infof("RegisterGroups%v %v", i, v)
		}
	}
	if mdev.mainConfig.CommonConfig.Mode == "UART" {
		hwPort, err := hwportmanager.GetHwPort(mdev.mainConfig.PortUuid)
		if err != nil {
			return err
		}
		if hwPort.Busy {
			return fmt.Errorf("UART is busying now, Occupied By:%s", hwPort.OccupyBy)
		}
		switch tCfg := hwPort.Config.(type) {
		case hwportmanager.UartConfig:
			{
				mdev.hwPortConfig = tCfg
			}
		default:
			{
				return fmt.Errorf("Invalid config:%s", hwPort.Config)
			}
		}
	}
	return nil
}

func (mdev *GenericModbusMaster) groupTags(registers []*common.RegisterRW) []*ModbusMasterGroupedTag {
	/**
	0、分组，Frequency采集时间需要相同
	1、寄存器类型分类
	2、tag排序
	3、限制单次数据采集数量为32个
	4、tag address必须连续
	*/
	sort.Sort(common.RegisterList(registers))
	result := make([]*ModbusMasterGroupedTag, 0)
	for i := 0; i < len(registers); {
		start := i
		end := i
		cursor := i
		tagGroup := &ModbusMasterGroupedTag{
			Function:  registers[start].Function,
			SlaverId:  registers[start].SlaverId,
			Address:   registers[start].Address,
			Frequency: registers[start].Frequency,
		}
		result = append(result, tagGroup)
		tagGroup.Registers = make(map[string]*common.RegisterRW)

		regMaxAddr := uint16(0)
		for end < len(registers) {
			curReg := registers[cursor]
			evaluateReg := registers[end]
			curRegAddr := curReg.Address + curReg.Quantity - 1
			if curRegAddr > regMaxAddr {
				regMaxAddr = curRegAddr
			}
			if tagGroup.SlaverId != evaluateReg.SlaverId {
				break
			}
			if tagGroup.Function != evaluateReg.Function {
				break
			}
			if tagGroup.Frequency != evaluateReg.Frequency {
				break
			}
			if evaluateReg.Address > regMaxAddr+1 {
				break
			}
			totalQuantity := evaluateReg.Address + evaluateReg.Quantity - tagGroup.Address
			if totalQuantity > mdev.mainConfig.CommonConfig.MaxRegNum {
				// 寄存器数量超过单次最大采集寄存器个数
				break
			}
			tagGroup.Registers[evaluateReg.UUID] = evaluateReg
			tagGroup.Quantity = totalQuantity
			cursor = end
			end++
		}
		i = end
	}
	return result
}

// 启动
func (mdev *GenericModbusMaster) Start(cctx typex.CCTX) error {
	mdev.Ctx = cctx.Ctx
	mdev.CancelCTX = cctx.CancelCTX
	mdev.retryTimes = 0
	if mdev.mainConfig.CommonConfig.Mode == "UART" {
		hwPort, err := hwportmanager.GetHwPort(mdev.mainConfig.PortUuid)
		if err != nil {
			return err
		}
		if hwPort.Busy {
			return fmt.Errorf("UART is busying now, Occupied By:%s", hwPort.OccupyBy)
		}

		mdev.rtuHandler = modbus.NewRTUClientHandler(hwPort.Name)
		mdev.rtuHandler.BaudRate = mdev.hwPortConfig.BaudRate
		mdev.rtuHandler.DataBits = mdev.hwPortConfig.DataBits
		mdev.rtuHandler.Parity = mdev.hwPortConfig.Parity
		mdev.rtuHandler.StopBits = mdev.hwPortConfig.StopBits
		// timeout 最大不能超过20, 不然无意义
		mdev.rtuHandler.Timeout = time.Duration(mdev.hwPortConfig.Timeout) * time.Millisecond
		if core.GlobalConfig.AppDebugMode {
			mdev.rtuHandler.Logger = golog.New(glogger.GLogger.Writer(),
				"Modbus RTU Mode: "+mdev.PointId, golog.LstdFlags)
		}

		if err := mdev.rtuHandler.Connect(); err != nil {
			return err
		}
		hwportmanager.SetInterfaceBusy(mdev.mainConfig.PortUuid, hwportmanager.HwPortOccupy{
			UUID: mdev.PointId,
			Type: "DEVICE",
			Name: mdev.Details().Name,
		})
		mdev.Client = modbus.NewClient(mdev.rtuHandler)
	}
	if mdev.mainConfig.CommonConfig.Mode == "TCP" {
		mdev.tcpHandler = modbus.NewTCPClientHandler(
			fmt.Sprintf("%s:%v", mdev.mainConfig.HostConfig.Host, mdev.mainConfig.HostConfig.Port),
		)
		if core.GlobalConfig.AppDebugMode {
			mdev.tcpHandler.Logger = golog.New(glogger.GLogger.Writer(),
				"Modbus TCP Mode: "+mdev.PointId, golog.LstdFlags)
		}
		if err := mdev.tcpHandler.Connect(); err != nil {
			return err
		}
		mdev.Client = modbus.NewClient(mdev.tcpHandler)
	}
	//---------------------------------------------------------------------------------
	// Start
	//---------------------------------------------------------------------------------
	if *mdev.mainConfig.CommonConfig.AutoRequest {
		mdev.retryTimes = 0
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					{
						return
					}
				default:
					{
					}
				}
				ReadRegisterValues := []ReadRegisterValue{}
				if mdev.mainConfig.CommonConfig.Mode == "UART" {
					ReadRegisterValues = mdev.RTURead()
				}
				if mdev.mainConfig.CommonConfig.Mode == "TCP" {
					ReadRegisterValues = mdev.TCPRead()
				}
				if !*mdev.mainConfig.CommonConfig.BatchRequest {
					for _, ReadRegisterValue := range ReadRegisterValues {
						if bytes, errMarshal := json.Marshal(ReadRegisterValue); errMarshal != nil {
							mdev.retryTimes++
							glogger.GLogger.Error(errMarshal)
						} else {
							mdev.RuleEngine.WorkDevice(mdev.Details(), string(bytes))
						}
					}
				} else {
					if bytes, errMarshal := json.Marshal(ReadRegisterValues); errMarshal != nil {
						mdev.retryTimes++
						glogger.GLogger.Error(errMarshal)
					} else {
						mdev.RuleEngine.WorkDevice(mdev.Details(), string(bytes))
					}
				}

			}

		}(mdev.Ctx)
	}

	mdev.status = typex.DEV_UP
	return nil
}
func (mdev *GenericModbusMaster) RTURead() []ReadRegisterValue {
	return mdev.modbusRead()
}
func (mdev *GenericModbusMaster) TCPRead() []ReadRegisterValue {
	return mdev.modbusRead()
}

// 从设备里面读数据出来
func (mdev *GenericModbusMaster) OnRead(cmd []byte, data []byte) (int, error) {
	return 0, nil
}

// 把数据写入设备
func (mdev *GenericModbusMaster) OnWrite(cmd []byte, data []byte) (int, error) {
	RegisterW := common.RegisterW{}
	if err := json.Unmarshal(data, &RegisterW); err != nil {
		return 0, err
	}
	dataMap := [1]common.RegisterW{RegisterW}
	for _, r := range dataMap {
		if mdev.mainConfig.CommonConfig.Mode == "TCP" {
			mdev.tcpHandler.SlaveId = r.SlaverId
		}
		if mdev.mainConfig.CommonConfig.Mode == "UART" {
			mdev.rtuHandler.SlaveId = r.SlaverId
		}
		// 5
		if r.Function == common.WRITE_SINGLE_COIL {
			if len(r.Values) > 0 {
				if r.Values[0] == 0 {
					_, err := mdev.Client.WriteSingleCoil(r.Address,
						binary.BigEndian.Uint16([]byte{0x00, 0x00}))
					if err != nil {
						return 0, err
					}
				}
				if r.Values[0] == 1 {
					_, err := mdev.Client.WriteSingleCoil(r.Address,
						binary.BigEndian.Uint16([]byte{0xFF, 0x00}))
					if err != nil {
						return 0, err
					}
				}

			}

		}
		// 15
		if r.Function == common.WRITE_MULTIPLE_COILS {
			_, err := mdev.Client.WriteMultipleCoils(r.Address, r.Quantity, r.Values)
			if err != nil {
				return 0, err
			}
		}
		// 6
		if r.Function == common.WRITE_SINGLE_HOLDING_REGISTER {
			_, err := mdev.Client.WriteSingleRegister(r.Address, binary.BigEndian.Uint16(r.Values))
			if err != nil {
				return 0, err
			}
		}
		// 16
		if r.Function == common.WRITE_MULTIPLE_HOLDING_REGISTERS {

			_, err := mdev.Client.WriteMultipleRegisters(r.Address,
				uint16(len(r.Values))/2, maybePrependZero(r.Values))
			if err != nil {
				return 0, err
			}
		}
	}
	return 0, nil
}
func maybePrependZero(slice []byte) []byte {
	if len(slice)%2 != 0 {
		slice = append([]byte{0}, slice...)
	}
	return slice
}

// 设备当前状态
func (mdev *GenericModbusMaster) Status() typex.DeviceState {
	// 容错5次
	if mdev.retryTimes > 5 {
		return typex.DEV_DOWN
	}
	return mdev.status
}

// 停止设备
func (mdev *GenericModbusMaster) Stop() {
	mdev.status = typex.DEV_DOWN
	if mdev.CancelCTX != nil {
		mdev.CancelCTX()
	}
	if mdev.mainConfig.CommonConfig.Mode == "UART" {
		hwportmanager.FreeInterfaceBusy(mdev.mainConfig.PortUuid)
		if mdev.rtuHandler != nil {
			mdev.rtuHandler.Close()
		}
	}
	if mdev.mainConfig.CommonConfig.Mode == "TCP" {
		if mdev.tcpHandler != nil {
			mdev.tcpHandler.Close()
		}
	}
	intercache.UnRegisterSlot(mdev.PointId) // 卸载点位表
}

// 真实设备
func (mdev *GenericModbusMaster) Details() *typex.Device {
	return mdev.RuleEngine.GetDevice(mdev.PointId)
}

// 状态
func (mdev *GenericModbusMaster) SetState(status typex.DeviceState) {
	mdev.status = status
}

func (mdev *GenericModbusMaster) OnDCACall(UUID string, Command string, Args interface{}) typex.DCAResult {
	return typex.DCAResult{}
}
func (mdev *GenericModbusMaster) OnCtrl([]byte, []byte) ([]byte, error) {
	return []byte{}, nil
}

/*
*
* 返回给Lua的数据结构,经过精简后的寄存器
*
 */
type ReadRegisterValue struct {
	Tag           string `json:"tag"`
	Alias         string `json:"alias"`
	SlaverId      byte   `json:"slaverId"`
	LastFetchTime uint64 `json:"lastFetchTime"`
	Value         string `json:"value"`
}

/*
*
* 串口模式
*
 */
func (mdev *GenericModbusMaster) modbusRead() []ReadRegisterValue {
	if *mdev.mainConfig.CommonConfig.EnableOptimize {
		return mdev.modbusGroupRead()
	} else {
		return mdev.modbusSingleRead()
	}
}

func (mdev *GenericModbusMaster) modbusSingleRead() []ReadRegisterValue {
	var err error
	var results []byte
	RegisterRWs := []ReadRegisterValue{}
	count := len(mdev.Registers)
	if mdev.Client == nil {
		return RegisterRWs
	}
	// modbusRead: 当读多字节寄存器的时候，需要考虑UTF8
	// Modbus收到的数据全部放进这个全局缓冲区内
	var __modbusReadResult = [256]byte{0} // 放在栈上提高效率
	for uuid, r := range mdev.Registers {
		if mdev.mainConfig.CommonConfig.Mode == "TCP" {
			// 下面这行代码在 SlaveId TCP末实现不会生效
			// 主要和这个库有关，后期要把这个SlaverId拿到点位表里面去
			mdev.tcpHandler.SlaveId = r.SlaverId
		}
		if mdev.mainConfig.CommonConfig.Mode == "UART" {
			mdev.rtuHandler.SlaveId = r.SlaverId
		}
		// 1 字节
		if r.Function == common.READ_COIL {
			results, err = mdev.Client.ReadCoils(r.Address, r.Quantity)
			lastTimes := uint64(time.Now().UnixMilli())
			if err != nil {
				count--
				glogger.GLogger.Error(err)
				mdev.retryTimes++
				intercache.SetValue(mdev.PointId, uuid, intercache.CacheValue{
					UUID:          uuid,
					Status:        1,
					Value:         "",
					LastFetchTime: lastTimes,
					ErrMsg:        err.Error(),
				})
				continue
			}
			// ValidData := [4]byte{0, 0, 0, 0}
			copy(__modbusReadResult[:], results[:])
			Value := utils.ParseModbusValue(r.DataType, r.DataOrder, float32(r.Weight), __modbusReadResult)
			Reg := ReadRegisterValue{
				Tag:           r.Tag,
				SlaverId:      r.SlaverId,
				Alias:         r.Alias,
				Value:         Value,
				LastFetchTime: lastTimes,
			}
			RegisterRWs = append(RegisterRWs, Reg)
			intercache.SetValue(mdev.PointId, uuid, intercache.CacheValue{
				UUID:          uuid,
				Status:        0,
				Value:         Value,
				LastFetchTime: lastTimes,
				ErrMsg:        "",
			})
		}
		// 2 字节
		if r.Function == common.READ_DISCRETE_INPUT {
			results, err = mdev.Client.ReadDiscreteInputs(r.Address, r.Quantity)
			lastTimes := uint64(time.Now().UnixMilli())
			if err != nil {
				count--
				glogger.GLogger.Error(err)
				mdev.retryTimes++
				intercache.SetValue(mdev.PointId, uuid, intercache.CacheValue{
					UUID:          uuid,
					Status:        1,
					Value:         "",
					LastFetchTime: lastTimes,
					ErrMsg:        err.Error(),
				})
				continue
			}
			// ValidData := [4]byte{0, 0, 0, 0}
			copy(__modbusReadResult[:], results[:])
			Value := utils.ParseModbusValue(r.DataType, r.DataOrder, float32(r.Weight), __modbusReadResult)
			Reg := ReadRegisterValue{
				Tag:           r.Tag,
				SlaverId:      r.SlaverId,
				Alias:         r.Alias,
				Value:         Value,
				LastFetchTime: lastTimes,
			}
			RegisterRWs = append(RegisterRWs, Reg)
			intercache.SetValue(mdev.PointId, uuid, intercache.CacheValue{
				UUID:          uuid,
				Status:        0,
				Value:         Value,
				LastFetchTime: lastTimes,
				ErrMsg:        "",
			})
		}
		// 2 字节
		//
		if r.Function == common.READ_HOLDING_REGISTERS {
			results, err = mdev.Client.ReadHoldingRegisters(r.Address, r.Quantity)
			lastTimes := uint64(time.Now().UnixMilli())
			if err != nil {
				count--
				glogger.GLogger.Error(err)
				mdev.retryTimes++
				intercache.SetValue(mdev.PointId, uuid, intercache.CacheValue{
					UUID:          uuid,
					Status:        1,
					Value:         "",
					LastFetchTime: lastTimes,
					ErrMsg:        err.Error(),
				})
				continue
			}
			// ValidData := [4]byte{0, 0, 0, 0}
			copy(__modbusReadResult[:], results[:])
			Value := utils.ParseModbusValue(r.DataType, r.DataOrder, float32(r.Weight), __modbusReadResult)

			Reg := ReadRegisterValue{
				Tag:           r.Tag,
				SlaverId:      r.SlaverId,
				Alias:         r.Alias,
				Value:         Value,
				LastFetchTime: lastTimes,
			}
			RegisterRWs = append(RegisterRWs, Reg)
			intercache.SetValue(mdev.PointId, uuid, intercache.CacheValue{
				UUID:          uuid,
				Status:        0,
				Value:         Value,
				LastFetchTime: lastTimes,
				ErrMsg:        "",
			})

		}
		// 2 字节
		if r.Function == common.READ_INPUT_REGISTERS {
			results, err = mdev.Client.ReadInputRegisters(r.Address, r.Quantity)
			lastTimes := uint64(time.Now().UnixMilli())
			if err != nil {
				count--
				glogger.GLogger.Error(err)
				mdev.retryTimes++
				intercache.SetValue(mdev.PointId, uuid, intercache.CacheValue{
					UUID:          uuid,
					Status:        1,
					Value:         "",
					LastFetchTime: lastTimes,
					ErrMsg:        err.Error(),
				})
				continue
			}
			// ValidData := [4]byte{0, 0, 0, 0}
			copy(__modbusReadResult[:], results[:])
			Value := utils.ParseModbusValue(r.DataType, r.DataOrder, float32(r.Weight), __modbusReadResult)
			Reg := ReadRegisterValue{
				Tag:           r.Tag,
				SlaverId:      r.SlaverId,
				Alias:         r.Alias,
				Value:         Value,
				LastFetchTime: lastTimes,
			}
			RegisterRWs = append(RegisterRWs, Reg)
			intercache.SetValue(mdev.PointId, uuid, intercache.CacheValue{
				UUID:          uuid,
				Status:        0,
				Value:         Value,
				LastFetchTime: lastTimes,
				ErrMsg:        "",
			})
		}
		time.Sleep(time.Duration(r.Frequency) * time.Millisecond)
	}
	return RegisterRWs
}

func (mdev *GenericModbusMaster) modbusGroupRead() []ReadRegisterValue {
	jsonValueGroups := make([]ReadRegisterValue, 0)
	var __modbusReadResult = [256]byte{0} // 放在栈上提高效率

	for _, group := range mdev.RegisterGroups {
		if mdev.mainConfig.CommonConfig.Mode == "TCP" {
			mdev.tcpHandler.SlaveId = group.SlaverId
		}
		if mdev.mainConfig.CommonConfig.Mode == "UART" {
			mdev.rtuHandler.SlaveId = group.SlaverId
		}
		if group.Function == common.READ_COIL {
			buf, err := mdev.Client.ReadCoils(group.Address, group.Quantity)
			if err != nil {
				glogger.GLogger.Error(err)
				mdev.retryTimes++
				continue
			}
			for uuid, r := range group.Registers {
				offsetAddr := r.Address - group.Address
				offsetByte := offsetAddr / uint16(8)
				offsetBit := offsetAddr % uint16(8)
				value := (buf[offsetByte] >> offsetBit) & 0x1
				ts := time.Now().UnixMilli()
				jsonVal := ReadRegisterValue{
					Tag:           r.Tag,
					SlaverId:      r.SlaverId,
					Alias:         r.Alias,
					Value:         strconv.Itoa(int(value)),
					LastFetchTime: uint64(ts),
				}
				jsonValueGroups = append(jsonValueGroups, jsonVal)
				intercache.SetValue(mdev.PointId, uuid, intercache.CacheValue{
					UUID:          uuid,
					Status:        0,
					Value:         strconv.Itoa(int(value)),
					LastFetchTime: uint64(ts),
					ErrMsg:        "",
				})
			}
		}
		if group.Function == common.READ_DISCRETE_INPUT {
			buf, err := mdev.Client.ReadDiscreteInputs(group.Address, group.Quantity)
			if err != nil {
				glogger.GLogger.Error(err)
				mdev.retryTimes++
				continue
			}
			for uuid, r := range group.Registers {
				offsetAddr := r.Address - group.Address
				offsetByte := offsetAddr / uint16(8)
				offsetBit := offsetAddr % uint16(8)
				value := (buf[offsetByte] >> offsetBit) & 0x1

				ts := time.Now().UnixMilli()
				jsonVal := ReadRegisterValue{
					Tag:           r.Tag,
					SlaverId:      r.SlaverId,
					Alias:         r.Alias,
					Value:         strconv.Itoa(int(value)),
					LastFetchTime: uint64(ts),
				}
				jsonValueGroups = append(jsonValueGroups, jsonVal)
				intercache.SetValue(mdev.PointId, uuid, intercache.CacheValue{
					UUID:          uuid,
					Status:        0,
					Value:         strconv.Itoa(int(value)),
					LastFetchTime: uint64(ts),
					ErrMsg:        "",
				})
			}
		}
		if group.Function == common.READ_HOLDING_REGISTERS {
			buf, err := mdev.Client.ReadHoldingRegisters(group.Address, group.Quantity)
			if err != nil {
				glogger.GLogger.Error(err)
				mdev.retryTimes++
				continue
			}
			for uuid, r := range group.Registers {
				offsetByte := (r.Address - group.Address) * 2
				offsetByteEnd := offsetByte + r.Quantity*2
				copy(__modbusReadResult[:], buf[offsetByte:offsetByteEnd])
				value := utils.ParseModbusValue(r.DataType, r.DataOrder, float32(r.Weight), __modbusReadResult)

				ts := time.Now().UnixMilli()
				jsonVal := ReadRegisterValue{
					Tag:           r.Tag,
					SlaverId:      r.SlaverId,
					Alias:         r.Alias,
					Value:         value,
					LastFetchTime: uint64(ts),
				}
				jsonValueGroups = append(jsonValueGroups, jsonVal)

				intercache.SetValue(mdev.PointId, uuid, intercache.CacheValue{
					UUID:          uuid,
					Status:        0,
					Value:         value,
					LastFetchTime: uint64(ts),
					ErrMsg:        "",
				})
			}
		}
		if group.Function == common.READ_INPUT_REGISTERS {
			buf, err := mdev.Client.ReadHoldingRegisters(group.Address, group.Quantity)
			if err != nil {
				glogger.GLogger.Error(err)
				mdev.retryTimes++
				continue
			}
			for uuid, r := range group.Registers {
				offsetByte := (r.Address - group.Address) * 2
				offsetByteEnd := offsetByte + r.Quantity*2
				copy(__modbusReadResult[:], buf[offsetByte:offsetByteEnd])
				value := utils.ParseModbusValue(r.DataType, r.DataOrder, float32(r.Weight), __modbusReadResult)

				ts := time.Now().UnixMilli()
				jsonVal := ReadRegisterValue{
					Tag:           r.Tag,
					SlaverId:      r.SlaverId,
					Alias:         r.Alias,
					Value:         value,
					LastFetchTime: uint64(ts),
				}
				jsonValueGroups = append(jsonValueGroups, jsonVal)

				intercache.SetValue(mdev.PointId, uuid, intercache.CacheValue{
					UUID:          uuid,
					Status:        0,
					Value:         value,
					LastFetchTime: uint64(ts),
					ErrMsg:        "",
				})
			}
		}

		time.Sleep(time.Duration(group.Frequency) * time.Millisecond)
	}
	return jsonValueGroups
}
