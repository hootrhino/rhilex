# 设备接入模块开发指南
## 概述
主要阐述如何快速开发一个设备接入模块。
## 模板
下面是一个设备的基础模板。
```go

package device

import (
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)
// TemplateDevice 是一个设备模板结构体，它包含一个 XStatus 类型，后者可能定义了设备的状态信息。
type TemplateDevice struct {
	typex.XStatus
}

// NewTemplateDevice 是 TemplateDevice 的构造函数，它接受一个 Rhilex 类型的参数 e，
// 并返回一个实现了 XDevice 接口的 TemplateDevice 实例。
func NewTemplateDevice(e typex.Rhilex) typex.XDevice {
	hd := new(TemplateDevice)
	hd.RuleEngine = e // 将 RuleEngine 设置为传入的 Rhilex 实例
	return hd
}

// Init 方法用于初始化 TemplateDevice 实例。它接受设备ID和配置信息映射表作为参数，
// 并将配置信息绑定到 hd 的 mainConfig 字段。如果配置绑定失败，则返回错误。
func (hd *TemplateDevice) Init(devId string, configMap map[string]interface{}) error {
	hd.PointId = devId // 设置设备的 PointId 为传入的 devId
	if err := utils.BindSourceConfig(configMap, &hd.mainConfig); err != nil {
		glogger.GLogger.Error(err) // 如果绑定配置出错，记录错误日志
		return err                 // 返回错误信息
	}
	return nil // 如果配置绑定成功，返回 nil
}

// Start 方法用于启动 TemplateDevice 实例。它接受一个 CCTX 类型的上下文参数，
// 并设置 TemplateDevice 的上下文和取消函数。然后，它将设备状态设置为 DEV_UP。
func (hd *TemplateDevice) Start(cctx typex.CCTX) error {
	hd.Ctx = cctx.Ctx           // 设置上下文
	hd.CancelCTX = cctx.CancelCTX // 设置取消函数
	hd.status = typex.DEV_UP    // 设置设备状态为 UP
	return nil                  // 返回 nil，表示启动成功
}

// OnRead 方法是 TemplateDevice 的读取回调函数。它接受命令和数据字节切片作为参数，
// 但在这个实现中，它什么也没做，只是返回 0 和 nil。
func (hd *TemplateDevice) OnRead(cmd []byte, data []byte) (int, error) {
	return 0, nil // 返回 0 表示没有数据读取，nil 表示没有错误
}

// OnWrite 方法是 TemplateDevice 的写入回调函数。它接受命令和数据字节切片作为参数，
// 但在这个实现中，它什么也没做，只是返回 0 和 nil。
func (hd *TemplateDevice) OnWrite(cmd []byte, b []byte) (int, error) {
	return 0, nil // 返回 0 表示没有数据写入，nil 表示没有错误
}

// Status 方法返回 TemplateDevice 的当前状态。在这个实现中，它总是返回 DEV_UP。
func (hd *TemplateDevice) Status() typex.DeviceState {
	return typex.DEV_UP // 表示设备状态为 UP
}

// Stop 方法用于停止 TemplateDevice 实例。它将设备状态设置为 DEV_DOWN 并调用取消函数。
func (hd *TemplateDevice) Stop() {
	hd.status = typex.DEV_DOWN // 设置设备状态为 DOWN
	hd.CancelCTX()             // 调用取消函数，可能用于取消上下文中的操作
}

// Details 方法返回 TemplateDevice 实例关联的真实设备详细信息。
func (hd *TemplateDevice) Details() *typex.Device {
	return hd.RuleEngine.GetDevice(hd.PointId) // 通过 RuleEngine 获取并返回设备的详细信息
}

// SetState 方法用于设置 TemplateDevice 的状态。
func (hd *TemplateDevice) SetState(status typex.DeviceState) {
	hd.status = status // 设置设备状态
}

// OnDCACall 方法是 TemplateDevice 的 DCA（设备控制命令）回调函数。它接受 UUID、命令和参数，
// 并返回一个 DCAResult 类型的结果。在这个实现中，它返回一个空的 DCAResult 实例。
func (hd *TemplateDevice) OnDCACall(UUID string, Command string, Args interface{}) typex.DCAResult {
	return typex.DCAResult{} // 返回一个空的 DCAResult 实例
}

// OnCtrl 方法是 TemplateDevice 的控制命令回调函数。它接受命令和数据字节切片作为参数，
// 并返回响应数据和错误。在这个实现中，它返回一个空的字节切片和 nil。
func (hd *TemplateDevice) OnCtrl(cmd []byte, args []byte) ([]byte, error) {
	return []byte{}, nil // 返回空的字节切片和 nil，表示没有响应数据和错误
}

```

## 关键接口
- Init(devId string, configMap map[string]interface{}) error
  初始化设备参数，一般在这里准备好数据，校验规则等。
- Start(cctx typex.CCTX) error
  主要工作线程，比如客户端可以在这里开启。
- Stop() error
  停止的时候回调，用来释放资源

## 运行时数据
运行时数据使用cache模块，比如Modbus初始化时向缓存器注册一个槽位：
```go
func (mdev *generic_modbus_device) Init(string, map[string]interface{}) error {
	mdev.PointId = devId
	modbuscache.RegisterSlot(mdev.PointId)
    // ....
}
```
停止的时候卸载缓存模块:
```go
func (mdev *generic_modbus_device) Stop() {
	mdev.status = typex.DEV_DOWN
	if mdev.CancelCTX != nil {
		mdev.CancelCTX()
	}
	modbuscache.UnRegisterSlot(mdev.PointId)
}
```

## 点位操作
如果涉及到点位类型的接入设备，直接在Sqlite里建表：
```sql
CREATE TABLE m_modbus_data_points (
    id          INTEGER  PRIMARY KEY AUTOINCREMENT,
    created_at  DATETIME,
    uuid        TEXT     NOT NULL,
    device_uuid TEXT     NOT NULL,
    tag         TEXT     NOT NULL,
    alias       TEXT     NOT NULL,
    function    INTEGER  NOT NULL,
    slaver_id   INTEGER  NOT NULL,
    address     INTEGER  NOT NULL,
    frequency   INTEGER  NOT NULL,
    quantity    INTEGER  NOT NULL,
    data_type   TEXT     NOT NULL,
    data_order  TEXT     NOT NULL,
    weight      REAL     NOT NULL
);

```
然后对数据CURD即可。

## 加载点位

interdb是RHILEX自带的内部村存储器，如果在接入设备里面使用全局数据库, 直接用`interdb.DB()`，以Modbus加载点位表为例：
```go
var ModbusPointList []ModbusPoint
modbusPointLoadErr := interdb.DB().Table("m_modbus_data_points").
	Where("device_uuid=?", devId).Find(&ModbusPointList).Error
if modbusPointLoadErr != nil {
	return modbusPointLoadErr
}
```