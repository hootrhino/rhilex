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

import "github.com/hootrhino/rhilex/typex"

/*
*
* 服务调用接口
*
 */
type NgrokResponse struct {
	ServerAddr     string `json:"server_addr"`
	ServerEndpoint string `json:"server_endpoint"`
	Domain         string `json:"domain"`
	LocalSchema    string `json:"local_schema"`
	LocalHost      string `json:"local_host"`
	LocalPort      int    `json:"local_port"`
	AuthToken      string `json:"auth_token"`
}

func (dm *NgrokClient) Service(arg typex.ServiceArg) typex.ServiceResult {
	if arg.Name == "start" {
		if err := dm.startClient(); err != nil {
			return typex.ServiceResult{Out: err}
		}
		dm.busy = true
		return typex.ServiceResult{Out: "Ngrok Client Started"}
	}
	if arg.Name == "stop" {
		dm.cancel()
		if err := dm.forwarder.Close(); err != nil {
			return typex.ServiceResult{Out: err}
		}
		dm.busy = false
		return typex.ServiceResult{Out: "Ngrok Client Stopped"}
	}
	if arg.Name == "get_config" {
		return typex.ServiceResult{Out: NgrokResponse{
			ServerAddr:     dm.serverAddr,
			Domain:         dm.mainConfig.Domain,
			ServerEndpoint: dm.mainConfig.ServerEndpoint,
			LocalSchema:    dm.mainConfig.LocalSchema,
			LocalHost:      dm.mainConfig.LocalHost,
			LocalPort:      dm.mainConfig.LocalPort,
			AuthToken:      dm.mainConfig.AuthToken,
		}}
	}
	return typex.ServiceResult{Out: "Unsupported command:" + arg.Name}
}
