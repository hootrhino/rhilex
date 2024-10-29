package device

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)

/*
*
* HTTP
*
 */
type __HttpConfig struct {
	Url     string            `json:"url" validate:"required" title:"URL"`
	Headers map[string]string `json:"headers" validate:"required" title:"HTTP Headers"`
}
type __HttpCommonConfig struct {
	Timeout     *int   `json:"timeout" validate:"required"`
	AutoRequest *bool  `json:"autoRequest" validate:"required"`
	Frequency   *int64 `json:"frequency" validate:"required"`
}
type __HttpMainConfig struct {
	CommonConfig __HttpCommonConfig `json:"commonConfig" validate:"required"`
	HttpConfig   __HttpConfig       `json:"httpConfig" validate:"required"`
}

type GenericHttpDevice struct {
	typex.XStatus
	client     http.Client
	status     typex.DeviceState
	RuleEngine typex.Rhilex
	mainConfig __HttpMainConfig
	locker     sync.Locker
}

/*
*
* 通用串口透传
*
 */
func NewGenericHttpDevice(e typex.Rhilex) typex.XDevice {
	hd := new(GenericHttpDevice)
	hd.locker = &sync.Mutex{}
	hd.client = *http.DefaultClient
	hd.mainConfig = __HttpMainConfig{
		CommonConfig: __HttpCommonConfig{
			AutoRequest: func() *bool {
				b := false
				return &b
			}(),
			Timeout: func() *int {
				b := 3000
				return &b
			}(),
		},
	}
	hd.RuleEngine = e
	return hd
}

//  初始化
func (hd *GenericHttpDevice) Init(devId string, configMap map[string]interface{}) error {
	hd.PointId = devId
	if err := utils.BindSourceConfig(configMap, &hd.mainConfig); err != nil {
		glogger.GLogger.Error(err)
		return err
	}
	if _, err := isValidHTTP_URL(hd.mainConfig.HttpConfig.Url); err != nil {
		return fmt.Errorf("Invalid url format:%s, %s", hd.mainConfig.HttpConfig.Url, err)
	}
	return nil
}

// 启动
func (hd *GenericHttpDevice) Start(cctx typex.CCTX) error {
	hd.Ctx = cctx.Ctx
	hd.CancelCTX = cctx.CancelCTX

	if *hd.mainConfig.CommonConfig.AutoRequest {
		go func() {
			ticker := time.NewTicker(time.Duration(*hd.mainConfig.CommonConfig.Frequency) * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-hd.Ctx.Done():
					{
						return
					}
				default:
					{
					}
				}
				body := httpGet(hd.client, hd.mainConfig.HttpConfig.Url)
				if body != "" {
					hd.RuleEngine.WorkDevice(hd.Details(), body)
				}
				<-ticker.C
			}
		}()

	}
	hd.status = typex.DEV_UP
	return nil
}

func (hd *GenericHttpDevice) OnRead(cmd []byte, data []byte) (int, error) {

	return 0, nil
}

// 把数据写入设备
func (hd *GenericHttpDevice) OnWrite(cmd []byte, b []byte) (int, error) {
	return 0, nil
}

// 设备当前状态
func (hd *GenericHttpDevice) Status() typex.DeviceState {
	return hd.status
}

// 停止设备
func (hd *GenericHttpDevice) Stop() {
	hd.status = typex.DEV_DOWN
	if hd.CancelCTX != nil {
		hd.CancelCTX()
	}
}

// 真实设备
func (hd *GenericHttpDevice) Details() *typex.Device {
	return hd.RuleEngine.GetDevice(hd.PointId)
}

// 状态
func (hd *GenericHttpDevice) SetState(status typex.DeviceState) {
	hd.status = status

}

// --------------------------------------------------------------------------------------------------
//
// --------------------------------------------------------------------------------------------------

func (hd *GenericHttpDevice) OnDCACall(UUID string, Command string, Args interface{}) typex.DCAResult {
	return typex.DCAResult{}
}
func (hd *GenericHttpDevice) OnCtrl(cmd []byte, args []byte) ([]byte, error) {
	return []byte{}, nil
}

/*
*
* HTTP GET
*
 */
func httpGet(client http.Client, url string) string {
	var err error
	client.Timeout = 2 * time.Second
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		glogger.GLogger.Warn(err)
		return ""
	}

	response, err := client.Do(request)
	if err != nil {
		glogger.GLogger.Warn(err)
		return ""
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		glogger.GLogger.Warn(err)
		return ""
	}
	return string(body)
}

/*
*
* 验证URL语法
*
 */
func isValidHTTP_URL(urlStr string) (bool, error) {
	r, err := url.Parse(urlStr)
	if err != nil {
		return false, fmt.Errorf("error parsing URL: %w", err)
	}
	if r.Scheme != "http" && r.Scheme != "https" {
		return false, fmt.Errorf("Invalid scheme; must be http or https")
	}
	return true, nil
}
