package rhilexlib

import (
	lua "github.com/hootrhino/gopher-lua"
	rhilexg1 "github.com/hootrhino/rhilex/periphery/rhilexg1"
	"github.com/hootrhino/rhilex/typex"
)

/*
*
* DI2(0/1)
*
 */
func RHILEXG1_DO1Set(rx typex.Rhilex, uuid string) func(*lua.LState) int {
	return func(l *lua.LState) int {
		value := l.ToNumber(2)
		if value == 0 || value == 1 {
			e := rhilexg1.RHILEXG1_GPIOSetDO1((int(value)))
			if e != nil {
				l.Push(lua.LString(e.Error()))
			} else {
				l.Push(lua.LNil)
			}
		} else {
			l.Push(lua.LString("DO2 Only can set '0' or '1'."))
		}
		return 1
	}
}
func RHILEXG1_DO1Get(rx typex.Rhilex, uuid string) func(*lua.LState) int {
	return func(l *lua.LState) int {
		v, e := rhilexg1.RHILEXG1_GPIOGetDO1()
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

func RHILEXG1_DO2Set(rx typex.Rhilex, uuid string) func(*lua.LState) int {
	return func(l *lua.LState) int {
		value := l.ToNumber(2)
		if value == 0 || value == 1 {
			e := rhilexg1.RHILEXG1_GPIOSetDO2(int(value))
			if e != nil {
				l.Push(lua.LString(e.Error()))
			} else {
				l.Push(lua.LNil)
			}
		} else {
			l.Push(lua.LString("DO2 Only can set '0' or '1'."))
		}
		return 1
	}
}
func RHILEXG1_DO2Get(rx typex.Rhilex, uuid string) func(*lua.LState) int {
	return func(l *lua.LState) int {
		v, e := rhilexg1.RHILEXG1_GPIOGetDO2()
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
* DI 1,2,3 -> gpio 8-9-10
*
 */
func RHILEXG1_DI1Get(rx typex.Rhilex, uuid string) func(*lua.LState) int {
	return func(l *lua.LState) int {
		Value, e := rhilexg1.RHILEXG1_GPIOGetDI1()
		if e != nil {
			l.Push(lua.LNil)
			l.Push(lua.LString(e.Error()))
		} else {
			l.Push(lua.LNumber(Value))
			l.Push(lua.LNil)
		}
		return 2
	}
}
func RHILEXG1_DI2Get(rx typex.Rhilex, uuid string) func(*lua.LState) int {
	return func(l *lua.LState) int {
		Value, e := rhilexg1.RHILEXG1_GPIOGetDI2()
		if e != nil {
			l.Push(lua.LNil)
			l.Push(lua.LString(e.Error()))
		} else {
			l.Push(lua.LNumber(Value))
			l.Push(lua.LNil)
		}
		return 2
	}
}
func RHILEXG1_DI3Get(rx typex.Rhilex, uuid string) func(*lua.LState) int {
	return func(l *lua.LState) int {
		v, e := rhilexg1.RHILEXG1_GPIOGetDI3()
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

// User Gpio operation
// 注意：低电平亮
func Led1On(rx typex.Rhilex, uuid string) func(*lua.LState) int {
	return func(l *lua.LState) int {
		e := rhilexg1.RHILEXG1_GPIOSetUserGpio(0)
		if e != nil {
			l.Push(lua.LString(e.Error()))
		} else {
			l.Push(lua.LNil)
		}
		return 1
	}
}
func Led1Off(rx typex.Rhilex, uuid string) func(*lua.LState) int {
	return func(l *lua.LState) int {
		e := rhilexg1.RHILEXG1_GPIOSetUserGpio(1)
		if e != nil {
			l.Push(lua.LString(e.Error()))
		} else {
			l.Push(lua.LNil)
		}
		return 1
	}
}
