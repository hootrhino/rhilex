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

package engine

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	wdog "github.com/hootrhino/rhilex/plugin/generic_watchdog"
	modbusscrc "github.com/hootrhino/rhilex/plugin/modbus_crc_tools"
	modbusscanner "github.com/hootrhino/rhilex/plugin/modbus_scanner"
	mqttserver "github.com/hootrhino/rhilex/plugin/mqtt_server"
	ttyterminal "github.com/hootrhino/rhilex/plugin/ttyd_terminal"
	usbmonitor "github.com/hootrhino/rhilex/plugin/usb_monitor"
	"gopkg.in/ini.v1"

	httpserver "github.com/hootrhino/rhilex/component/rhilex_api_server"
	"github.com/hootrhino/rhilex/core"
	"github.com/hootrhino/rhilex/glogger"
	icmpsender "github.com/hootrhino/rhilex/plugin/icmp_sender"
	license_manager "github.com/hootrhino/rhilex/plugin/license_manager"
	"github.com/hootrhino/rhilex/typex"
)

// 启动 rhilex
func RunRhilex(iniPath string) {
	mainConfig := core.InitGlobalConfig(iniPath)
	//----------------------------------------------------------------------------------------------
	// Init logger
	//----------------------------------------------------------------------------------------------
	glogger.StartGLogger(
		core.GlobalConfig.LogLevel,
		mainConfig.EnableConsole,
		mainConfig.AppDebugMode,
		core.GlobalConfig.LogPath,
		mainConfig.AppId, mainConfig.AppName,
	)
	glogger.StartNewRealTimeLogger(core.GlobalConfig.LogLevel)
	//----------------------------------------------------------------------------------------------
	// Init Component
	//----------------------------------------------------------------------------------------------
	core.StartStore(core.GlobalConfig.MaxQueueSize)
	core.SetDebugMode(mainConfig.EnablePProf)
	core.SetGomaxProcs(mainConfig.GomaxProcs)
	//
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGABRT, syscall.SIGTERM)
	engine := InitRuleEngine(mainConfig)
	engine.Start()
	// Load Http api Server
	httpServer := httpserver.NewHttpApiServer(engine)
	if err := engine.LoadPlugin("plugin.http_server", httpServer); err != nil {
		glogger.GLogger.Error(err)
		return
	}
	license_manager := license_manager.NewLicenseManager(engine)
	if err := engine.LoadPlugin("plugin.license_manager", license_manager); err != nil {
		glogger.GLogger.Error(err)
		return
	}

	// Load Plugin
	loadPlugin(engine)

	s := <-c
	glogger.GLogger.Warn("RHILEX Receive Stop Signal: ", s)
	typex.GCancel()
	engine.Stop()
	os.Exit(0)
}

// loadPlugin 根据Ini配置信息，加载插件
func loadPlugin(engine typex.Rhilex) {
	cfg, _ := ini.ShadowLoad(core.INIPath)
	sections := cfg.ChildSections("plugin")
	for _, section := range sections {
		name := strings.TrimPrefix(section.Name(), "plugin.")
		enable, err := section.GetKey("enable")
		if err != nil {
			continue
		}
		if !enable.MustBool(false) {
			glogger.GLogger.Warnf("Plugin is disable:%s", name)
			continue
		}
		var plugin typex.XPlugin
		if name == "mqtt_server" {
			plugin = mqttserver.NewMqttServer()
		}
		if name == "usbmonitor" {
			plugin = usbmonitor.NewUsbMonitor()
		}
		if name == "icmpsender" {
			plugin = icmpsender.NewICMPSender()
		}
		if name == "modbus_scanner" {
			plugin = modbusscanner.NewModbusScanner()
		}
		if name == "ttyd" {
			plugin = ttyterminal.NewWebTTYPlugin()
		}
		if name == "modbus_crc_tools" {
			plugin = modbusscrc.NewModbusCrcCalculator()
		}
		if name == "soft_wdog" {
			plugin = wdog.NewGenericWatchDog()
		}
		if plugin != nil {
			if err := engine.LoadPlugin(section.Name(), plugin); err != nil {
				glogger.GLogger.Error(err)
			}
		}
	}
}
