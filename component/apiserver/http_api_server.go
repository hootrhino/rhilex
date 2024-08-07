package httpserver

import (
	"fmt"
	"strconv"
	"time"

	cron_task "github.com/hootrhino/rhilex/component/crontask"
	"github.com/hootrhino/rhilex/component/hwportmanager"
	"github.com/hootrhino/rhilex/component/internotify"
	"github.com/shirou/gopsutil/cpu"

	"github.com/hootrhino/rhilex/component/apiserver/apis"
	"github.com/hootrhino/rhilex/component/apiserver/apis/shelly"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/component/apiserver/server"
	"github.com/hootrhino/rhilex/component/apiserver/service"
	"github.com/hootrhino/rhilex/component/appstack"
	"github.com/hootrhino/rhilex/component/interdb"
	"github.com/hootrhino/rhilex/component/trailer"

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
	/*
	*
	* 加载Port
	*
	 */
	loadAllPortConfig()
	//
	// Load inend from sqlite
	//
	for _, minEnd := range service.AllMInEnd() {
		if err := server.LoadNewestInEnd(minEnd.UUID, engine); err != nil {
			glogger.GLogger.Error("InEnd load failed:", err)
		}
	}

	//
	// Load out from sqlite
	//
	for _, mOutEnd := range service.AllMOutEnd() {
		if err := server.LoadNewestOutEnd(mOutEnd.UUID, engine); err != nil {
			glogger.GLogger.Error("OutEnd load failed:", err)
		}
	}
	// 加载设备
	for _, mDevice := range service.AllDevices() {
		glogger.GLogger.Debug("LoadNewestDevice mDevice.BindRules: ", mDevice.BindRules.String())
		if err := server.LoadNewestDevice(mDevice.UUID, engine); err != nil {
			glogger.GLogger.Error("Device load failed:", err)
		}

	}
	// 加载外挂
	for _, mGoods := range service.AllGoods() {
		newGoods := trailer.GoodsInfo{
			UUID:        mGoods.UUID,
			AutoStart:   mGoods.AutoStart,
			LocalPath:   mGoods.LocalPath,
			NetAddr:     mGoods.NetAddr,
			Args:        mGoods.Args,
			ExecuteType: mGoods.ExecuteType,
			Description: mGoods.Description,
		}
		if err := trailer.StartProcess(newGoods); err != nil {
			glogger.GLogger.Error("Goods load failed:", err)
		}
	}
	//
	// APP stack
	//
	for _, mApp := range service.AllApp() {
		app := appstack.NewApplication(
			mApp.UUID,
			mApp.Name,
			mApp.Version,
		)
		if err := appstack.LoadApp(app, mApp.LuaSource); err != nil {
			glogger.GLogger.Error(err)
			continue
		}
		if *mApp.AutoStart {
			glogger.GLogger.Debug("App autoStart allowed:", app.UUID)
			if err1 := appstack.StartApp(app.UUID); err1 != nil {
				glogger.GLogger.Error("App autoStart failed:", err1)
			}
		}
	}
	//
	// load Cron Task
	for _, task := range service.AllEnabledCronTask() {
		if err := cron_task.GetCronManager().AddTask(task); err != nil {
			glogger.GLogger.Error(err)
			continue
		}
	}

}

/*
*
* 从数据库拿端口配置
*
 */
func loadAllPortConfig() {
	MHwPorts, err := service.AllHwPort()
	if err != nil {
		glogger.GLogger.Fatal(err)
		return
	}
	for _, MHwPort := range MHwPorts {
		Port := hwportmanager.SystemHwPort{
			UUID:        MHwPort.UUID,
			Name:        MHwPort.Name,
			Type:        MHwPort.Type,
			Alias:       MHwPort.Alias,
			Description: MHwPort.Description,
		}
		// 串口
		if MHwPort.Type == "UART" {
			config := hwportmanager.UartConfig{}
			if err := utils.BindSourceConfig(MHwPort.GetConfig(), &config); err != nil {
				glogger.GLogger.Error(err) // 这里必须不能出错
				continue
			}
			Port.Config = config
			hwportmanager.SetHwPort(Port)
		}
		// 未知接口参数为空，以后扩展，比如FD
		if MHwPort.Type != "UART" {
			Port.Config = "NULL"
			hwportmanager.SetHwPort(Port)
		}
	}
}

