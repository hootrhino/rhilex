package rhilexlib

import (
	"github.com/hootrhino/rhilex/typex"

	lua "github.com/hootrhino/gopher-lua"
)

func Throw(rx typex.Rhilex, uuid string) func(l *lua.LState) int {
	return func(l *lua.LState) int {
		l.RaiseError(l.ToString(2))
		return 0
	}
}
