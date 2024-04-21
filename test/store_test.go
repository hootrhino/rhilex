package test

import (
	"testing"

	ruleengine "github.com/hootrhino/rhilex/ruleengine"
)

func Test_get_set(t *testing.T) {
	ruleengine.StartStore(1024)
	ruleengine.GlobalStore.Set("k", "v")
	t.Log(ruleengine.GlobalStore.Get("k"))
}
