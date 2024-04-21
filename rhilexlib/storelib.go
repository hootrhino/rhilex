package rhilexlib

import (
	"time"

	lua "github.com/hootrhino/gopher-lua"
	"github.com/hootrhino/rhilex/component/interkv"
	"github.com/hootrhino/rhilex/typex"
)

func StoreSet(rx typex.Rhilex, uuid string) func(l *lua.LState) int {
	return func(l *lua.LState) int {
		k := l.ToString(2)
		v := l.ToString(3)
		interkv.GlobalStore.Set(k, v)
		return 0
	}
}
func StoreGet(rx typex.Rhilex, uuid string) func(l *lua.LState) int {
	return func(l *lua.LState) int {
		k := l.ToString(2)
		v := interkv.GlobalStore.Get(k)
		if v == "" {
			l.Push(lua.LNil)
		} else {
			l.Push(lua.LString(v))
		}
		return 1
	}

}
func StoreDelete(rx typex.Rhilex, uuid string) func(l *lua.LState) int {
	return func(l *lua.LState) int {
		k := l.ToString(2)
		interkv.GlobalStore.Delete(k)
		return 0
	}
}

func StoreSetWithDuration(rx typex.Rhilex, uuid string) func(l *lua.LState) int {
	return func(l *lua.LState) int {
		k := l.ToString(2)
		v := l.ToString(3)
		d := l.ToInt64(4) // second
		duration := time.Duration(d) * time.Second
		interkv.GlobalStore.SetWithDuration(k, v, duration)
		return 0
	}
}
