package typex

import (
	"context"
	"sync"
)

// Global config
type ExtLib struct {
	Value []string `ini:"value,,allowshadow"`
}
type Secret struct {
	Value []string `ini:"value,,allowshadow"`
}
type RhilexConfig struct {
	IniPath               string
	AppId                 string   `ini:"app_id" json:"appId"`
	MaxQueueSize          int      `ini:"max_queue_size" json:"maxQueueSize"`
	SourceRestartInterval int      `ini:"resource_restart_interval" json:"sourceRestartInterval"`
	GomaxProcs            int      `ini:"gomax_procs" json:"gomaxProcs"`
	EnablePProf           bool     `ini:"enable_pprof" json:"enablePProf"`
	EnableConsole         bool     `ini:"enable_console" json:"enableConsole"`
	AppDebugMode          bool     `ini:"app_debug_mode" json:"appDebugMode"`
	LogLevel              string   `ini:"log_level" json:"logLevel"`
	LogPath               string   `ini:"log_path" json:"logPath"`
	LogMaxSize            int      `ini:"log_max_size" json:"logMaxSize"`
	LogMaxBackups         int      `ini:"log_max_backups" json:"logMaxBackups"`
	LogMaxAge             int      `ini:"log_max_age" json:"logMaxAge"`
	LogCompress           bool     `ini:"log_compress" json:"logCompress"`
	MaxKvStoreSize        int      `ini:"max_kv_store_size" json:"maxKvStoreSize"`
	MaxLostCacheSize      int      `ini:"max_lost_cache_size" json:"maxLostCacheSize"`
	ExtLibs               []string `ini:"ext_libs,,allowshadow" json:"extLibs"`
	DataSchemaSecret      []string `ini:"dataschema_secrets,,allowshadow" json:"dataSchemaSecret"`
}

// Rhilex interface
type Rhilex interface {
	//
	// 启动规则引擎
	//
	Start() *RhilexConfig

	//
	// 执行任务
	//
	WorkInEnd(*InEnd, string) (bool, error)
	WorkDevice(*Device, string) (bool, error)
	//
	// 获取配置
	//
	GetConfig() *RhilexConfig
	//
	// 加载输入
	//
	LoadInEndWithCtx(in *InEnd, ctx context.Context, cancelCTX context.CancelFunc) error
	//
	// 获取输入
	//
	GetInEnd(string) *InEnd
	//
	// 保存输入
	//
	SaveInEnd(*InEnd)
	//
	// 删除输入
	//
	RemoveInEnd(string)
	//
	// 所有输入列表
	//
	AllInEnds() *sync.Map
	//
	// 加载输出
	//
	LoadOutEndWithCtx(in *OutEnd, ctx context.Context, cancelCTX context.CancelFunc) error
	//
	// 所有输出
	//
	AllOutEnds() *sync.Map
	//
	// 获取输出
	//
	GetOutEnd(string) *OutEnd
	//
	// 保存输出
	//
	SaveOutEnd(*OutEnd)
	//
	// 删除输出
	//
	RemoveOutEnd(string)
	//
	// 加载插件
	//
	LoadPlugin(string, XPlugin) error
	//
	// 所有插件列表
	//
	AllPlugins() *sync.Map
	//
	// 加载规则
	//
	LoadRule(*Rule) error
	//
	// 所有规则列表
	//
	AllRules() *sync.Map
	//
	// 获取规则
	//
	GetRule(id string) *Rule
	//
	// 删除规则
	//
	RemoveRule(uuid string)
	//
	// 运行 lua 回调
	//
	RunSourceCallbacks(*InEnd, string)
	RunDeviceCallbacks(*Device, string)
	//
	// 获取版本
	//
	Version() VersionInfo

	//
	// 停止规则引擎
	//
	Stop()
	//
	// Snapshot Dump
	//
	SnapshotDump() string
	//
	// 加载设备
	//
	LoadDeviceWithCtx(in *Device, ctx context.Context, cancelCTX context.CancelFunc) error
	//
	// 获取设备
	//
	GetDevice(string) *Device
	//
	// 保存设备
	//
	SaveDevice(*Device)
	//
	//
	//
	AllDevices() *sync.Map
	//
	// 删除设备
	//
	RemoveDevice(string)
	//
	// 重启源
	//
	RestartInEnd(uuid string) error
	//
	// 重启目标
	//
	RestartOutEnd(uuid string) error
	//
	// 重启设备
	//
	RestartDevice(uuid string) error
	//
	SetDeviceStatus(uuid string, s DeviceState)
	//
	SetSourceStatus(uuid string, s SourceState)
}
