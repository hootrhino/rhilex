package icmpsender

import (
	"fmt"
	"time"

	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

type ICMPSender struct {
	uuid    string
	pinging bool
}

func NewICMPSender() *ICMPSender {
	return &ICMPSender{
		uuid:    "ICMPSender",
		pinging: false,
	}
}

func (dm *ICMPSender) Init(config *ini.Section) error {
	return nil
}

func (dm *ICMPSender) Start(typex.Rhilex) error {
	return nil
}
func (dm *ICMPSender) Stop() error {
	return nil
}

func (dm *ICMPSender) PluginMetaInfo() typex.XPluginMetaInfo {
	return typex.XPluginMetaInfo{
		UUID:     dm.uuid,
		Name:     "ICMP Sender",
		Version:  "v1.0.0",
		Homepage: "https://www.hootrhino.com",
		HelpLink: "https://www.hootrhino.com",
		Author:   "RHILEXTeam",
		Email:    "RHILEXTeam@hootrhino.com",
		License:  "",
	}
}

/*
*
* 服务调用接口
*
 */
func (dm *ICMPSender) Service(arg typex.ServiceArg) typex.ServiceResult {
	// ping 8.8.8.8
	Fields := logrus.Fields{
		"topic": "plugin/ICMPSenderPing/ICMPSender",
	}
	out := typex.ServiceResult{Out: []map[string]interface{}{}}
	if dm.pinging {
		glogger.GLogger.WithFields(Fields).Info("ICMPSender pinging now:", arg.Args)
		return out
	}
	if arg.Name == "ping" {
		dm.pinging = true
		go func(cs *ICMPSender) {
			defer func() {
				cs.pinging = false
			}()
			select {
			case <-typex.GCTX.Done():
				{
					return
				}
			default:
				{
				}
			}
			switch tt := arg.Args.(type) {
			case []interface{}:
				if len(tt) < 1 {
					break
				}
				for i := 0; i < 5; i++ {
					switch ip := tt[0].(type) {
					case string:
						if Duration, err := pingQ(ip, 1000*time.Millisecond); err != nil {
							glogger.GLogger.WithFields(Fields).Info(fmt.Sprintf(
								"[Count:%d] Ping Error:%s", i, err.Error()))
						} else {
							glogger.GLogger.WithFields(Fields).Info(fmt.Sprintf(
								"[Count:%d] Ping Reply From %s: time=%v ms TTL=128", i, tt, Duration))
						}
						// 300毫秒
						time.Sleep(100 * time.Millisecond)
					}

				}
			default:
				{
					glogger.GLogger.WithFields(Fields).Info("Unknown service name:", arg.Name)
				}
			}
		}(dm)

	}
	return out
}
