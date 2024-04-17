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

package rhilexlib

import (
	"encoding/json"

	lua "github.com/hootrhino/gopher-lua"
	"github.com/hootrhino/rhilex/component/datacenter"
	"github.com/hootrhino/rhilex/typex"
)

/*
*
* 数据中心本地执行
*
 */
func LocalDBQuery(rx typex.Rhilex) func(*lua.LState) int {
	return func(l *lua.LState) int {
		sql := l.ToString(2)
		Map, err := datacenter.Query("rhilex_internal_datacenter", sql)
		if err != nil {
			l.Push(lua.LNil)
			l.Push(lua.LString(err.Error()))
			return 2
		}
		bytes, err1 := json.Marshal(Map)
		if err1 != nil {
			l.Push(lua.LNil)
			l.Push(lua.LString(err1.Error()))
			return 2
		}
		l.Push(lua.LString(string(bytes)))
		l.Push(lua.LNil)
		return 2
	}
}
