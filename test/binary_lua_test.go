package test

import (
	"context"
	"testing"
	"time"

	httpserver "github.com/hootrhino/rhilex/component/apiserver"
	"github.com/hootrhino/rhilex/component/rhilexrpc"
	"github.com/hootrhino/rhilex/core"
	"github.com/hootrhino/rhilex/engine"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/plugin/demo_plugin"
	"github.com/hootrhino/rhilex/typex"

	"google.golang.org/grpc"
)

func Test_Binary_LUA_Parse(t *testing.T) {
	engine := engine.InitRuleEngine(core.InitGlobalConfig("conf/rhilex.ini"))
	engine.Start()

	hh := httpserver.NewHttpApiServer(engine)

	// HttpApiServer loaded default
	if err := engine.LoadPlugin("plugin.http_server", hh); err != nil {
		glogger.GLogger.Fatal("Rule load failed:", err)
	}
	// Load a demo plugin
	if err := engine.LoadPlugin("plugin.demo", demo_plugin.NewDemoPlugin()); err != nil {
		glogger.GLogger.Error("Rule load failed:", err)
	}
	// Grpc Inend
	grpcInend := typex.NewInEnd("GRPC", "rhilex Grpc InEnd", "rhilex Grpc InEnd", map[string]interface{}{
		"port": "2581",
	})
	ctx, cancelF := typex.NewCCTX()
	if err := engine.LoadInEndWithCtx(grpcInend, ctx, cancelF); err != nil {
		glogger.GLogger.Error("Rule load failed:", err)
	}
	//
	// Load Rule
	//
	rule := typex.NewRule(engine,
		"uuid",
		"Just a test",
		"Just a test",
		grpcInend.UUID,
		"",
		`function Success() print("[LUA Success Callback]=> OK") end`,
		`
		Actions = {
			--        ┌───────────────────────────────────────────────┐
			-- data = |00 00 00 01|00 00 00 02|00 00 00 03|00 00 00 04|
			--        └───────────────────────────────────────────────┘
			function(args)
				local json = require("json")
				local V6 = json.encode(rhilexlib:MB("<a:8 b:8 c:8 d:8", data, false))
				print("[LUA Actions Callback 5, rhilex.MatchBinary] ==>", V6)
				return true, args
			end
		}`,
		`function Failed(error) print("[LUA Failed Callback]", error) end`)
	if err := engine.LoadRule(rule); err != nil {
		glogger.GLogger.Error(err)
	}
	conn, err := grpc.Dial("127.0.0.1:2581")
	if err != nil {
		glogger.GLogger.Error(err)
	}
	defer conn.Close()
	client := rhilexrpc.NewRhilexRpcClient(conn)
	for i := 0; i < 2; i++ {
		glogger.GLogger.Infof("rhilex Rpc Call ==========================>>: %v", i)
		resp, err := client.Work(context.Background(), &rhilexrpc.Data{
			Value: string([]byte{
				1, 2, 3, 4, 5, 6, 7, 8, 9,
				10, 11, 12, 13, 14, 15, 16}),
		})
		if err != nil {
			glogger.GLogger.Error(err)

		}
		glogger.GLogger.Infof("rhilex Rpc Call Result ====>>: %v", resp.GetMessage())
	}

	time.Sleep(1 * time.Second)
	engine.Stop()
}
