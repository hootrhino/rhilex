package rhilexlib

import (
	"errors"

	"github.com/hootrhino/rhilex/component/interqueue"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
)

func handleDataFormat(e typex.Rhilex, uuid string, incoming string) error {
	outEnd := e.GetOutEnd(uuid)
	if outEnd != nil {
		return interqueue.DefaultDataCacheQueue.PushOutQueue(outEnd, incoming)
	}
	msg := "target not found:" + uuid
	glogger.GLogger.Error(msg)
	return errors.New(msg)

}
