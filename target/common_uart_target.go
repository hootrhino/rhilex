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

package target

import (
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	serial "github.com/wwhai/goserial"

	"github.com/hootrhino/rhilex/component/hwportmanager"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)

type CommonConfig struct {
	// 是否允许Ping操作
	AllowPing *bool `json:"allowPing" validate:"required"`
	// 数据模式，可以是RAW或HEX
	DataMode string `json:"dataMode" validate:"required"` // RAW|HEX
	// Ping请求的数据包内容
	PingPacket string `json:"pingPacket" validate:"required"`
	// 请求超时时间，单位为秒
	Timeout *int `json:"timeout" validate:"required"`
}

type CommonUartMainConfig struct {
	// 通用配置，包含AllowPing、DataMode、PingPacket、Timeout等
	CommonConfig CommonConfig `json:"commonConfig" validate:"required"`
	// 端口UUID，用于识别特定的串口设备
	PortUuid string `json:"portUuid" validate:"required"`
}

type CommonUart struct {
	typex.XStatus
	hwPortConfig hwportmanager.UartConfig
	status       typex.SourceState
	locker       sync.Mutex
	serialPort   serial.Port
	maxError     int
	mainConfig   CommonUartMainConfig
}

func NewCommonUart(e typex.Rhilex) typex.XTarget {
	mdev := new(CommonUart)
	mdev.RuleEngine = e
	mdev.mainConfig = CommonUartMainConfig{
		PortUuid: "/dev/ttyS1",
		CommonConfig: CommonConfig{
			DataMode:   "RAW",
			PingPacket: "RHILEX",
			AllowPing: func() *bool {
				b := false
				return &b
			}(),
			Timeout: func() *int {
				b := 3000
				return &b
			}(),
		},
	}
	mdev.maxError = 5
	mdev.locker = sync.Mutex{}
	mdev.status = typex.SOURCE_DOWN
	return mdev
}

func (mdev *CommonUart) Init(outEndId string, configMap map[string]interface{}) error {
	mdev.PointId = outEndId
	mdev.maxError = 0

	if err := utils.BindSourceConfig(configMap, &mdev.mainConfig); err != nil {
		return err
	}

	return nil

}
func (mdev *CommonUart) Start(cctx typex.CCTX) error {
	mdev.Ctx = cctx.Ctx
	mdev.CancelCTX = cctx.CancelCTX
	hwPort, err := hwportmanager.GetHwPort(mdev.mainConfig.PortUuid)
	if err != nil {
		return err
	}
	if hwPort.Busy {
		return fmt.Errorf("mdev is busying now, Occupied By:%s", hwPort.OccupyBy)
	}
	switch tCfg := hwPort.Config.(type) {
	case hwportmanager.UartConfig:
		{
			mdev.hwPortConfig = tCfg
		}
	default:
		{
			return fmt.Errorf("invalid config:%s", hwPort.Config)
		}
	}
	config := serial.Config{
		Address:  mdev.hwPortConfig.Uart,
		BaudRate: mdev.hwPortConfig.BaudRate,
		DataBits: mdev.hwPortConfig.DataBits,
		Parity:   mdev.hwPortConfig.Parity,
		StopBits: mdev.hwPortConfig.StopBits,
		Timeout:  time.Duration(mdev.hwPortConfig.Timeout) * time.Millisecond,
	}
	serialPort, err := serial.Open(&config)
	if err != nil {
		glogger.GLogger.Error(err)
		return err
	}
	if *mdev.mainConfig.CommonConfig.AllowPing {
		go func(serialPort serial.Port) {
			ticker := time.NewTicker(time.Duration(time.Second * 5))
			defer ticker.Stop()
			for {
				select {
				case <-mdev.Ctx.Done():
					if mdev.serialPort != nil {
						mdev.serialPort.Close()
					}
					return
				default:
				}
				if mdev.serialPort != nil {
					_, err := mdev.serialPort.Write([]byte(mdev.mainConfig.CommonConfig.PingPacket))
					if err != nil {
						glogger.GLogger.Error(err)
						mdev.maxError++
					}
				}
				<-ticker.C
			}
		}(serialPort)
	}
	mdev.serialPort = serialPort
	mdev.status = typex.SOURCE_UP
	glogger.GLogger.Info("CommonUart started:", mdev.hwPortConfig.Uart)
	return nil
}

func (mdev *CommonUart) Status() typex.SourceState {
	if mdev.maxError >= 5 {
		mdev.status = typex.SOURCE_DOWN
	}
	if mdev.serialPort != nil {
		_, err := mdev.serialPort.Write([]byte(mdev.mainConfig.CommonConfig.PingPacket))
		if err != nil {
			glogger.GLogger.Error(err)
			mdev.maxError++
		}
	}
	return mdev.status
}

/*
*
* 数据写到串口
*
 */
func (mdev *CommonUart) To(data interface{}) (interface{}, error) {
	if mdev.serialPort == nil {
		mdev.maxError++
		return 0, fmt.Errorf("serial Port invalid")
	}
	if mdev.mainConfig.CommonConfig.DataMode == "RAW" {
		switch S := data.(type) {
		case string:
			return mdev.serialPort.Write([]byte(S))
		}
	}
	if mdev.mainConfig.CommonConfig.DataMode == "HEX" {
		switch t := data.(type) {
		case string:
			Hex, err := hex.DecodeString(t)
			if err != nil {
				return nil, err
			}
			return mdev.serialPort.Write(Hex)
		}
	}
	mdev.maxError++
	return 0, fmt.Errorf("invalid data:%v", data)
}

func (mdev *CommonUart) Stop() {
	mdev.status = typex.SOURCE_DOWN
	if mdev.CancelCTX != nil {
		mdev.CancelCTX()
	}
	mdev.serialPort.Close()
}
func (mdev *CommonUart) Details() *typex.OutEnd {
	return mdev.RuleEngine.GetOutEnd(mdev.PointId)
}
