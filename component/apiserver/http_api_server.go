package httpserver

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hootrhino/rhilex/alarmcenter"
	"github.com/hootrhino/rhilex/component/crontask"
	dataschema "github.com/hootrhino/rhilex/component/dataschema"
	"github.com/hootrhino/rhilex/component/eventbus"
	"github.com/shirou/gopsutil/cpu"

	"github.com/hootrhino/rhilex/applet"
	"github.com/hootrhino/rhilex/component/apiserver/apis"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/component/apiserver/server"
	"github.com/hootrhino/rhilex/component/apiserver/service"
	"github.com/hootrhino/rhilex/component/interdb"

	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"

	"gopkg.in/ini.v1"

	_ "github.com/mattn/go-sqlite3"
)

type _serverConfig struct {
	DbPath string `ini:"dbpath"`
	Port   int    `ini:"port"`
}
type ApiServerPlugin struct {
	uuid       string
	ruleEngine typex.Rhilex
	mainConfig _serverConfig
}

func NewHttpApiServer(ruleEngine typex.Rhilex) *ApiServerPlugin {
	return &ApiServerPlugin{
		uuid:       "HTTP-API-SERVER",
		mainConfig: _serverConfig{Port: 2580},
		ruleEngine: ruleEngine,
	}
}

/*
*
* 初始化RHILEX, 初始化数据到运行时
*
 */
func initRhilex(engine typex.Rhilex) {
	go GetCpuUsage()
	for _, mAlarmRule := range service.AllAlarmRules() {
		ExprDefines := []alarmcenter.ExprDefine{}
		for _, exprDefine := range mAlarmRule.GetExprDefine() {
			ExprDefines = append(ExprDefines, alarmcenter.ExprDefine{
				Expr:      exprDefine.Expr,
				EventType: exprDefine.EventType,
			})
		}
		alarmcenter.LoadAlarmRule(mAlarmRule.UUID, alarmcenter.AlarmRule{
			Interval:    time.Duration(mAlarmRule.Interval) * time.Second,
			Threshold:   mAlarmRule.Threshold,
			HandleId:    mAlarmRule.HandleId,
			ExprDefines: ExprDefines,
		},
		)
	}

	for _, minEnd := range service.AllMInEnd() {
		if err := server.LoadNewestInEnd(minEnd.UUID, engine); err != nil {
			glogger.GLogger.Error("InEnd load failed:", err)
		}
	}
	for _, mOutEnd := range service.AllMOutEnd() {
		if err := server.LoadNewestOutEnd(mOutEnd.UUID, engine); err != nil {
			glogger.GLogger.Error("OutEnd load failed:", err)
		}
	}
	for _, mDevice := range service.AllDevices() {
		glogger.GLogger.Debug("LoadNewestDevice mDevice.BindRules: ", mDevice.BindRules.String())
		if err := server.LoadNewestDevice(mDevice.UUID, engine); err != nil {
			glogger.GLogger.Error("Device load failed:", err)
		}
	}
	//
	// APP stack
	//
	for _, mApp := range service.AllApp() {
		app := applet.NewApplication(
			mApp.UUID,
			mApp.Name,
			mApp.Version,
		)
		if err := applet.LoadApp(app, mApp.LuaSource); err != nil {
			glogger.GLogger.Error(err)
			continue
		}
		if *mApp.AutoStart {
			glogger.GLogger.Debug("App autoStart allowed:", app.UUID)
			if err1 := applet.StartApp(app.UUID); err1 != nil {
				glogger.GLogger.Error("App autoStart failed:", err1)
			}
		}
	}

}

func (hs *ApiServerPlugin) Init(config *ini.Section) error {
	if err := utils.InIMapToStruct(config, &hs.mainConfig); err != nil {
		return err
	}
	server.StartRhilexApiServer(hs.ruleEngine, hs.mainConfig.Port)
	interdb.InterDb().Exec("VACUUM;")
	interdb.InterDbRegisterModel(
		&model.MInEnd{},
		&model.MOutEnd{},
		&model.MRule{},
		&model.MUser{},
		&model.MDevice{},
		&model.MCecolla{},
		&model.MApplet{},
		&model.MMultiMedia{},
		&alarmcenter.MAlarmRule{},
		&model.MGenericGroup{},
		&model.MGenericGroupRelation{},
		&model.MNetworkConfig{},
		&model.MIotSchema{},
		&model.MIotProperty{},
		&model.MIpRoute{},
		&model.MUart{},
		&model.MUserLuaTemplate{},
		&model.MModbusDataPoint{},
		&model.MSiemensDataPoint{},
		&model.MSnmpOid{},
		&model.MCjt1882004DataPoint{},
		&model.MDlt6452007DataPoint{},
		&model.MSzy2062016DataPoint{},
		&model.MUserProtocolDataPoint{},
		&model.MBacnetDataPoint{},
		&model.MBacnetRouterDataPoint{},
		&model.MMBusDataPoint{},
		&model.MCronRebootConfig{},
	)
	// 初始化所有预制参数
	server.DefaultApiServer.InitializeProduct()
	server.DefaultApiServer.InitializeGenericOSData()
	server.DefaultApiServer.InitializeWindowsData()
	server.DefaultApiServer.InitializeUnixData()
	// InitDataSchemaCache
	dataschema.InitDataSchemaCache(hs.ruleEngine)
	// Cron Reboot Executor
	crontask.InitCronRebootExecutor(hs.ruleEngine)
	initRhilex(hs.ruleEngine)
	return nil
}

