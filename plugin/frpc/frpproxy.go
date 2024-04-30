package frpc

import (
	"github.com/hootrhino/rhilex/typex"
	"gopkg.in/ini.v1"
)

type FrpcConfig struct {
	Enable     bool   `ini:"enable"`
	ServerAddr string `ini:"server_addr"`
	ServerPort int    `ini:"server_port"`
	Name       string `ini:"name"`
	Type       string `ini:"type"`
	LocalIP    string `ini:"local_ip"`
	LocalPort  int    `ini:"local_port"`
	RemotePort int    `ini:"remote_port"`
}

type FrpcProxy struct {
	mainConfig FrpcConfig
}

func NewFrpcProxy() *FrpcProxy {
	return &FrpcProxy{
		mainConfig: FrpcConfig{
			Enable:     false,
			ServerAddr: "127.0.0.1",
			ServerPort: 7000,
			Name:       "rhilex-web-dashboard",
			Type:       "tcp",
			LocalIP:    "127.0.0.1",
			LocalPort:  2580,
			RemotePort: 20000,
		},
	}
}

func (dm *FrpcProxy) Init(config *ini.Section) error {
	return nil
}

func (dm *FrpcProxy) Start(typex.Rhilex) error {
	return nil
}
func (dm *FrpcProxy) Stop() error {
	return nil
}

func (dm *FrpcProxy) PluginMetaInfo() typex.XPluginMetaInfo {
	return typex.XPluginMetaInfo{
		UUID:     "RHILEX-NAT-TUNNEL-PLUGIN",
		Name:     "Frpc Proxy Client",
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
func (dm *FrpcProxy) Service(arg typex.ServiceArg) typex.ServiceResult {
	return typex.ServiceResult{}
}
