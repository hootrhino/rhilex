package test

import (
	"context"
	"testing"
	"time"

	httpserver "github.com/hootrhino/rhilex/component/apiserver"
	"github.com/hootrhino/rhilex/component/rhilexrpc"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Test_S7_PLC_Parse(t *testing.T) {
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
	})
	ctx, cancelF := typex.NewCCTX() // ,ctx, cancelF
	if err := engine.LoadInEndWithCtx(grpcInend, ctx, cancelF); err != nil {
		glogger.GLogger.Error("Rule load failed:", err)
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
				local V0 = rhilexlib:MB(">a:16 b:16 c:16 d:16 e:16", data, false)
				local a = rhilexlib:T2J(V0['a'])
				local b = rhilexlib:T2J(V0['b'])
				local c = rhilexlib:T2J(V0['c'])
				local d = rhilexlib:T2J(V0['d'])
				local e = rhilexlib:T2J(V0['e'])
				print('a ==> ', a, ' ->', rhilexlib:BS2B(a), '==> ', rhilexlib:B2I64('>', rhilexlib:BS2B(a)))
				print('b ==> ', b, ' ->', rhilexlib:BS2B(a), '==> ', rhilexlib:B2I64('>', rhilexlib:BS2B(b)))
				print('c ==> ', c, ' ->', rhilexlib:BS2B(a), '==> ', rhilexlib:B2I64('>', rhilexlib:BS2B(c)))
				print('d ==> ', d, ' ->', rhilexlib:BS2B(a), '==> ', rhilexlib:B2I64('>', rhilexlib:BS2B(d)))
				print('e ==> ', e, ' ->', rhilexlib:BS2B(a), '==> ', rhilexlib:B2I64('>', rhilexlib:BS2B(e)))
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
		Value: string([]byte{
			1, 2, 3, 4, 5, 6, 7, 8, 9,
			10, 11, 12, 13, 14, 15, 16}),
	})
	if err != nil {
		glogger.GLogger.Errorf("grpc.Dial err: %v", err)
	}
	glogger.GLogger.Infof("rhilex Rpc Call Result ====>>: %v", resp.GetMessage())

	time.Sleep(1 * time.Second)
	engine.Stop()
}
