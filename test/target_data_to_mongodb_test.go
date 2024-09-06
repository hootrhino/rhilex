package test

import (
	"context"
	"fmt"

	"testing"
	"time"

	httpserver "github.com/hootrhino/rhilex/component/apiserver"
	"github.com/hootrhino/rhilex/component/rhilexrpc"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Test_DataToMongoDB(t *testing.T) {
	engine := RunTestEngine()
	engine.Start()

	hh := httpserver.NewHttpApiServer(engine)
	// HttpApiServer loaded default
	if err := engine.LoadPlugin("plugin.http_server", hh); err != nil {
		glogger.GLogger.Fatal("Rule load failed:", err)
	}
	// Grpc Inend
	grpcInend := typex.NewInEnd("GRPC", "rhilex Grpc InEnd", "rhilex Grpc InEnd", map[string]interface{}{
		"port": 2581,
		"host": "127.0.0.1",
	})
	ctx, cancelF := typex.NewCCTX() // ,ctx, cancelF
	if err := engine.LoadInEndWithCtx(grpcInend, ctx, cancelF); err != nil {
		glogger.GLogger.Error("Rule load failed:", err)
	}
	//
	ts := fmt.Sprintf("%v", time.Now().UnixMicro())
	mongoOut := typex.NewOutEnd(typex.MONGO_SINGLE,
		"MONGO_SINGLE",
		"MONGO_SINGLE", map[string]interface{}{
			"mongoUrl":   "mongodb://root:root@127.0.0.1:27017/?connect=direct",
			"database":   "temp_gateway_test_" + ts,
			"collection": "temp_gateway_test_" + ts,
		})
	mongoOut.UUID = "mongoOut"
	ctx1, cancelF1 := typex.NewCCTX() // ,ctx, cancelF
	if err := engine.LoadOutEndWithCtx(mongoOut, ctx1, cancelF1); err != nil {
		glogger.GLogger.Fatal(err)
	}
	rule := typex.NewRule(engine,
		"uuid",
		"Just a test",
		"Just a test",
		[]string{grpcInend.UUID}[0],
		"",
		`function Success() print("[LUA Success Callback]=> OK") end`,
		`
		Actions = {
			function(args)
			    local err = data:ToMongoDB('mongoOut', data)
				print("[LUA DataToMongoDB] ==>", err)
				return true, args
			end
		}`,
		`function Failed(error) print("[LUA Failed Callback]", error) end`)
	if err := engine.LoadRule(rule); err != nil {
		glogger.GLogger.Error(err)
	}
	conn, err := grpc.Dial("127.0.0.1:2581", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		glogger.GLogger.Error(err)
	}
	defer conn.Close()
	client := rhilexrpc.NewRhilexRpcClient(conn)

	resp, err := client.Work(context.Background(), &rhilexrpc.Data{
		Value: `[{"co2":10,"hum":30,"lex":22,"temp":100}]`,
	})
	if err != nil {
		glogger.GLogger.Error(err)
	}
	glogger.GLogger.Infof("rhilex Rpc Call Result ====>>: %v", resp.GetMessage())

	time.Sleep(1 * time.Second)
	engine.Stop()
}