func (hs *ApiServerPlugin) Init(config *ini.Section) error {
	if err := utils.InIMapToStruct(config, &hs.mainConfig); err != nil {
		return err
	}
	server.StartRhilexApiServer(hs.ruleEngine, hs.mainConfig.Port)

	interdb.DB().Exec("VACUUM;")
	interdb.RegisterModel(
		&model.MInEnd{},
		&model.MOutEnd{},
		&model.MRule{},
		&model.MUser{},
		&model.MDevice{},
		&model.MGoods{},
		&model.MApp{},
		&model.MGenericGroup{},
		&model.MGenericGroupRelation{},
		&model.MNetworkConfig{},
		&model.MWifiConfig{},
		&model.MIotSchema{},
		&model.MIotProperty{},
		&model.MIpRoute{},
		&model.MCronTask{},
		&model.MCronResult{},
		&model.MHwPort{},
		&model.MInternalNotify{},
		&model.MUserLuaTemplate{},
		&model.MModbusDataPoint{},
		&model.MSiemensDataPoint{},
		&model.MHnc8DataPoint{},
		&model.MKnd8DataPoint{},
		&model.MSnmpOid{},
		&model.MBacnetDataPoint{},
		&model.MBacnetRouterDataPoint{},
		&model.MDataPoint{},
	)
	// 初始化所有预制参数
	server.DefaultApiServer.InitializeGenericOSData()
	server.DefaultApiServer.InitializeRHILEXG1Data()
	server.DefaultApiServer.InitializeWindowsData()
	server.DefaultApiServer.InitializeUnixData()
	server.DefaultApiServer.InitializeConfigCtl()
	initRhilex(hs.ruleEngine)
	return nil
}

/*
*
* 加载路由
*
 */
