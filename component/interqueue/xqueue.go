package interqueue

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hootrhino/rhilex/component/intermetric"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
)

var DefaultDataCacheQueue XQueue

/*
*
* XQueue
*
 */
type XQueue interface {
	GetQueue() chan QueueData
	GetInQueue() chan QueueData
	GetOutQueue() chan QueueData
	GetDeviceQueue() chan QueueData
	GetSize() int
	Push(QueueData) error
	PushInQueue(in *typex.InEnd, data string) error
	PushOutQueue(in *typex.OutEnd, data string) error
	PushDeviceQueue(in *typex.Device, data string) error
}

type QueueData struct {
	Debug bool // 是否是Debug消息
	I     *typex.InEnd
	O     *typex.OutEnd
	D     *typex.Device
	E     typex.Rhilex
	Data  string
}

func (qd QueueData) String() string {
	return "QueueData@In:" + qd.I.UUID + ", Data:" + qd.Data
}

/*
*
* NewXQueue
*
 */

/*
*
* DataCacheQueue
*
 */
type DataCacheQueue struct {
	Queue       chan QueueData
	OutQueue    chan QueueData
	InQueue     chan QueueData
	DeviceQueue chan QueueData
	rhilex      typex.Rhilex
}

func InitDataCacheQueue(rhilex typex.Rhilex, maxQueueSize int) XQueue {
	DefaultDataCacheQueue = &DataCacheQueue{
		Queue:       make(chan QueueData, maxQueueSize),
		OutQueue:    make(chan QueueData, maxQueueSize),
		InQueue:     make(chan QueueData, maxQueueSize),
		DeviceQueue: make(chan QueueData, maxQueueSize),
		rhilex:      rhilex,
	}
	return DefaultDataCacheQueue
}
func (q *DataCacheQueue) GetSize() int {
	return cap(q.Queue)
}

/*
*
* Push
*
 */
func (q *DataCacheQueue) Push(d QueueData) error {
	// 动态扩容
	// if len(q.Queue)+1 > q.GetSize() {
	// }
	if len(q.Queue)+1 > q.GetSize() {
		msg := fmt.Sprintf("attached max queue size, max size is:%v, current size is: %v",
			q.GetSize(), len(q.Queue)+1)
		glogger.GLogger.Error(msg)
		return errors.New(msg)
	} else {
		q.Queue <- d
		return nil
	}
}

func processOutQueueData(qd QueueData, e typex.Rhilex) {
	if qd.O != nil {
		v, ok := e.AllOutEnds().Load(qd.O.UUID)
		if ok {
			target := v.(*typex.OutEnd).Target
			if target == nil {
				return
			}
			if _, err := target.To(qd.Data); err != nil {
				glogger.GLogger.Error(err)
				intermetric.IncOutFailed()
			} else {
				intermetric.IncOut()
			}
		}
	}
}

/*
*
* GetQueue
*
 */
func (q *DataCacheQueue) GetQueue() chan QueueData {
	return q.Queue
}

/*
*
* GetQueue
*
 */
func (q *DataCacheQueue) GetInQueue() chan QueueData {
	return q.InQueue
}

/*
*
* GetQueue
*
 */
func (q *DataCacheQueue) GetOutQueue() chan QueueData {
	return q.OutQueue
}

/*
*
*GetDeviceQueue
*
 */
func (q *DataCacheQueue) GetDeviceQueue() chan QueueData {
	return q.DeviceQueue
}

// TODO: 下个版本更换为可扩容的Chan
func StartDataCacheQueue() {
	ctx := typex.GCTX
	xQueue := DefaultDataCacheQueue
	go func(ctx context.Context, xQueue XQueue) {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				// 优雅地处理上下文取消
				return
			case qd := <-xQueue.GetInQueue():
				if qd.I != nil {
					qd.E.RunSourceCallbacks(qd.I, qd.Data)
				}
			case qd := <-xQueue.GetDeviceQueue():
				if qd.D != nil {
					qd.E.RunDeviceCallbacks(qd.D, qd.Data)
				}
			case qd := <-xQueue.GetOutQueue():
				processOutQueueData(qd, qd.E)
			case <-ticker.C:

			}
		}
	}(ctx, xQueue)
}

/*
*
*PushInQueue
*
 */
func (q *DataCacheQueue) PushInQueue(in *typex.InEnd, data string) error {
	qd := QueueData{
		E:    q.rhilex,
		I:    in,
		O:    nil,
		Data: data,
	}
	err := q.pushIn(qd)
	if err != nil {
		glogger.GLogger.Error("Push InQueue error:", err)
		intermetric.IncInFailed()
	} else {
		intermetric.IncIn()
	}
	return err
}

/*
*
* PushDeviceQueue
*
 */
func (q *DataCacheQueue) PushDeviceQueue(Device *typex.Device, data string) error {
	qd := QueueData{
		D:    Device,
		E:    q.rhilex,
		I:    nil,
		O:    nil,
		Data: data,
	}
	err := q.pushDevice(qd)
	if err != nil {
		glogger.GLogger.Error("Push Device Queue error:", err)
		intermetric.IncInFailed()
	} else {
		intermetric.IncIn()
	}
	return err
}

/*
*
* PushOutQueue
*
 */
func (q *DataCacheQueue) PushOutQueue(out *typex.OutEnd, data string) error {
	qd := QueueData{
		E:    q.rhilex,
		D:    nil,
		I:    nil,
		O:    out,
		Data: data,
	}
	err := q.pushOut(qd)
	if err != nil {
		glogger.GLogger.Error("Push OutQueue error:", err)
		intermetric.IncInFailed()
	} else {
		intermetric.IncIn()
	}
	return err
}

/*
*
* Push
*
 */
func (q *DataCacheQueue) pushIn(d QueueData) error {
	// 动态扩容
	// if len(q.Queue)+1 > q.GetSize() {
	// }
	if len(q.InQueue)+1 > q.GetSize() {
		msg := fmt.Sprintf("attached max queue size, max size is:%v, current size is: %v",
			q.GetSize(), len(q.InQueue)+1)
		glogger.GLogger.Error(msg)
		return errors.New(msg)
	} else {
		q.InQueue <- d
		return nil
	}
}

/*
*
* Push
*
 */
func (q *DataCacheQueue) pushOut(d QueueData) error {
	// 动态扩容
	// if len(q.Queue)+1 > q.GetSize() {
	// }
	if len(q.OutQueue)+1 > q.GetSize() {
		msg := fmt.Sprintf("attached max queue size, max size is:%v, current size is: %v",
			q.GetSize(), len(q.OutQueue)+1)
		glogger.GLogger.Error(msg)
		return errors.New(msg)
	} else {
		q.OutQueue <- d
		return nil
	}
}

/*
*
* Push
*
 */
func (q *DataCacheQueue) pushDevice(d QueueData) error {
	// 动态扩容
	// if len(q.Queue)+1 > q.GetSize() {
	// }
	if len(q.DeviceQueue)+1 > q.GetSize() {
		msg := fmt.Sprintf("attached max queue size, max size is:%v, current size is: %v",
			q.GetSize(), len(q.DeviceQueue)+1)
		glogger.GLogger.Error(msg)
		return errors.New(msg)
	} else {
		q.DeviceQueue <- d
		return nil
	}
}
