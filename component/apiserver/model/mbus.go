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

// 点位
type MMBusDataPoint struct {
	RhilexModel
	UUID         string  `gorm:"not null"`
	DeviceUuid   string  `gorm:"not null"`
	SlaverId     string  `gorm:"not null"`
	Type         string  `gorm:"not null"`
	Manufacturer string  `gorm:"not null"`
	Tag          string  `gorm:"not null"`
	Alias        string  `gorm:"not null"`
	Frequency    uint64 `gorm:"not null"`
	DataLength   uint64 `gorm:"not null"`
}

func (MMBusDataPoint) TableName() string {
	return "m_mbus_data_points"
}
