package demo_plugin

import (
	"github.com/hootrhino/rhilex/typex"
	"gopkg.in/ini.v1"
)

type DemoPlugin struct {
	uuid string
}

func NewDemoPlugin() *DemoPlugin {
	return &DemoPlugin{
		uuid: "DEMO01",
	}
}

func (dm *DemoPlugin) Init(config *ini.Section) error {
	return nil
}

func (dm *DemoPlugin) Start(typex.Rhilex) error {
	return nil
}
func (dm *DemoPlugin) Stop() error {
	return nil
}

func (dm *DemoPlugin) PluginMetaInfo() typex.XPluginMetaInfo {
	return typex.XPluginMetaInfo{
		UUID:     dm.uuid,
		Name:     "DemoPlugin",
		Version:  "v0.0.1",
		Homepage: "/",
		HelpLink: "/",
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
func (dm *DemoPlugin) Service(arg typex.ServiceArg) typex.ServiceResult {
	return typex.ServiceResult{}
}
