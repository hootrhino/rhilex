// Copyright (C) 2023 wwhai
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
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package modbus_crc_tools

import (
	"github.com/hootrhino/rhilex/typex"
	"gopkg.in/ini.v1"
)

type modbusCRCCalculator struct {
	uuid string
}

func NewModbusCrcCalculator() typex.XPlugin {
	return &modbusCRCCalculator{
		uuid: "MODBUS_CRC_CALCULATOR",
	}
}

func (ms *modbusCRCCalculator) Init(config *ini.Section) error {
	return nil
}

func (ms *modbusCRCCalculator) Start(typex.Rhilex) error {
	return nil
}
func (ms *modbusCRCCalculator) Stop() error {
	return nil
}

func (ms *modbusCRCCalculator) PluginMetaInfo() typex.XPluginMetaInfo {
	return typex.XPluginMetaInfo{
		UUID:        ms.uuid,
		Name:        "CRC Calculator",
		Version:     "v0.0.1",
		Description: "Simple CRC Calculator",
	}
}
