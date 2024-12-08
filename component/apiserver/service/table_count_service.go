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

package service

import "github.com/hootrhino/rhilex/component/interdb"

/*
*
* 计算 Count
*
 */
func CountModel(m any) int64 {
	var count int64
	interdb.InterDb().Model(m).Count(&count)
	return count
}

/*
*
* 计算 Count
*
 */
func CountTable(table string) int64 {
	var count int64
	interdb.InterDb().Table(table).Count(&count)
	return count
}
