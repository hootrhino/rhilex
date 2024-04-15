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
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)

type ShellyGen1ProxyGateway struct {
	typex.XStatus
	status     typex.DeviceState
	mainConfig ShellyGen1ProxyGatewayConfig
}
type ShellyGen1ProxyGatewayConfig struct {
}

func NewShellyGen1ProxyGateway(e typex.RuleX) typex.XDevice {
	Shelly := new(ShellyGen1ProxyGateway)
	Shelly.mainConfig = ShellyGen1ProxyGatewayConfig{}
	Shelly.RuleEngine = e
	return Shelly
}

//  初始化
func (Shelly *ShellyGen1ProxyGateway) Init(devId string, configMap map[string]interface{}) error {
	Shelly.PointId = devId
	if err := utils.BindSourceConfig(configMap, &Shelly.mainConfig); err != nil {
		glogger.GLogger.Error(err)
		return err
	}

	return nil
}

// 启动
func (Shelly *ShellyGen1ProxyGateway) Start(cctx typex.CCTX) error {
	Shelly.Ctx = cctx.Ctx
	Shelly.CancelCTX = cctx.CancelCTX

	Shelly.status = typex.DEV_UP
	return nil
}

func (Shelly *ShellyGen1ProxyGateway) OnRead(cmd []byte, data []byte) (int, error) {

	return 0, nil
}

// 把数据写入设备
func (Shelly *ShellyGen1ProxyGateway) OnWrite(cmd []byte, b []byte) (int, error) {
	return 0, nil
}

// 设备当前状态
func (Shelly *ShellyGen1ProxyGateway) Status() typex.DeviceState {
	return typex.DEV_UP
}

// 停止设备
func (Shelly *ShellyGen1ProxyGateway) Stop() {
	Shelly.status = typex.DEV_DOWN
	Shelly.CancelCTX()
}

// 真实设备
func (Shelly *ShellyGen1ProxyGateway) Details() *typex.Device {
	return Shelly.RuleEngine.GetDevice(Shelly.PointId)
}

// 状态
func (Shelly *ShellyGen1ProxyGateway) SetState(status typex.DeviceState) {
	Shelly.status = status

}

// 驱动
func (Shelly *ShellyGen1ProxyGateway) Driver() typex.XExternalDriver {
	return nil
}

// --------------------------------------------------------------------------------------------------
//
// --------------------------------------------------------------------------------------------------

func (Shelly *ShellyGen1ProxyGateway) OnDCACall(UUID string, Command string, Args interface{}) typex.DCAResult {
	return typex.DCAResult{}
}
func (Shelly *ShellyGen1ProxyGateway) OnCtrl(cmd []byte, args []byte) ([]byte, error) {
	return []byte{}, nil
}
