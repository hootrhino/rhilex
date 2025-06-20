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
package source

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/hootrhino/rhilex/component/protocol"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)

type CustomProtocolConfig struct {
	Host          string `json:"host" validate:"required"`
	Port          int    `json:"port" validate:"required"`
	Timeout       int    `json:"timeout" validate:"required"`
	ProtocolExpr  string `json:"protocolExpr" validate:"required"`  // 数据解析表达式
	MaxDataLength int    `json:"maxDataLength" validate:"required"` // 最长数据1024
}
type CustomProtocol struct {
	typex.XStatus
	mainConfig CustomProtocolConfig
	status     typex.SourceState
	Listener   *net.TCPListener
}

func NewCustomProtocol(e typex.Rhilex) typex.XSource {
	h := CustomProtocol{
		mainConfig: CustomProtocolConfig{
			ProtocolExpr:  "",
			MaxDataLength: 1024,
			Host:          "127.0.0.1",
			Port:          7930,
			Timeout:       3000,
		},
	}
	h.RuleEngine = e
	return &h
}

func (hh *CustomProtocol) Init(inEndId string, configMap map[string]any) error {
	hh.PointId = inEndId
	if err := utils.BindSourceConfig(configMap, &hh.mainConfig); err != nil {
		return err
	}
	if hh.mainConfig.MaxDataLength < 1 {
		return fmt.Errorf("Invalid Max Data Length:%d", hh.mainConfig.MaxDataLength)
	}
	if hh.mainConfig.MaxDataLength > 1024 {
		return fmt.Errorf("Invalid Max Data Length:%d", hh.mainConfig.MaxDataLength)
	}
	if err := validateExpression(hh.mainConfig.ProtocolExpr); err != nil {
		return fmt.Errorf("Invalid Protocol Expression:%s", err)
	}
	return nil
}

func (hh *CustomProtocol) Start(cctx typex.CCTX) error {
	hh.Ctx = cctx.Ctx
	hh.CancelCTX = cctx.CancelCTX
	Listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d",
		hh.mainConfig.Host, hh.mainConfig.Port))
	if err != nil {
		glogger.GLogger.Errorf("Failed to listen on %s:%d: %v", hh.mainConfig.Host, hh.mainConfig.Port, err)
		return err
	}
	hh.Listener = Listener.(*net.TCPListener)

	go func(ctx context.Context, Listener *net.TCPListener) {
		defer func() {
			if Listener != nil {
				Listener.Close()
			}
		}()

		maxRetries := 10
		retryDelay := 5 * time.Second
		for {
			select {
			case <-ctx.Done():
				return
			default:
				conn, err := Listener.Accept()
				if err != nil {
					if maxRetries > 0 {
						glogger.GLogger.Errorf("Error accepting connection: %v, retries left: %d", err, maxRetries)
						maxRetries--
						time.Sleep(retryDelay)
						continue
					} else {
						glogger.GLogger.Errorf("Max retries reached, giving up: %v", err)
						return
					}
				}
				maxRetries = 10 // Reset retries on successful connection
				glogger.GLogger.Info("Accepting connection:", conn.RemoteAddr())
				config := protocol.ExchangeConfig{
					Port:         conn,
					ReadTimeout:  100,
					WriteTimeout: 0,
					Logger:       glogger.Logrus,
				}
				ctx, cancel := context.WithCancel(hh.Ctx)
				TransportSlaver := protocol.NewGenericProtocolSlaver(ctx, cancel, config)
				TransportSlaver.StartLoop(func(ApplicationFrame *protocol.ApplicationFrame, err error) {
					if err != nil {
						glogger.GLogger.Error(err)
						return
					}
					ParsedData, errParse := utils.ParseBinary(hh.mainConfig.ProtocolExpr, ApplicationFrame.Payload)
					if errParse != nil {
						glogger.GLogger.Error(errParse)
						return
					}
					ClientDataBytes, _ := json.Marshal(ParsedData)
					if len(ClientDataBytes) > 2 {
						hh.RuleEngine.WorkInEnd(hh.Details(), string(ClientDataBytes))
					}
				})
			}
		}
	}(hh.Ctx, hh.Listener)
	hh.status = typex.SOURCE_UP
	return nil
}

func (hh *CustomProtocol) Stop() {
	hh.status = typex.SOURCE_DOWN
	if hh.CancelCTX != nil {
		hh.CancelCTX()
	}
	if hh.Listener != nil {
		hh.Listener.Close()
	}
}

func (hh *CustomProtocol) Status() typex.SourceState {
	return hh.status
}

func (hh *CustomProtocol) Details() *typex.InEnd {
	return hh.RuleEngine.GetInEnd(hh.PointId)
}

/**
 * 验证格式
 *
 */
func validateExpression(expression string) error {
	fieldPattern := regexp.MustCompile(`(\w+):(\d+):(I|S|F|int|string|float):(B|L|BE|LE);`)
	if !fieldPattern.Match([]byte(expression)) {
		return fmt.Errorf("Invalid expression syntax")
	}
	line := strings.Split(expression, ";")
	for _, fields := range line {
		if fields == "" {
			continue
		}
		fieldPair := strings.Split(fields, ":")
		if len(fieldPair) != 4 {
			return fmt.Errorf("invalid fields length")
		}
		if len(fieldPair[0]) > 64 {
			return fmt.Errorf("filed key too large")
		}
		if len(fieldPair[1]) > 6 {
			return fmt.Errorf("filed value too large")
		}
		if !utils.SContains([]string{"I", "S", "F", "int", "string", "float"}, fieldPair[2]) {
			return fmt.Errorf("invalid field type")
		}
		if !utils.SContains([]string{"B", "L", "BE", "LE"}, fieldPair[3]) {
			return fmt.Errorf("invalid field Endian")
		}
	}
	return nil
}
