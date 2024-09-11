package test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	httpserver "github.com/hootrhino/rhilex/component/apiserver"
	"github.com/hootrhino/rhilex/component/rhilexrpc"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type _rpcCodecServer struct {
	rhilexrpc.UnimplementedRhilexRpcServer
}

func (_rpcCodecServer) Decode(c context.Context, req *rhilexrpc.RpcRequest) (resp *rhilexrpc.RpcResponse, err error) {
	glogger.GLogger.Debug("[REQUEST]=====================> ", req.String())
	resp = new(rhilexrpc.RpcResponse)
	resp.Data = []byte("DecodeOK")
	return resp, nil
}
func (_rpcCodecServer) Encode(c context.Context, req *rhilexrpc.RpcRequest) (resp *rhilexrpc.RpcResponse, err error) {
	glogger.GLogger.Debug("[REQUEST]=====================> ", req.String())
	resp = new(rhilexrpc.RpcResponse)
	resp.Data = []byte("EncodeOK")
	return resp, nil
}

/*
*
*
*
 */
func _startServer() {
	listener, err := net.Listen("tcp", ":1998")
	if err != nil {
		glogger.GLogger.Fatal(err)
		return
	}
	rpcServer := grpc.NewServer()
	rhilexrpc.RegisterRhilexRpcServer(rpcServer, new(_rpcCodecServer))
	go func(c context.Context) {
		defer listener.Close()
		glogger.GLogger.Info("rpcCodecServer started on", listener.Addr())
		rpcServer.Serve(listener)
	}(context.TODO())

}
func Test_Codec(t *testing.T) {
	engine := RunTestEngine()
	engine.Start()

	_startServer()
	hh := httpserver.NewHttpApiServer(engine)
	// HttpApiServer loaded default
	if err := engine.LoadPlugin("plugin.http_server", hh); err != nil {
		glogger.GLogger.Fatal("Rule load failed:", err)
	}
	// Grpc Inend
	grpcInend := typex.NewInEnd("GRPC",
		"rhilex Grpc InEnd",
		"rhilex Grpc InEnd", map[string]interface{}{
			"port": 2581,
		})
	ctx, cancelF := typex.NewCCTX() // ,ctx, cancelF
	if err := engine.LoadInEndWithCtx(grpcInend, ctx, cancelF); err != nil {
		glogger.GLogger.Fatal("Rule load failed:", err)
	}
	grpcCodec1 := typex.NewOutEnd("GRPC_CODEC_TARGET",
		"GRPC_CODEC_TARGET",
		"GRPC_CODEC_TARGET", map[string]interface{}{
			"host": "127.0.0.1",
			"port": 1998,
			"type": "DECODE",
		})
	grpcCodec1.UUID = "grpcCodec001"
	ctx1, cancelF1 := typex.NewCCTX() // ,ctx, cancelF
	if err := engine.LoadOutEndWithCtx(grpcCodec1, ctx1, cancelF1); err != nil {
		glogger.GLogger.Fatal("grpcCodec load failed:", err)
	}
	grpcCodec2 := typex.NewOutEnd("GRPC_CODEC_TARGET",
		"GRPC_CODEC_TARGET",
		"GRPC_CODEC_TARGET", map[string]interface{}{
			"host": "127.0.0.1",
			"port": 1998,
			"type": "ENCODE",
		})
	grpcCodec2.UUID = "grpcCodec002"
	ctx2, cancelF2 := typex.NewCCTX() // ,ctx, cancelF
	if err := engine.LoadOutEndWithCtx(grpcCodec2, ctx2, cancelF2); err != nil {
		glogger.GLogger.Fatal("grpcCodec load failed:", err)
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
			print('rhilexlib:RPCDEC ==> ', rhilexlib:RPCDEC('grpcCodec001', data))
			print('rhilexlib:RPCENC ==> ', rhilexlib:RPCENC('grpcCodec002', data))
				return true, args
			end
		}`,
		`function Failed(error) print("[LUA Failed Callback]", error) end`)
	if err := engine.LoadRule(rule); err != nil {
		glogger.GLogger.Fatal(err)
	}
	grpcConnection, err := grpc.Dial("127.0.0.1:2581", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		glogger.GLogger.Error(err)
	}
	defer grpcConnection.Close()
	client := rhilexrpc.NewRhilexRpcClient(grpcConnection)

	for i := 0; i < 10; i++ {
		resp, err := client.Request(context.Background(), &rhilexrpc.RpcRequest{
			Value: fmt.Sprintf(`{"co2":10,"hum":30,"lex":22,"temp":100,"idx":%d}`, i),
		})
		if err != nil {
			t.Fatalf("grpc.Dial err: %v", err)
		}
		t.Logf("rhilex Rpc Call Result ====>>: %v", resp.GetMessage())
	}
	time.Sleep(1 * time.Second)
	engine.Stop()
}
