package rulexlib

import (
	lua "github.com/hootrhino/gopher-lua"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/sirupsen/logrus"
)

// | Ws log topic                     | 用途               |
// | -------------------------------- | ------------------ |
// | plugin/ICMPSenderPing/ICMPSender | 网络测速插件的日志 |
// | rule/test/${uuid}                | 规则测试日志       |
// | rule/log/${uuid}                 | 规则运行时的日志   |
// | app/console/${uuid}              | app                |
// | device/rule/test/${uuid}         | Test device        |
// | inend/rule/test/${uuid}          | Test inend         |
// | outend/rule/test/${uuid}         | Test outend        |

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
