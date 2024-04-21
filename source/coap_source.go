package source

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hootrhino/rhilex/common"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
	coap "github.com/plgd-dev/go-coap/v3"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
)

type coAPInEndSource struct {
	typex.XStatus
	router     *mux.Router
	mainConfig common.HostConfig
	status     typex.SourceState
}

func NewCoAPInEndSource(e typex.Rhilex) typex.XSource {
	c := coAPInEndSource{}
	c.router = mux.NewRouter()
	c.mainConfig = common.HostConfig{}
	c.RuleEngine = e
	return &c
}

func (cc *coAPInEndSource) Init(inEndId string, configMap map[string]interface{}) error {
	cc.PointId = inEndId
	if err := utils.BindSourceConfig(configMap, &cc.mainConfig); err != nil {
		return err
	}

	return nil
}

// 输入数据
type InData struct {
	Ts      uint64 `json:"ts"`
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

func (cc *coAPInEndSource) Start(cctx typex.CCTX) error {
	cc.Ctx = cctx.Ctx
	cc.CancelCTX = cctx.CancelCTX
	port := fmt.Sprintf(":%v", cc.mainConfig.Port)
	cc.router.Handle("/", mux.HandlerFunc(func(w mux.ResponseWriter, msg *mux.Message) {
		Body, err := msg.ReadBody()
		if err != nil {
			glogger.GLogger.Error(err)
			return
		}
		// glogger.GLogger.Debug(msg.RouteParams.Vars, "; ", string(Body), "; ", w.Conn().RemoteAddr())
		Payload := InData{
			Ts:      uint64(time.Now().UnixMilli()),
			Type:    "POST",
			Payload: string(Body),
		}
		if bites, err := json.Marshal(Payload); err != nil {
			glogger.GLogger.Error(err)
		} else {
			glogger.GLogger.Debug(string(bites))
			cc.RuleEngine.WorkInEnd(cc.RuleEngine.GetInEnd(cc.PointId), string(bites))
		}
		if err := w.SetResponse(codes.Content, message.AppOctets,
			bytes.NewReader([]byte{200})); err != nil {
			glogger.GLogger.Errorf("Cannot set response: %v", err)
		}
	}))

	go func(ctx context.Context) {
		err := coap.ListenAndServe("udp", port, cc.router)
		if err != nil {
			glogger.GLogger.Error(err)
			return
		}
	}(cc.Ctx)
	cc.status = typex.SOURCE_UP
	glogger.GLogger.Info("Coap source started on [udp]" + port)
	return nil
}

func (cc *coAPInEndSource) Stop() {
	cc.status = typex.SOURCE_DOWN
	if cc.CancelCTX != nil {
		cc.CancelCTX()
	}
}

func (cc *coAPInEndSource) DataModels() []typex.XDataModel {
	return []typex.XDataModel{}
}

func (cc *coAPInEndSource) Status() typex.SourceState {
	return cc.status
}

func (cc *coAPInEndSource) Test(inEndId string) bool {
	return true
}

func (cc *coAPInEndSource) Details() *typex.InEnd {
	return cc.RuleEngine.GetInEnd(cc.PointId)
}

func (*coAPInEndSource) DownStream([]byte) (int, error) {
	return 0, nil
}

func (*coAPInEndSource) UpStream([]byte) (int, error) {
	return 0, nil
}
