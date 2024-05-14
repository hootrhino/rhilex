package device

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/hootrhino/rhilex/component/hwportmanager"

	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
	serial "github.com/wwhai/goserial"
)

type _UartCommonConfig struct {
	Tag         string `json:"tag" validate:"required"`
	AutoRequest *bool  `json:"autoRequest" validate:"required"`
}

type _UartMainConfig struct {
	CommonConfig _UartCommonConfig `json:"commonConfig" validate:"required"`
	PortUuid     string            `json:"portUuid" validate:"required"`
}

type genericUartDevice struct {
	typex.XStatus
	serialPort   serial.Port
	hwPortConfig hwportmanager.UartConfig
	status       typex.DeviceState
	RuleEngine   typex.Rhilex
	mainConfig   _UartMainConfig
	locker       sync.Locker
}

/*
*
* 通用串口透传
*
 */
func NewGenericUartDevice(e typex.Rhilex) typex.XDevice {
	uart := new(genericUartDevice)
	uart.locker = &sync.Mutex{}
	uart.mainConfig = _UartMainConfig{
		CommonConfig: _UartCommonConfig{
			Tag: "uart",
			AutoRequest: func() *bool {
				b := true
				return &b
			}(),
		},
	}
	uart.RuleEngine = e
	return uart
}

//  初始化
func (uart *genericUartDevice) Init(devId string, configMap map[string]interface{}) error {
	uart.PointId = devId

	if err := utils.BindSourceConfig(configMap, &uart.mainConfig); err != nil {
		glogger.GLogger.Error(err)
		return err
	}

	hwPort, err := hwportmanager.GetHwPort(uart.mainConfig.PortUuid)
	if err != nil {
		return err
	}
	if hwPort.Busy {
		return fmt.Errorf("UART is busying now, Occupied By:%s", hwPort.OccupyBy)
	}
	switch tCfg := hwPort.Config.(type) {
	case hwportmanager.UartConfig:
		{
			uart.hwPortConfig = tCfg
		}
	default:
		{
			return fmt.Errorf("invalid config:%s", hwPort.Config)
		}
	}
	return nil
}

// 启动
func (uart *genericUartDevice) Start(cctx typex.CCTX) error {
	uart.Ctx = cctx.Ctx
	uart.CancelCTX = cctx.CancelCTX

	config := serial.Config{
		Address:  uart.hwPortConfig.Uart,
		BaudRate: uart.hwPortConfig.BaudRate,
		DataBits: uart.hwPortConfig.DataBits,
		Parity:   uart.hwPortConfig.Parity,
		StopBits: uart.hwPortConfig.StopBits,
		Timeout:  time.Duration(50) * time.Millisecond, // 固定写法，表示串口最小一个包耗时，一般50毫秒足够
	}
	serialPort, err := serial.Open(&config)
	if err != nil {
		glogger.GLogger.Error("serialPort start failed:", err)
		return err
	}

	hwportmanager.SetInterfaceBusy(uart.mainConfig.PortUuid,
		hwportmanager.HwPortOccupy{
			UUID: uart.PointId,
			Type: "DEVICE",
			Name: uart.Details().Name,
		})
	uart.serialPort = serialPort
	if !*uart.mainConfig.CommonConfig.AutoRequest {
		uart.status = typex.DEV_UP
		return nil
	}
	go func(ctx context.Context) {
		result := [2048]byte{}
		sliceTimer := time.NewTimer((50) * time.Millisecond)
		sliceTimer.Stop()
		peerCount := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-sliceTimer.C:
				// glogger.GLogger.Debug(result[:peerCount])
				mapV := map[string]string{
					"tag":   uart.mainConfig.CommonConfig.Tag,
					"value": hex.EncodeToString(result[:peerCount]),
				}
				bytes, _ := json.Marshal(mapV)
				uart.RuleEngine.WorkDevice(uart.Details(), string(bytes))
				peerCount = 0 // init index
			default:
				n, errR := io.ReadAtLeast(uart.serialPort, result[peerCount:], 1)
				if errR != nil {
					if !strings.Contains(errR.Error(), "timeout") {
						glogger.GLogger.Error(errR)
					}
				}
				if n != 0 {
					peerCount += n
					sliceTimer.Reset((50) * time.Millisecond)
				}
			}

		}
	}(uart.Ctx)
	uart.status = typex.DEV_UP
	return nil
}

// 从设备里面读数据出来:
//
//	{
//	    "tag":"data tag",
//	    "value":"value s"
//	}
//
// t1.txt="OK"\xff\xff\xff
func (uart *genericUartDevice) OnCtrl(cmd []byte, args []byte) ([]byte, error) {
	result := [2048]byte{}
	if string(cmd) == "HEX" {
		hexs, err1 := hex.DecodeString(string(cmd))
		if err1 != nil {
			glogger.GLogger.Error(err1)
			return nil, err1
		}
		n, errSliceRequest := utils.SliceRequest(uart.Ctx, uart.serialPort,
			hexs, result[:], false, (50)*time.Millisecond)
		if errSliceRequest != nil {
			return []byte{}, errSliceRequest
		}
		return result[:n], nil
	}
	if string(cmd) == "STRING" {
		n, err := uart.serialPort.Write(args)
		if err != nil {
			return nil, err
		}
		// n, errSliceRequest := utils.SliceRequest(uart.Ctx, uart.serialPort,
		// 	args, result[:], false, (50)*time.Millisecond)
		// if errSliceRequest != nil {
		// 	return []byte{}, errSliceRequest
		// }
		return result[:n], nil
	}
	return []byte{}, fmt.Errorf("unsupported cmd, must one of : STRING|HEX")
}

// 设备当前状态
func (uart *genericUartDevice) Status() typex.DeviceState {
	if uart.serialPort != nil {
		_, err := uart.serialPort.Write([]byte("\r\n"))
		if err != nil {
			uart.status = typex.DEV_DOWN
		}
	} else {
		uart.status = typex.DEV_DOWN
	}
	return uart.status
}

// 停止设备
func (uart *genericUartDevice) Stop() {
	uart.status = typex.DEV_DOWN
	if uart.CancelCTX != nil {
		uart.CancelCTX()
	}
	if uart.serialPort != nil {
		uart.serialPort.Close()
		hwportmanager.FreeInterfaceBusy(uart.mainConfig.PortUuid)
	}

}

// 真实设备
func (uart *genericUartDevice) Details() *typex.Device {
	return uart.RuleEngine.GetDevice(uart.PointId)
}

// 状态
func (uart *genericUartDevice) SetState(status typex.DeviceState) {
	uart.status = status

}

// --------------------------------------------------------------------------------------------------
//
// --------------------------------------------------------------------------------------------------

func (uart *genericUartDevice) OnDCACall(UUID string, Command string, Args interface{}) typex.DCAResult {
	return typex.DCAResult{}
}
func (uart *genericUartDevice) OnRead(cmd []byte, data []byte) (int, error) {
	return 0, nil
}

// 把数据写入设备
func (uart *genericUartDevice) OnWrite(cmd []byte, b []byte) (int, error) {
	return 0, nil
}