/*
*
* 加载路由
*
 */
func (hs *ApiServerPlugin) LoadRoute() {
	// User
	apis.InitUserRoute()
	// CE collaboration
	apis.InitCecollaRoute()
	// In End
	apis.InitInEndRoute()
	// Rules
	apis.InitRulesRoute()
	// Out End
	apis.InitOutEndRoute()
	// System API
	apis.InitSystemRoute()
	// backup
	apis.InitBackupRoute()
	// 设备管理
	apis.InitDeviceRoute()
	// Modbus Slaver
	apis.InitModbusSlaverRoute()
	// S1200 点位表
	apis.InitSiemensS7Route()
	// applet
	apis.InitAppletRoute()
	// plugins
	apis.InitPluginsRoute()
	// 分组管理
	apis.InitGroupRoute()
	// 用户LUA代码段管理
	apis.InitUserLuaRoute()
	// System Permission
	apis.InitSysMenuPermissionRoute()
	// System Settings
	apis.LoadSystemSettingsAPI()
	// Modbus
	apis.InitModbusRoute()
	// Mbus
	apis.InitMBusRoute()
	// DLT645
	apis.InitDlt6452007Route()
	// CJT188-2004
	apis.InitCjt1882004Route()
	// Szy206
	apis.InitSzy2062016Route()
	// User Protocol
	apis.InitUserProtocolRoute()
	// Init Internal Notify Route
	apis.InitInternalNotifyRoute()
	// Snmp Route
	apis.InitSnmpRoute()
	// Bacnet Route
	apis.InitBacnetIpRoute()
	// Bacnet Router
	apis.InitBacnetRouterRoute()
	// Data Schema
	apis.InitDataSchemaApi()
	// Data Center
	apis.InitDataCenterApi()
	// Transceiver
	apis.InitTransceiverRoute()
	// Mqtt Server
	apis.InitMqttSourceServerRoute()
	// Cron Reboot
	apis.InitCronRebootRoute()
	// 告警规则
	apis.LoadAlarmRuleRoute()
	// 告警日志
	apis.LoadAlarmLogRoute()
	// 多媒体
	apis.InitMultiMediaRoute()
}

// ApiServerPlugin Start
func (hs *ApiServerPlugin) Start(r typex.Rhilex) error {
	hs.ruleEngine = r
	hs.LoadRoute()
	glogger.GLogger.Infof("Http server started on :%v", hs.mainConfig.Port)
	return nil
}

func (hs *ApiServerPlugin) Stop() error {
	dataschema.FlushDataSchemaCache()
	return nil
}

func (hs *ApiServerPlugin) PluginMetaInfo() typex.XPluginMetaInfo {
	return typex.XPluginMetaInfo{
		UUID:        hs.uuid,
		Name:        "RHILEX HTTP RESTFul Api Server",
		Version:     "v1.0.0",
		Description: "RHILEX HTTP RESTFul Api Server",
	}
}

/*
*
* 服务调用接口
*
 */
func (*ApiServerPlugin) Service(arg typex.ServiceArg) typex.ServiceResult {
	return typex.ServiceResult{Out: "ApiServerPlugin"}
}
func GetCpuUsage() {
	for {
		select {
		case <-context.Background().Done():
			{
				return
			}
		default:
			{
			}
		}
		cpuPercent, _ := cpu.Percent(time.Duration(10)*time.Second, true)
		V := calculateCpuPercent(cpuPercent)
		// TODO 这个比例需要通过参数适配
		if V > 90 {
			eventbus.Publish("system.cpu.load", eventbus.EventMessage{
				Topic:   "system.cpu.load.HTTP-API-SERVER",
				From:    "HTTP-API-SERVER",
				Type:    "SYSTEM",
				Event:   `system.cpu.load`,
				Ts:      uint64(time.Now().UnixMilli()),
				Payload: fmt.Sprintf("High CPU Usage: %.2f%%, please maintain the device", V),
			})
		}
	}

}

// 计算CPU平均使用率
func calculateCpuPercent(cpus []float64) float64 {
	var acc float64 = 0
	for _, v := range cpus {
		acc += v
	}
	value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", acc/float64(len(cpus))), 64)
	return value
}
