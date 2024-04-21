package test

import (
	"encoding/json"
	"testing"

	httpserver "github.com/hootrhino/rhilex/component/apiserver"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/component/apiserver/service"
	"github.com/hootrhino/rhilex/core"
	"github.com/hootrhino/rhilex/engine"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
)

// 初始化一些测试数据
func TestInitData(t *testing.T) {
	engine := engine.InitRuleEngine(core.InitGlobalConfig("conf/rhilex.ini"))
	engine.Start()
	hh := httpserver.NewHttpApiServer(engine)
	// HttpApiServer loaded default
	if err := engine.LoadPlugin("plugin.http_server", hh); err != nil {
		glogger.GLogger.Fatal("Rule load failed:", err)
	}
	// Grpc Inend
	grpcInend := typex.NewInEnd("GRPC", "rhilex Grpc InEnd", "rhilex Grpc InEnd", map[string]interface{}{
		"port": "2581",
	})
	b1, _ := json.Marshal(grpcInend.Config)
	service.InsertMInEnd(&model.MInEnd{
		UUID:        grpcInend.UUID,
		Type:        grpcInend.Type.String(),
		Name:        grpcInend.Name,
		Config:      string(b1),
		Description: grpcInend.Description,
	})
	// CoAP Inend
	coapInend := typex.NewInEnd("COAP", "rhilex COAP InEnd", "rhilex COAP InEnd", map[string]interface{}{
		"port": "2582",
	})
	b2, _ := json.Marshal(coapInend.Config)
	service.InsertMInEnd(&model.MInEnd{
		UUID:        coapInend.UUID,
		Type:        coapInend.Type.String(),
		Name:        coapInend.Name,
		Config:      string(b2),
		Description: coapInend.Description,
	})
	// Http Inend
	httpInend := typex.NewInEnd("HTTP", "rhilex HTTP InEnd", "rhilex HTTP InEnd", map[string]interface{}{
		"port": "2583",
	})
	b3, _ := json.Marshal(httpInend.Config)
	service.InsertMInEnd(&model.MInEnd{
		UUID:        httpInend.UUID,
		Type:        httpInend.Type.String(),
		Name:        httpInend.Name,
		Config:      string(b3),
		Description: httpInend.Description,
	})

	// Udp Inend
	udpInend := typex.NewInEnd("UDP", "rhilex UDP InEnd", "rhilex UDP InEnd", map[string]interface{}{
		"port": "2584",
	})
	b4, _ := json.Marshal(udpInend.Config)
	service.InsertMInEnd(&model.MInEnd{
		UUID:        udpInend.UUID,
		Type:        udpInend.Type.String(),
		Name:        udpInend.Name,
		Config:      string(b4),
		Description: udpInend.Description,
	})

	rule := typex.NewRule(engine,
		"uuid",
		"Just a test",
		"Just a test",
		[]string{grpcInend.UUID}[0],
		"",
		`function Success() print("[LUA Success]OK") end`,
		`
			Actions = {
				function(args)
					print("[LUA Actions Callback]", data)
					return true, args
				end
			}`,
		`function Failed(error) print("[LUA Failed]OK", error) end`)
	service.InsertMRule(&model.MRule{
		Name:        rule.Name,
		Description: rule.Description,
		SourceId:    rule.FromSource,
		DeviceId:    rule.FromDevice,
		Actions:     rule.Actions,
		Success:     rule.Success,
		Failed:      rule.Failed,
	})
	engine.Stop()
}
