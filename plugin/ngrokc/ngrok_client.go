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
	"github.com/hootrhino/rhilex/typex"
	"gopkg.in/ini.v1"
)

type NgrokClient struct {
	uuid string
}

func NewNgrokClient() *NgrokClient {
	return &NgrokClient{
		uuid: "NGROK",
	}
}

func (dm *NgrokClient) Init(config *ini.Section) error {
	return nil
}

func (dm *NgrokClient) Start(typex.Rhilex) error {
	return nil
}
func (dm *NgrokClient) Stop() error {
	return nil
}

func (dm *NgrokClient) PluginMetaInfo() typex.XPluginMetaInfo {
	return typex.XPluginMetaInfo{
		UUID:     dm.uuid,
		Name:     "NgrokClient",
		Version:  "v0.0.1",
		Homepage: "/",
		HelpLink: "/",
		Author:   "RHILEXTeam",
		Email:    "RHILEXTeam@hootrhino.com",
		License:  "AGPL",
	}
}

/*
*
* 服务调用接口
*
 */
func (dm *NgrokClient) Service(arg typex.ServiceArg) typex.ServiceResult {
	return typex.ServiceResult{}
}
