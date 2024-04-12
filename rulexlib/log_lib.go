package rulexlib

import (
	lua "github.com/hootrhino/gopher-lua"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/sirupsen/logrus"
)

// Topic
// app:         app/console/$uuid
// rule:        rule/$uuid
// Test device: device/rule/test/$uuid
// Test inend:  inend/rule/test/$uuid
// Test outend: outend/rule/test/$uuid
/*
*
* APP debug输出, Debug(".....")
*
 */
func DebugAPP(rx typex.RuleX, uuid string) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		top := L.GetTop()
		content := ""
		for i := 1; i <= top; i++ {
			content += L.ToStringMeta(L.Get(i)).String()
			if i != top {
				content += "\t"
			}
		}
		glogger.GLogger.WithFields(logrus.Fields{
			"topic": "app/console/" + uuid,
		}).Info(content)
		return 0
	}
}

/*
*
* 辅助Debug使用, 用来向前端Dashboard打印日志的时候带上ID
*
 */
func DebugRule(rx typex.RuleX, uuid string) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		top := L.GetTop()
		content := ""
		for i := 1; i <= top; i++ {
			content += L.ToStringMeta(L.Get(i)).String()
			if i != top {
				content += "\t"
			}
		}
		glogger.GLogger.WithFields(logrus.Fields{
			"topic": "rule/log/" + uuid,
		}).Info(content)
		return 0
	}
}
