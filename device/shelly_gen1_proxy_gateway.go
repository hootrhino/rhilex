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
	"sync"
	"time"

	"github.com/hootrhino/rhilex/component/shellymanager"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
	"github.com/hootrhino/rhilex/utils/tinyarp"
)

type ShellyConfig struct {
	CommonConfig ShellyCommonConfig `json:"commonConfig" validate:"required"`
}
type ShellyGen1ProxyGateway struct {
	typex.XStatus
	status              typex.DeviceState
	mainConfig          ShellyConfig
	BlackList           map[string]string
	locker              sync.Mutex
	shellyWebHookServer *shellymanager.ShellyWebHookServer
}

/*
*
* 配置
*
 */
type ShellyCommonConfig struct {
	// CIDR
	NetworkCidr string `json:"networkCidr" validate:"required"`
	// AutoScan
	AutoScan *bool `json:"autoScan" validate:"required"`
	// 扫描超时
	ScanTimeout int `json:"timeout" validate:"required"`
	// Request Frequency, default 5 second
	Frequency int64 `json:"frequency" validate:"required"`
	// WebHook Listen Port
	WebHookPort int `json:"webHookPort" validate:"required"`
}

/*
*
* 初始化
*
 */
func NewShellyGen1ProxyGateway(e typex.Rhilex) typex.XDevice {
	Shelly := new(ShellyGen1ProxyGateway)
	Shelly.BlackList = map[string]string{}
	Shelly.locker = sync.Mutex{}
	Shelly.mainConfig = ShellyConfig{
		CommonConfig: ShellyCommonConfig{
			NetworkCidr: "192.168.1.0/24",
			AutoScan: func() *bool {
				b := true
				return &b
			}(),
			ScanTimeout: 3000, //ms
			Frequency:   5000, //ms
			WebHookPort: 6400,
		},
	}
	Shelly.RuleEngine = e

	return Shelly
}

//  初始化
func (Shelly *ShellyGen1ProxyGateway) Init(devId string, configMap map[string]interface{}) error {
	Shelly.PointId = devId
	if err := utils.BindSourceConfig(configMap, &Shelly.mainConfig.CommonConfig); err != nil {
		glogger.GLogger.Error(err)
		return err
	}
	shellymanager.RegisterSlot(devId)
	Shelly.shellyWebHookServer = shellymanager.NewShellyWebHookServer(
		Shelly.RuleEngine,
		Shelly.mainConfig.CommonConfig.WebHookPort,
	)
	return nil
}

// 启动
func (Shelly *ShellyGen1ProxyGateway) Start(cctx typex.CCTX) error {
	Shelly.Ctx = cctx.Ctx
	Shelly.CancelCTX = cctx.CancelCTX
	Shelly.shellyWebHookServer.SetEventCallBack(func(Notify shellymanager.ShellyDeviceNotify) {
		glogger.GLogger.Debug(Notify)
		Shelly.RuleEngine.WorkDevice(Shelly.Details(), Notify.String())
	})
	Shelly.shellyWebHookServer.StartServer(Shelly.Ctx)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		if *Shelly.mainConfig.CommonConfig.AutoScan {
			for {
				select {
				case <-Shelly.Ctx.Done():
					return
				case <-ticker.C:
					glogger.GLogger.Debug("Clear BlackList")
					// 黑名单30秒刷新一次
					Shelly.locker.Lock()
					for k := range Shelly.BlackList {
						delete(Shelly.BlackList, k)
					}
					Shelly.locker.Unlock()
				default:
				}
				glogger.GLogger.Debug("ScanDevice")
				Shelly.ScanDevice(Shelly.PointId)
				time.Sleep(time.Duration(Shelly.mainConfig.CommonConfig.Frequency) * time.Millisecond)
				<-ticker.C
			}
		}
	}()
	Shelly.status = typex.DEV_UP
	return nil
}

// 停止设备
func (Shelly *ShellyGen1ProxyGateway) Stop() {
	Shelly.status = typex.DEV_DOWN
	Shelly.CancelCTX()
	shellymanager.UnRegisterSlot(Shelly.PointId)
	Shelly.shellyWebHookServer.Stop()
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

// 真实设备
func (Shelly *ShellyGen1ProxyGateway) Details() *typex.Device {
	return Shelly.RuleEngine.GetDevice(Shelly.PointId)
}

// 状态
func (Shelly *ShellyGen1ProxyGateway) SetState(status typex.DeviceState) {
	Shelly.status = status

}

func (Shelly *ShellyGen1ProxyGateway) OnDCACall(UUID string, Command string, Args interface{}) typex.DCAResult {
	return typex.DCAResult{}
}
func (Shelly *ShellyGen1ProxyGateway) OnCtrl(cmd []byte, args []byte) ([]byte, error) {
	return []byte{}, nil
}

// --------------------------------------------------------------------------------------------------
// Shelly API
// --------------------------------------------------------------------------------------------------

func (Shelly *ShellyGen1ProxyGateway) ScanDevice(Slot string) {
	IPs, err := shellymanager.ScanCIDR(Shelly.mainConfig.CommonConfig.NetworkCidr, 5*time.Second)
	if err != nil {
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(len(IPs))
	// 1 将第一次扫出来请求失败的设备拉进黑名单,防止浪费资源
	// 2 已经有在列表里面的就不再扫描
	for _, Ip := range IPs {
		glogger.GLogger.Debugf("Scan Device [%s]", Ip)
		if AlreadyExistsMac, ok := Shelly.BlackList[Ip]; ok {
			if AlreadyExistsMac == Ip {
				continue
			}
		}
		if shellymanager.Exists(Slot, Ip) {
			continue
		}
		go func(Ip string) {
			defer wg.Done()
			if tinyarp.IsValidIP(Ip) {
				DeviceInfo, err := shellymanager.GetShellyDeviceInfo(Ip)
				if err != nil {
					Shelly.locker.Lock()
					Shelly.BlackList[Ip] = Ip
					Shelly.locker.Unlock()
					glogger.GLogger.Error(err)
					return
				}
				// 注册设备到Registry
				DeviceInfo.Ip = Ip
				if DeviceInfo.Name == nil {
					DName := "UNKNOWN"
					DeviceInfo.Name = &DName
				}
				if utils.IsValidMacAddress1(DeviceInfo.Mac) ||
					utils.IsValidMacAddress2(DeviceInfo.Mac) {
					shellymanager.SetValue(Slot, DeviceInfo.Mac, shellymanager.ShellyDevice{
						Ip:         DeviceInfo.Ip,
						Name:       DeviceInfo.Name,
						ID:         DeviceInfo.ID,
						Mac:        DeviceInfo.Mac,
						Slot:       DeviceInfo.Slot,
						Model:      DeviceInfo.Model,
						Gen:        DeviceInfo.Gen,
						FwID:       DeviceInfo.FwID,
						Ver:        DeviceInfo.Ver,
						App:        DeviceInfo.App,
						AuthEn:     DeviceInfo.AuthEn,
						AuthDomain: DeviceInfo.AuthDomain,
					})
				} else {
					glogger.GLogger.Error("Invalid Mac Address")
				}
			} else {
				Shelly.locker.Lock()
				Shelly.BlackList[Ip] = Ip
				Shelly.locker.Unlock()
			}
		}(Ip)
	}
	wg.Wait()
}