func (hs *ApiServerPlugin) LoadRoute() {
	systemApi := server.RouteGroup(server.ContextUrl("/"))
	{
		systemApi.GET(("/ping"), server.AddRoute(apis.Ping))
	}

	//
	//
	//
	server.DefaultApiServer.Route().GET(server.ContextUrl("drivers"), server.AddRoute(apis.Drivers))

	//
	// Get statistics data
	//
	server.DefaultApiServer.Route().GET(server.ContextUrl("statistics"), server.AddRoute(apis.Statistics))
	//
	// Auth
	//
	userApi := server.RouteGroup(server.ContextUrl("/users"))
	{
		// userApi.GET(("/"), server.AddRoute(apis.Users))
		userApi.POST(("/"), server.AddRoute(apis.CreateUser))
		userApi.PUT(("/update"), server.AddRoute(apis.UpdateUser))
		userApi.GET(("/detail"), server.AddRoute(apis.UserDetail))
		userApi.POST(("/logout"), server.AddRoute(apis.LogOut))

	}

	//
	//
	//
	server.DefaultApiServer.Route().POST(server.ContextUrl("login"), server.AddRoute(apis.Login))
	//
	//
	//
	server.DefaultApiServer.Route().GET(server.ContextUrl("info"), server.AddRoute(apis.Info))
	//
	InEndApi := server.RouteGroup(server.ContextUrl("/inends"))
	{
		InEndApi.GET(("/detail"), server.AddRoute(apis.InEndDetail))
		InEndApi.GET(("/list"), server.AddRoute(apis.InEnds))
		InEndApi.POST(("/create"), server.AddRoute(apis.CreateInend))
		InEndApi.DELETE(("/del"), server.AddRoute(apis.DeleteInEnd))
		InEndApi.PUT(("/update"), server.AddRoute(apis.UpdateInend))
		InEndApi.PUT("/restart", server.AddRoute(apis.RestartInEnd))
		InEndApi.GET("/clients", server.AddRoute(apis.GetInEndClients))
	}

	rulesApi := server.RouteGroup(server.ContextUrl("/rules"))
	{
		rulesApi.POST(("/create"), server.AddRoute(apis.CreateRule))
		rulesApi.PUT(("/update"), server.AddRoute(apis.UpdateRule))
		rulesApi.DELETE(("/del"), server.AddRoute(apis.DeleteRule))
		rulesApi.GET(("/list"), server.AddRoute(apis.Rules))
		rulesApi.GET(("/detail"), server.AddRoute(apis.RuleDetail))
		//
		rulesApi.POST(("/testIn"), server.AddRoute(apis.TestSourceCallback))
		rulesApi.POST(("/testOut"), server.AddRoute(apis.TestOutEndCallback))
		rulesApi.POST(("/testDevice"), server.AddRoute(apis.TestDeviceCallback))
		rulesApi.GET(("/byInend"), server.AddRoute(apis.ListByInend))
		rulesApi.GET(("/byDevice"), server.AddRoute(apis.ListByDevice))
		//
		rulesApi.GET(("/getCanUsedResources"), server.AddRoute(apis.GetAllResources))
		//
		rulesApi.POST(("/formatLua"), server.AddRoute(apis.FormatLua))

	}
	OutEndApi := server.RouteGroup(server.ContextUrl("/outends"))
	{
		OutEndApi.GET(("/detail"), server.AddRoute(apis.OutEndDetail))
		OutEndApi.GET(("/list"), server.AddRoute(apis.OutEnds))
		OutEndApi.POST(("/create"), server.AddRoute(apis.CreateOutEnd))
		OutEndApi.DELETE(("/del"), server.AddRoute(apis.DeleteOutEnd))
		OutEndApi.PUT(("/update"), server.AddRoute(apis.UpdateOutEnd))
		OutEndApi.PUT("/restart", server.AddRoute(apis.RestartOutEnd))
	}

	//
	// 验证 lua 语法
	//
	server.DefaultApiServer.Route().POST(server.ContextUrl("validateRule"), server.AddRoute(apis.ValidateLuaSyntax))

	//
	// 网络适配器列表
	//
	osApi := server.RouteGroup(server.ContextUrl("/os"))
	{
		osApi.GET(("/netInterfaces"), server.AddRoute(apis.GetNetInterfaces))
		osApi.GET(("/osRelease"), server.AddRoute(apis.CatOsRelease))
		osApi.GET(("/system"), server.AddRoute(apis.System))
		osApi.GET(("/startedAt"), server.AddRoute(apis.StartedAt))
		osApi.GET(("/getVideos"), server.AddRoute(apis.GetVideos))
		osApi.GET(("/getGpuInfo"), server.AddRoute(apis.GetGpuInfo))
		osApi.POST(("/resetInterMetric"), server.AddRoute(apis.ResetInterMetric))
	}
	backupApi := server.RouteGroup(server.ContextUrl("/backup"))
	{
		backupApi.GET(("/download"), server.AddRoute(apis.DownloadSqlite))
		backupApi.POST(("/upload"), server.AddRoute(apis.UploadSqlite))
		backupApi.GET(("/snapshot"), server.AddRoute(apis.SnapshotDump))
		backupApi.GET(("/runningLog"), server.AddRoute(apis.GetRunningLog))
	}
	//
	// 设备管理
	//
	deviceApi := server.RouteGroup(server.ContextUrl("/devices"))
	{
		deviceApi.POST(("/create"), server.AddRoute(apis.CreateDevice))
		deviceApi.PUT(("/update"), server.AddRoute(apis.UpdateDevice))
		deviceApi.DELETE(("/del"), server.AddRoute(apis.DeleteDevice))
		deviceApi.GET(("/detail"), server.AddRoute(apis.DeviceDetail))
		deviceApi.GET("/group", server.AddRoute(apis.ListDeviceGroup))
		deviceApi.GET("/listByGroup", server.AddRoute(apis.ListDeviceByGroup))
		deviceApi.GET("/list", server.AddRoute(apis.ListDevice))
		deviceApi.PUT("/restart", server.AddRoute(apis.RestartDevice))
		deviceApi.GET("/deviceErrMsg", server.AddRoute(apis.GetDeviceErrorMsg))
	}
	// Modbus 点位表
	modbusMasterApi := server.RouteGroup(server.ContextUrl("/modbus_master_sheet"))
	{
		modbusMasterApi.POST(("/sheetImport"), server.AddRoute(apis.ModbusMasterSheetImport))
		modbusMasterApi.GET(("/sheetExport"), server.AddRoute(apis.ModbusMasterPointsExport))
		modbusMasterApi.GET(("/list"), server.AddRoute(apis.ModbusMasterSheetPageList))
		modbusMasterApi.POST(("/update"), server.AddRoute(apis.ModbusMasterSheetUpdate))
		modbusMasterApi.DELETE(("/delIds"), server.AddRoute(apis.ModbusMasterSheetDelete))
		modbusMasterApi.DELETE(("/delAll"), server.AddRoute(apis.ModbusMasterSheetDeleteAll))
	}
	modbusApi := server.RouteGroup(server.ContextUrl("/modbus_slaver_sheet"))
	{
		modbusApi.GET(("/list"), server.AddRoute(apis.ModbusSlaverSheetPageList))
	}
	// S1200 点位表
	SIEMENS_PLC := server.RouteGroup(server.ContextUrl("/s1200_data_sheet"))
	{
		SIEMENS_PLC.POST(("/sheetImport"), server.AddRoute(apis.SiemensSheetImport))
		SIEMENS_PLC.GET(("/sheetExport"), server.AddRoute(apis.SiemensPointsExport))
		SIEMENS_PLC.GET(("/list"), server.AddRoute(apis.SiemensSheetPageList))
		SIEMENS_PLC.POST(("/update"), server.AddRoute(apis.SiemensSheetUpdate))
		SIEMENS_PLC.DELETE(("/delIds"), server.AddRoute(apis.SiemensSheetDelete))
		SIEMENS_PLC.DELETE(("/delAll"), server.AddRoute(apis.SiemensSheetDeleteAll))
	}
	// 华中数控 点位表
	Hnc8 := server.RouteGroup(server.ContextUrl("/hnc8_data_sheet"))
	{
		Hnc8.POST(("/sheetImport"), server.AddRoute(apis.Hnc8SheetImport))
		Hnc8.GET(("/sheetExport"), server.AddRoute(apis.Hnc8PointsExport))
		Hnc8.GET(("/list"), server.AddRoute(apis.Hnc8SheetPageList))
		Hnc8.POST(("/update"), server.AddRoute(apis.Hnc8SheetUpdate))
		Hnc8.DELETE(("/delIds"), server.AddRoute(apis.Hnc8SheetDelete))
		Hnc8.DELETE(("/delAll"), server.AddRoute(apis.Hnc8SheetDeleteAll))
	}

	// ----------------------------------------------------------------------------------------------
	// APP
	// ----------------------------------------------------------------------------------------------
	appApi := server.RouteGroup(server.ContextUrl("/app"))
	{
		appApi.GET(("/list"), server.AddRoute(apis.Apps))
		appApi.POST(("/create"), server.AddRoute(apis.CreateApp))
		appApi.PUT(("/update"), server.AddRoute(apis.UpdateApp))
		appApi.DELETE(("/del"), server.AddRoute(apis.RemoveApp))
		appApi.PUT(("/start"), server.AddRoute(apis.StartApp))
		appApi.PUT(("/stop"), server.AddRoute(apis.StopApp))
		appApi.GET(("/detail"), server.AddRoute(apis.AppDetail))
	}
	// ----------------------------------------------------------------------------------------------
	// Plugin
	// ----------------------------------------------------------------------------------------------
	pluginsApi := server.RouteGroup(server.ContextUrl("/plugware"))
	{
		pluginsApi.GET(("/list"), server.AddRoute(apis.Plugins))
		pluginsApi.POST(("/service"), server.AddRoute(apis.PluginService))
		pluginsApi.GET(("/detail"), server.AddRoute(apis.PluginDetail))
	}

	//
	// 分组管理
	//
	groupApi := server.RouteGroup(server.ContextUrl("/group"))
	{
		groupApi.POST("/create", server.AddRoute(apis.CreateGroup))
		groupApi.PUT("/update", server.AddRoute(apis.UpdateGroup))
		groupApi.GET("/list", server.AddRoute(apis.ListGroup))
		groupApi.GET("/detail", server.AddRoute(apis.GroupDetail))
		groupApi.POST("/bind", server.AddRoute(apis.BindResource))
		groupApi.PUT("/unbind", server.AddRoute(apis.UnBindResource))
		groupApi.DELETE("/del", server.AddRoute(apis.DeleteGroup))
	}
	//
	// 用户LUA代码段管理
	//
	userLuaApi := server.RouteGroup(server.ContextUrl("/userlua"))
	{
		userLuaApi.POST("/create", server.AddRoute(apis.CreateUserLuaTemplate))
		userLuaApi.PUT("/update", server.AddRoute(apis.UpdateUserLuaTemplate))
		userLuaApi.GET("/listByGroup", server.AddRoute(apis.ListUserLuaTemplateByGroup))
		userLuaApi.GET("/detail", server.AddRoute(apis.UserLuaTemplateDetail))
		userLuaApi.GET("/group", server.AddRoute(apis.ListUserLuaTemplateGroup))
		userLuaApi.DELETE("/del", server.AddRoute(apis.DeleteUserLuaTemplate))
		userLuaApi.GET("/search", server.AddRoute(apis.SearchUserLuaTemplateGroup))
	}

	trailerApi := server.RouteGroup(server.ContextUrl("/goods"))
	{
		trailerApi.GET("/list", server.AddRoute(apis.GoodsList))
		trailerApi.GET(("/detail"), server.AddRoute(apis.GoodsDetail))
		trailerApi.POST("/create", server.AddRoute(apis.CreateGoods))
		trailerApi.PUT("/update", server.AddRoute(apis.UpdateGoods))
		trailerApi.PUT("/cleanGarbage", server.AddRoute(apis.CleanGoodsUpload))
		trailerApi.PUT("/start", server.AddRoute(apis.StartGoods))
		trailerApi.PUT("/stop", server.AddRoute(apis.StopGoods))
		trailerApi.DELETE("/", server.AddRoute(apis.DeleteGoods))
	}
	// 硬件接口API
	HwIFaceApi := server.DefaultApiServer.GetGroup(server.ContextUrl("/hwiface"))
	{
		HwIFaceApi.GET("/detail", server.AddRoute(apis.GetHwPortDetail))
		HwIFaceApi.GET("/list", server.AddRoute(apis.AllHwPorts))
		HwIFaceApi.POST("/update", server.AddRoute(apis.UpdateHwPortConfig))
		HwIFaceApi.GET("/refresh", server.AddRoute(apis.RefreshPortList))
	}
	// 站内公告
	internalNotifyApi := server.DefaultApiServer.GetGroup(server.ContextUrl("/notify"))
	{
		internalNotifyApi.PUT("/clear", server.AddRoute(apis.ClearInternalNotifies))
		internalNotifyApi.PUT("/read", server.AddRoute(apis.ReadInternalNotifies))
		internalNotifyApi.GET("/pageList", server.AddRoute(apis.PageInternalNotifies))
	}
	//
	// 系统设置
	//
	apis.LoadSystemSettingsAPI()

	/**
	 * 定时任务
	 */
	crontaskApi := server.DefaultApiServer.GetGroup(server.ContextUrl("/crontask"))
	{
		crontaskApi.POST("/create", server.AddRouteV2(apis.CreateCronTask))
		crontaskApi.DELETE("/del", server.AddRouteV2(apis.DeleteCronTask))
		crontaskApi.PUT("/update", server.AddRouteV2(apis.UpdateCronTask))
		crontaskApi.GET("/list", server.AddRouteV2(apis.ListCronTask))
		crontaskApi.GET("/results/page", server.AddRouteV2(apis.PageCronTaskResult))
		crontaskApi.GET("/start", server.AddRouteV2(apis.StartTask))
		crontaskApi.GET("/stop", server.AddRouteV2(apis.StopTask))
	}
	//
	// jpegStream APi
	//
	jpegStream := server.DefaultApiServer.GetGroup(server.ContextUrl("/jpeg_stream"))
	{
		jpegStream.GET("/list", server.AddRoute(apis.GetJpegStreamList))
		jpegStream.GET("/detail", server.AddRoute(apis.GetJpegStreamDetail))
	}
	//
	//New Api
	//
	//Shelly
	shelly.InitShellyRoute()
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
	// ata Point Route
	apis.InitDataPointRoute()
	// Mqtt Server
	apis.InitMqttSourceServerRoute()
}

// ApiServerPlugin Start
func (hs *ApiServerPlugin) Start(r typex.Rhilex) error {
	hs.ruleEngine = r
	hs.LoadRoute()
	glogger.GLogger.Infof("Http server started on :%v", hs.mainConfig.Port)
	return nil
}

func (hs *ApiServerPlugin) Stop() error {
	return nil
}

func (hs *ApiServerPlugin) PluginMetaInfo() typex.XPluginMetaInfo {
	return typex.XPluginMetaInfo{
		UUID:        hs.uuid,
		Name:        "Api Server",
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
		case <-typex.GCTX.Done():
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
			internotify.Push(internotify.BaseEvent{
				Type:    `WARNING`,
				Event:   `system.cpu.load`,
				Ts:      uint64(time.Now().UnixMilli()),
				Summary: "High CPU Usage",
				Info:    fmt.Sprintf("High CPU Usage: %.2f%%, please maintain the device", V),
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
