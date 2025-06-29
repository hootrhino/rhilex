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

/*
*
* 主从模式
*
 */
type MBacnetDataPoint struct {
	RhilexModel
	UUID           string  `gorm:"not null"`
	DeviceUuid     string  `gorm:"not null"`
	Tag            string  `gorm:"not null"`
	Alias          string  `gorm:"not null"`
	BacnetDeviceId uint32  `gorm:"not null"`
	ObjectType     string  `gorm:"not null"`
	ObjectId       uint32  `gorm:"not null"`
	Frequency      *uint64 `gorm:"default:50"`
}

/*
*
* 路由模式
*
 */
type MBacnetRouterDataPoint struct {
	RhilexModel
	UUID       string `gorm:"not null"`
	DeviceUuid string `gorm:"not null"`
	Tag        string `gorm:"not null"`
	Alias      string `gorm:"not null"`
	ObjectId   uint32 `gorm:"not null"`
	ObjectType string `gorm:"not null"`
}
