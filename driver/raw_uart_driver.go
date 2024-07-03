// uart_driver相当于是升级版，这个是最原始的基础驱动
package driver

import (
	"context"
	"errors"
	"sync"

	serial "github.com/hootrhino/goserial"
	"github.com/hootrhino/rhilex/typex"
)

type rawUartDriver struct {
	state      typex.DriverState
	serialPort serial.Port
	ctx        context.Context
	RuleEngine typex.Rhilex
	device     *typex.Device
	locker     sync.Mutex
}

// 初始化一个驱动
func NewRawUartDriver(
	ctx context.Context,
	e typex.Rhilex,
	device *typex.Device,
	serialPort serial.Port,
) typex.XExternalDriver {
	return &rawUartDriver{
		RuleEngine: e,
		ctx:        ctx,
		serialPort: serialPort,
		device:     device,
		locker:     sync.Mutex{},
	}
}

func (a *rawUartDriver) Init(map[string]string) error {
	a.state = typex.DRIVER_UP

	return nil
}

func (a *rawUartDriver) Work() error {
	return nil
}
func (a *rawUartDriver) State() typex.DriverState {
	return a.state
}
func (a *rawUartDriver) Stop() error {
	a.state = typex.DRIVER_STOP
	return a.serialPort.Close()
}

func (a *rawUartDriver) Test() error {
	if a.serialPort == nil {
		return errors.New("serialPort is nil")
	}
	a.locker.Lock()
	_, err := a.serialPort.Write([]byte("\r\n"))
	a.locker.Unlock()
	return err
}

func (a *rawUartDriver) Read(cmd []byte, b []byte) (int, error) {
	a.locker.Lock()
	n, e := a.serialPort.Read(b)
	a.locker.Unlock()
	return n, e
}

func (a *rawUartDriver) Write(cmd []byte, b []byte) (int, error) {
	return a.serialPort.Write(b)
}
func (a *rawUartDriver) DriverDetail() typex.DriverDetail {
	return typex.DriverDetail{
		Name:        "Raw Uart Driver",
		Type:        "RAW_UART",
		Description: "Raw Uart Driver",
	}
}
