package rhilexlib

import (
	"encoding/json"

	lua "github.com/hootrhino/gopher-lua"
	"github.com/hootrhino/rhilex/typex"
)

/*
*
* 指令执行成功
*
 */

func FinishCmd(rx typex.Rhilex, uuid string) func(*lua.LState) int {
	return func(l *lua.LState) int {
		cmdId := l.ToString(2)
		stateTargetId := l.ToString(3)
		bytes, _ := json.Marshal(map[string]interface{}{
			"type":  "finishCmd",
			"cmdId": cmdId,
		})
		write(rx, stateTargetId, string(bytes))
		return 0
	}
}

/*
*
* 指令执行失败
*
 */

func FailedCmd(rx typex.Rhilex, uuid string) func(*lua.LState) int {
	return func(l *lua.LState) int {
		cmdId := l.ToString(2)
		stateTargetId := l.ToString(3)
		bytes, _ := json.Marshal(map[string]interface{}{
			"type":  "failedCmd",
			"cmdId": cmdId,
		})
		write(rx, stateTargetId, string(bytes))
		return 0
	}
}
func write(e typex.Rhilex, uuid string, incoming string) {
}
