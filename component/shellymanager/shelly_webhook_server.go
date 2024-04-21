// Copyright (C) 2024 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package shellymanager

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gin-contrib/graceful"
	"github.com/gin-gonic/gin"
	core "github.com/hootrhino/rhilex/config"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
)

type ShellyWebHookServer struct {
	ctx            context.Context
	webServer      *graceful.Graceful
	port           int // port
	NotifyCallback func(Notify ShellyDeviceNotify)
}

/*
*
* webhook server
*
 */
func NewShellyWebHookServer(e typex.Rhilex, port int) *ShellyWebHookServer {
	if core.GlobalConfig.AppDebugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	return &ShellyWebHookServer{
		port: port,
	}
}

/*
*
* 启动服务
*
 */
func (WebHookServer *ShellyWebHookServer) StartServer(ctx context.Context) {
	WebHookServer.ctx = ctx
	webServer, err := graceful.New(gin.New(),
		graceful.WithAddr(fmt.Sprintf(":%v", WebHookServer.port)))
	if err != nil {
		glogger.GLogger.Fatal(err)
		os.Exit(-1)
	}
	WebHookServer.webServer = webServer
	WebHookServer.webServer.POST("/", WebHookServer.CallBackApi)
	WebHookServer.webServer = webServer
	go func(ctx context.Context) {
		err := WebHookServer.webServer.RunWithContext(WebHookServer.ctx)
		if err != nil {
			glogger.GLogger.Error(err)
			return
		}
	}(context.Background())
	glogger.GLogger.Infof("Shelly Web Hook Server started on [0.0.0.0:%v]", WebHookServer.port)
}
func (WebHookServer *ShellyWebHookServer) SetEventCallBack(NotifyCallback func(Notify ShellyDeviceNotify)) {
	WebHookServer.NotifyCallback = NotifyCallback
}
func (WebHookServer *ShellyWebHookServer) Stop() {
	WebHookServer.webServer.Shutdown(WebHookServer.ctx)
}
func (WebHookServer *ShellyWebHookServer) CallBackApi(ctx *gin.Context) {
	ShellyDeviceEvent := ShellyDeviceEvent{}
	if err := ctx.ShouldBindJSON(&ShellyDeviceEvent); err != nil {
		fmt.Println(err)
		ctx.JSON(400, map[string]any{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}
	if WebHookServer.NotifyCallback != nil {
		WebHookServer.NotifyCallback(ShellyDeviceNotify{})
	}
	ctx.JSON(200, map[string]any{
		"code": 200,
		"msg":  "success",
	})
}

/*
*

  - WebHook产生的事件
    https://shelly-api-docs.shelly.cloud/gen2/ComponentsAndServices/Input#webhook-events

    for Input instances of type switch:
    input.toggle_on - produced when the input instance is toggled from off to on state
    input.toggle_off - produced when the input instance is toggled from on to off state

*
*/
type ShellyDeviceEvent struct {
	Src    string `json:"src"`
	Dst    string `json:"dst"`
	Method string `json:"method"`
	Params struct {
		Ts     float64 `json:"ts"`
		Events []struct {
			Component string  `json:"component"`
			ID        int     `json:"id"`
			Event     string  `json:"event"`
			Ts        float64 `json:"ts"`
		} `json:"events"`
	} `json:"params"`
}

func (E ShellyDeviceEvent) String() string {
	if bytes, err := json.Marshal(E); err != nil {
		return "{}"
	} else {
		return string(bytes)
	}
}

type ShellyDeviceNotify struct {
	Mac   string            `json:"mac"`
	IP    string            `json:"ip"`
	Event ShellyDeviceEvent `json:"event"`
}

func (E ShellyDeviceNotify) String() string {
	if bytes, err := json.Marshal(E); err != nil {
		return "{}"
	} else {
		return string(bytes)
	}
}
