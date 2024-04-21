package test

import (
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/hootrhino/rhilex/component/ruleengine"
	core "github.com/hootrhino/rhilex/config"
	"github.com/hootrhino/rhilex/engine"
	"github.com/hootrhino/rhilex/typex"
)

func TestLuaSyntax1(t *testing.T) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGABRT)
	engine := engine.InitRuleEngine(core.InitGlobalConfig("config/rhilex.ini"))
	engine.Start()
	// Grpc Inend
	grpcInend := typex.NewInEnd("GRPC", "rhilex Grpc InEnd", "rhilex Grpc InEnd", map[string]interface{}{
		"port": "2581",
	})
	rule := typex.NewRule(engine,
		"uuid",
		"Just a test",
		"Just a test",
		[]string{grpcInend.UUID}[0],
		"",
		`function Success() print("[LUA Success]==========================> OK") end`,
		`
		Actions = {
			function(args)
			    print("[LUA Actions Callback]==========================> ", data)
				return true, args
			end,
			function(args)
			    print("[LUA Actions Callback]==========================> ", data)
				return true, args
			end
		}`,
		`function Failed(error) print("[LUA Failed]==========================> OK", error) end`,
	)
	err := ruleengine.VerifyLuaSyntax(rule)
	t.Log(err)
}
func TestLuaSyntax2(t *testing.T) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGABRT)
	engine := engine.InitRuleEngine(core.InitGlobalConfig("config/rhilex.ini"))
	engine.Start()
	// Grpc Inend
	grpcInend := typex.NewInEnd("GRPC", "rhilex Grpc InEnd", "rhilex Grpc InEnd", map[string]interface{}{
		"port": "2581",
	})
	rule := typex.NewRule(engine,
		"uuid",
		"Just a test",
		"Just a test",
		[]string{grpcInend.UUID}[0],
		"",
		`function Success() print("[LUA Success]==========================> OK") end`,
		`
		Actions = {
			function(args)
				print("[LUA Actions Callback]==========================> ", data)
				return true, args
		    end,
			function(args)
			    print("[LUA Actions Callback]==========================> ", data)
				return true, args
			end,,,,,++1122++33++44
			function(args)
			    print("[LUA Actions Callback]==========================> ", data)
				return true, args
			end
		}`,
		`function Failed(error) print("[LUA Failed]==========================> OK", error) end`,
	)
	err := ruleengine.VerifyLuaSyntax(rule)
	t.Log(err)
}
