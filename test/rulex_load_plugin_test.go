package test

import (
	ttyterminal "github.com/hootrhino/rhilex/plugin/ttyd_terminal"
	"testing"

	"time"
)

func Test_rhilex_load_plugin(t *testing.T) {
	engine := RunTestEngine()
	ttyd := ttyterminal.NewWebTTYPlugin()
	if err := engine.LoadPlugin("plugin.ttyd", ttyd); err != nil {
		t.Fatal(err)
	}
	engine.Start()
	time.Sleep(20 * time.Second)
	engine.Stop()
}
