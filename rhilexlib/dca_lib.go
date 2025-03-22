package rhilexlib

import (
	lua "github.com/hootrhino/gopher-lua"
	"github.com/hootrhino/rhilex/typex"
)

/*
*
* 设备功能调用
*
 */
func DCACall(rx typex.Rhilex, uuid string) func(*lua.LState) int {
	return func(l *lua.LState) int {
		UUID := l.ToString(2)
		Command := l.ToString(3)
		// 参数必须是个Table: [arg0, arg1, arg2.....]
		LuaTArgs := l.ToTable(4)
		Device := rx.GetDevice(UUID)
		// glogger.GLogger.Infof("DCACall => %s:%s(%v)", UUID, Command, LuaTArgs)
		CallArgs := []any{}
		LuaTArgs.ForEach(func(k, v lua.LValue) {
			CallArgs = append(CallArgs, v)
		})
		if Device != nil {
			r := Device.Device.OnDCACall(UUID, Command, CallArgs)
			if r.Error != nil {
				l.Push(lua.LNil)
				l.Push(lua.LString(r.Error.Error()))
			} else {
				l.Push(lua.LString(r.Data))
				l.Push(lua.LNil)
			}
		} else {
			l.Push(lua.LNil)
			l.Push(lua.LString("Device not exists: " + UUID))
		}
		return 2
	}
}
