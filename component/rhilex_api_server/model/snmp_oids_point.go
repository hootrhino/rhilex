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

package model

// SNMP 点位表
type MSnmpOid struct {
	RhilexModel
	UUID       string
	DeviceUuid string `gorm:"not null"` // 所属设备
	Oid        string `gorm:"not null"` // .1.3.6.1.2.1.25.1.6.0
	Tag        string `gorm:"not null"` // temp
	Alias      string `gorm:"not null"` // 温度
	Frequency  int    `gorm:"not null"` // 请求频率
}
