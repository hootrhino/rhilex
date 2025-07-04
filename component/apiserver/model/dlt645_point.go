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

type MDlt6452007DataPoint struct {
	RhilexModel
	UUID       string
	DeviceUuid string
	MeterId    string
	Tag        string
	Alias      string
	Frequency  *uint64  `gorm:"default:50"`
	Weight     *Decimal `gorm:"column:weight;default:1"`
}
