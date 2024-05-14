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
package rhilexlib

import (
	"github.com/hootrhino/rhilex/typex"

	lua "github.com/hootrhino/gopher-lua"
)

/*
*
* 改变模型值
*
 */
func SetModelValue(rx typex.Rhilex, uuid string) func(L *lua.LState) int {
	return func(l *lua.LState) int {
		uuid := l.ToString(2)
		name := l.ToString(3)
		value := l.ToString(4)
		setValue(rx, uuid, name, value)
		return 0
	}
}

/*
*
* 改变值
*
 */
func setValue(rx typex.Rhilex, uuid, name, value string) {

	in := rx.GetInEnd(uuid)
	if in != nil {
		DataModel := in.DataModelsMap[name]
		DataModel.Value = value
		in.DataModelsMap[name] = DataModel
	}
}
