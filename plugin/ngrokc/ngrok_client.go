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

package ngrokc

import (
	"fmt"
	"net/url"
	"time"

	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
	"golang.org/x/exp/slices"
	"gopkg.in/ini.v1"

	"context"

	ngrok_log "golang.ngrok.com/ngrok/log"

	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
)

// Simple logger that forwards to the Go standard logger.
type logger struct {
}

func (l *logger) Log(ctx context.Context, lvl ngrok_log.LogLevel, msg string, data map[string]interface{}) {
	glogger.GLogger.Debugf("%s , %v", msg, data)
}

type NgrokConfig struct {
	Enable         bool   `ini:"enable" json:"enable"`
	ServerEndpoint string `ini:"server_endpoint" json:"server_endpoint"`
	Domain         string `ini:"domain" json:"domain"`
	LocalSchema    string `ini:"local_schema" json:"local_schema"`
	LocalHost      string `ini:"local_host" json:"local_host"`
	LocalPort      int    `ini:"local_port" json:"local_port"`
	AuthToken      string `ini:"auth_token" json:"auth_token"`
}
type NgrokClient struct {
	ctx        context.Context
	cancel     context.CancelFunc
	busy       bool
	forwarder  ngrok.Forwarder
	serverAddr string
	mainConfig NgrokConfig
}

func NewNgrokClient() *NgrokClient {
	return &NgrokClient{
		busy:       false,
		serverAddr: "",
		mainConfig: NgrokConfig{
			AuthToken:      "",
			Domain:         "default",
			ServerEndpoint: "default",
			LocalSchema:    "http",
			LocalHost:      "127.0.0.1",
			LocalPort:      2580,
		},
	}
}

func (dm *NgrokClient) Init(config *ini.Section) error {
	if err := utils.InIMapToStruct(config, &dm.mainConfig); err != nil {
		return err
	}
	if dm.mainConfig.AuthToken == "default" {
		return fmt.Errorf("invalid ngrok auth token, More detail go to: https://ngrok.com/docs/getting-started")
	}
	if !slices.Contains([]string{"tcp", "http", "https"}, dm.mainConfig.LocalSchema) {
		return fmt.Errorf("LocalSchema must one of tcp or http or https")
	}
	return nil
}
func (dm *NgrokClient) getTunnel() config.Tunnel {
	if dm.mainConfig.Domain == "" {
		return config.HTTPEndpoint()
	} else {
		return config.HTTPEndpoint(config.WithDomain(dm.mainConfig.Domain))
	}
}
func (dm *NgrokClient) startClient() error {
	URL, err := url.Parse(fmt.Sprintf("%s://%s:%v",
		dm.mainConfig.LocalSchema, dm.mainConfig.LocalHost, dm.mainConfig.LocalPort))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(typex.GCTX)
	dm.cancel = cancel
	dm.ctx = ctx
	if dm.mainConfig.LocalSchema == "tcp" {
		Forwarder, err := ngrok.ListenAndForward(ctx, URL, config.TCPEndpoint(),
			ngrok.WithAuthtoken(dm.mainConfig.AuthToken),
			ngrok.WithConnectHandler(func(ctx context.Context, session ngrok.Session) {
				glogger.GLogger.Debug(session.Warnings())
			}),
			ngrok.WithHeartbeatHandler(func(ctx context.Context, session ngrok.Session, latency time.Duration) {
				glogger.GLogger.Debug(latency)
			}),
			ngrok.WithDisconnectHandler(func(ctx context.Context, session ngrok.Session, err error) {
				glogger.GLogger.Error(err)
			}),
			ngrok.WithLogger(&logger{}))
		if err != nil {
			return err
		}
		dm.serverAddr = Forwarder.URL()
		glogger.GLogger.Debugf("Forwarder: %s connect to Ngrok success: %s",
			Forwarder.ID(), Forwarder.URL())
		dm.forwarder = Forwarder
		return nil
	}
	// workable-logically-tarpon.ngrok-free.app
	if dm.mainConfig.LocalSchema == "http" {
		Forwarder, err := ngrok.ListenAndForward(ctx, URL, dm.getTunnel(),
			ngrok.WithAuthtoken(dm.mainConfig.AuthToken),
			ngrok.WithLogger(&logger{}))
		if err != nil {
			return err
		}
		dm.serverAddr = Forwarder.URL()
		glogger.GLogger.Debugf("Forwarder: %s connect to Ngrok success: %s",
			Forwarder.ID(), Forwarder.URL())
		dm.forwarder = Forwarder
		return nil

	}
	if dm.mainConfig.LocalSchema == "https" {
		Forwarder, err := ngrok.ListenAndForward(ctx, URL, dm.getTunnel(),
			ngrok.WithAuthtoken(dm.mainConfig.AuthToken),
			ngrok.WithLogger(&logger{}))
		if err != nil {
			return err
		}
		dm.serverAddr = Forwarder.URL()
		glogger.GLogger.Debugf("Forwarder: %s connect to Ngrok success: %s",
			Forwarder.ID(), Forwarder.URL())
		dm.forwarder = Forwarder
		return nil
	}
	return fmt.Errorf("unsupported schema:%s", dm.mainConfig.LocalSchema)
}

// "2dInwP3b8reiSrKTcVnlreCOU1b_5t9z3J7spaF4WwRF8o8gM"
func (dm *NgrokClient) Start(typex.Rhilex) error {
	if err := dm.startClient(); err != nil {
		return err
	}
	dm.busy = true
	return nil
}
func (dm *NgrokClient) Stop() error {
	if dm.cancel != nil {
		dm.cancel()
	}
	if dm.forwarder != nil {
		dm.forwarder.Close()
		dm.forwarder = nil
	}
	dm.busy = false
	return nil
}

func (dm *NgrokClient) PluginMetaInfo() typex.XPluginMetaInfo {
	return typex.XPluginMetaInfo{
		UUID:     "NGROKC",
		Name:     "Ngrok Client",
		Version:  "v0.0.1",
		Homepage: "/",
		HelpLink: "/",
		Author:   "RHILEXTeam",
		Email:    "RHILEXTeam@hootrhino.com",
		License:  "",
	}
}
