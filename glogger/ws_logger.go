// Copyright (C) 2025 wwhai
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

package glogger

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var private_GRealtimeLogger *RealTimeLogger

type wsLogHook struct {
	levels []logrus.Level
}

func NewWSLogHook(ss string) wsLogHook {
	return wsLogHook{levels: level(ss)}
}
func (hk wsLogHook) Levels() []logrus.Level {
	return hk.levels
}
func (hk wsLogHook) Fire(e *logrus.Entry) error {
	msg, _ := e.String()
	private_GRealtimeLogger.Write([]byte(msg))
	return nil
}

func level(ss string) []logrus.Level {
	switch ss {
	case "fatal":
		return []logrus.Level{logrus.FatalLevel}
	case "error":
		return []logrus.Level{logrus.ErrorLevel}
	case "warn":
		return []logrus.Level{logrus.WarnLevel}
	case "debug":
		return []logrus.Level{logrus.DebugLevel}
	case "info":
		return []logrus.Level{logrus.InfoLevel}
	case "all", "trace":
		return []logrus.Level{
			logrus.TraceLevel,
			logrus.FatalLevel,
			logrus.WarnLevel,
			logrus.DebugLevel,
			logrus.InfoLevel,
			logrus.TraceLevel,
		}
	}
	return []logrus.Level{logrus.InfoLevel}
}

/*
*
* 这是用来给外部输出日志的websocket服务器，其功能非常简单，就是单纯的对外输出实时日志，方便调试使用。
* 注意：该功能需要配合HttpApiServer使用, 客户端连上以后必须在5s内发送一个 ‘WsTerminal’ 的固定字符
*       串到服务器才能过认证。
*
 */
type RealTimeLogger struct {
	WsServer websocket.Upgrader
	Clients  map[string]*websocket.Conn
	locker   sync.Mutex
}

func (w *RealTimeLogger) Write(p []byte) (n int, err error) {
	w.locker.Lock()
	defer w.locker.Unlock()
	for _, c := range w.Clients {
		err := c.WriteMessage(websocket.TextMessage, p)
		if err != nil {
			return 0, err
		}
	}
	return 0, nil
}

func StartNewRealTimeLogger(s string) *RealTimeLogger {
	private_GRealtimeLogger = &RealTimeLogger{
		WsServer: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		Clients: make(map[string]*websocket.Conn),
		locker:  sync.Mutex{},
	}
	Logrus.AddHook(NewWSLogHook(s))
	return private_GRealtimeLogger
}

/*
*
* 启动服务
*
 */

func WsLogger(c *gin.Context) {
	//upgrade get request to websocket protocol
	wsConn, err := private_GRealtimeLogger.WsServer.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	// 首先读第一个包是不是: WsTerminal
	wsConn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, b, err := wsConn.ReadMessage()
	if err != nil {
		wsConn.WriteMessage(websocket.CloseMessage, []byte("Invalid client token"))
		wsConn.Close()
		return
	}
	wsConn.SetReadDeadline(time.Time{})
	token := string(b)
	if token != "WsTerminal" {
		private_GRealtimeLogger.locker.Lock()
		wsConn.WriteMessage(1, []byte("Invalid client token"))
		private_GRealtimeLogger.locker.Unlock()
		wsConn.Close()
		return
	}
	// 最多允许连接10个客户端，实际情况下根本用不了那么多
	if len(private_GRealtimeLogger.Clients) > 5 {
		private_GRealtimeLogger.locker.Lock()
		wsConn.WriteMessage(websocket.CloseMessage, []byte{})
		private_GRealtimeLogger.locker.Unlock()
		wsConn.Close()
		return
	}
	private_GRealtimeLogger.locker.Lock()
	private_GRealtimeLogger.Clients[wsConn.RemoteAddr().String()] = wsConn
	wsConn.WriteMessage(websocket.TextMessage, []byte("Connected"))
	private_GRealtimeLogger.locker.Unlock()
	GLogger.Info("WebSocket Terminal connected:" + wsConn.RemoteAddr().String())
	wsConn.SetCloseHandler(func(code int, text string) error {
		GLogger.Info("wsConn Auto Close:", wsConn.RemoteAddr().String())
		private_GRealtimeLogger.locker.Lock()
		delete(private_GRealtimeLogger.Clients, wsConn.RemoteAddr().String())
		private_GRealtimeLogger.locker.Unlock()
		return nil
	})
	wsConn.SetPingHandler(func(appData string) error {
		return nil
	})
	wsConn.SetPongHandler(func(appData string) error {
		return nil
	})
	go func(wsConn *websocket.Conn) {
		defer func() {
			if wsConn != nil {
				GLogger.Info("wsConn Disconnect By accident:", wsConn.RemoteAddr().String())
				private_GRealtimeLogger.locker.Lock()
				delete(private_GRealtimeLogger.Clients, wsConn.RemoteAddr().String())
				private_GRealtimeLogger.locker.Unlock()
			}
		}()
		for {
			select {
			case <-context.Background().Done():
				{
					return
				}
			default:
				{
				}
			}
			_, _, err := wsConn.ReadMessage()
			if err != nil {
				GLogger.Error("GRealtimeLogger error:", wsConn.RemoteAddr().String(), ", Error:", err)
				wsConn.Close()
				private_GRealtimeLogger.locker.Lock()
				delete(private_GRealtimeLogger.Clients, wsConn.RemoteAddr().String())
				private_GRealtimeLogger.locker.Unlock()
				return
			}
			private_GRealtimeLogger.locker.Lock()
			err = wsConn.WriteMessage(websocket.PingMessage, []byte{})
			private_GRealtimeLogger.locker.Unlock()

			if err != nil {
				break
			}
			time.Sleep(5 * time.Second)
		}
	}(wsConn)
}
