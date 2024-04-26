package rhilexlib

import (
	lua "github.com/hootrhino/gopher-lua"
	archsupport "github.com/hootrhino/rhilex/bspsupport"
	"github.com/hootrhino/rhilex/typex"
)

/*
*
* 读GPIO， lua的函数调用应该是这样: rhinopi:GPIOGet(pin) -> v,error
*
 */
func RASPI4_GPIOGet(rx typex.Rhilex, uuid string) func(L *lua.LState) int {
	return func(l *lua.LState) int {
		pin := l.ToNumber(2)
		v, e := archsupport.RASPI4_GPIOGet(int(pin))
		if e != nil {
			l.Push(lua.LNil)
			l.Push(lua.LString(e.Error()))
		} else {
			l.Push(lua.LNumber(v))
			l.Push(lua.LNil)
		}
		return 2
	}
}

/*
*
* 写GPIO， lua的函数调用应该是这样: rhinopi:GPIOSet(pin, v) -> error
*
 */
func RASPI4_GPIOSet(rx typex.Rhilex, uuid string) func(L *lua.LState) int {
	return func(l *lua.LState) int {
		pin := l.ToNumber(2)
		value := l.ToNumber(3)
		_, e := archsupport.RASPI4_GPIOSet(int(pin), int(value))
		if e != nil {
			l.Push(lua.LString(e.Error()))
		} else {
			l.Push(lua.LNil)
		}
		return 1
	}
}
