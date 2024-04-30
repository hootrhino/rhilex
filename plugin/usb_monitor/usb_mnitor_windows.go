package usbmonitor

import (
	"errors"

	"github.com/hootrhino/rhilex/typex"

	"gopkg.in/ini.v1"
)

/*
*
* USB 热插拔监控器, 方便观察USB状态, 本插件只支持Linux！！！
*
 */
type usbMonitor struct {
	uuid string
}

func NewUsbMonitor() typex.XPlugin {
	return &usbMonitor{
		uuid: "USB-MONITOR",
	}
}
func (usbm *usbMonitor) Init(_ *ini.Section) error {
	return nil

}

func (usbm *usbMonitor) Start(_ typex.Rhilex) error {
	return errors.New("USB monitor plugin not support windows")
}

func (usbm *usbMonitor) Stop() error {
	return nil
}

func (usbm *usbMonitor) PluginMetaInfo() typex.XPluginMetaInfo {
	return typex.XPluginMetaInfo{
		UUID:     usbm.uuid,
		Name:     "USB Monitor",
		Version:  "v0.0.1",
		Homepage: "https://github.com/hootrhino/rhilex.git",
		HelpLink: "https://github.com/hootrhino/rhilex.git",
		Author:   "RHILEXTeam",
		Email:    "RHILEXTeam@hootrhino.com",
		License:  "",
	}
}

/*
*
* 服务调用接口
*
 */
func (usbm *usbMonitor) Service(arg typex.ServiceArg) typex.ServiceResult {
	return typex.ServiceResult{}
}
