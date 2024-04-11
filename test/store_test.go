package test

import (
	"testing"

	"github.com/hootrhino/rhilex/core"
)

func Test_get_set(t *testing.T) {
	core.StartStore(1024)
	core.GlobalStore.Set("k", "v")
	t.Log(core.GlobalStore.Get("k"))
}
