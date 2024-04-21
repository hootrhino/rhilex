package typex

import lua "github.com/hootrhino/gopher-lua"

// XLib: 库函数接口; TODO: V0.1.2废弃
// LibFun 方法注册一个 Lua 库函数。
// Rhilex 是 Lua 的注册表（registry），用于存储 Lua C 函数。
// 第二个参数是函数名称，在 Lua 中被调用。
// 返回值是一个 Lua 闭包，当在 Lua 中调用该函数时，会执行这个闭包。
// LibFun(Rhilex lua.Registry, name string) func(*lua.LState) int
type XLib interface {
	Name() string
	LibFun(Rhilex, string) func(*lua.LState) int
}
