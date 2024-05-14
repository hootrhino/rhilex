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
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"

	lua "github.com/hootrhino/gopher-lua"
)

/*
*
  - 改变模型值: data:ToRDS('uuid',{k=v}.....)
    local err = data:ToRDS('SCHEMAHHCOOYVY', {
    temp = 25.67,
    humi = 67.89,
    sw1 = true,
    warning = "Low Battery"
    })
*/
func InsertToDataCenterTable(rx typex.Rhilex, uuid string) func(L *lua.LState) int {
	return func(l *lua.LState) int {
		schema_uuid := l.ToString(2)
		kvs := l.ToTable(3)
		Row := map[string]interface{}{}
		kvs.ForEach(func(k, v lua.LValue) {
			// K 只能String
			if k.Type() == lua.LTString {
				switch v.Type() {
				case lua.LTString:
					Row[lua.LVAsString(k)] = lua.LVAsString(v)
				case lua.LTNumber:
					Row[lua.LVAsString(k)] = lua.LVAsString(v)
				case lua.LTBool:
					Row[lua.LVAsString(k)] = lua.LVAsBool(v)
				case lua.LTNil:
					Row[lua.LVAsString(k)] = lua.LVAsString(v)
				default:
					Row[lua.LVAsString(k)] = lua.LNil // 不支持其他类型
				}
			}
		})
		if err := saveToDataCenter(schema_uuid, Row); err != nil {
			l.Push(lua.LString(err.Error()))
		} else {
			l.Push(lua.LNil)
		}
		return 1
	}
}

// Save to local DataCenter
func saveToDataCenter(schema_uuid string, row map[string]interface{}) error {
	glogger.GLogger.Debug(schema_uuid, row)
	return nil
}

// Query List
// GET /api/v1/datacenter/queryDataList?uuid=<uuid>&current=<current>&size=<size>&order=<order>&select=<select> HTTP/1.1
// Host: 127.0.0.1:2580
func QueryDataCenterList(rx typex.Rhilex, uuid string) func(L *lua.LState) int {
	return func(l *lua.LState) int {
		schema_uuid := l.ToString(2)
		page := l.ToNumber(3)
		size := l.ToNumber(4)
		fields := l.ToString(5)
		if err := queryDataCenterList(schema_uuid, int(page), int(size), fields); err != nil {
			l.Push(lua.LString(err.Error()))
		} else {
			l.Push(lua.LNil)
		}
		return 1
	}
}
func queryDataCenterList(schema_uuid string, page, size int, fields string) error {
	glogger.GLogger.Debug(schema_uuid, page, size, fields)

	return nil
}

/*
*
* last data
*
 */
func QueryDataCenterLast(rx typex.Rhilex, uuid string) func(L *lua.LState) int {
	return func(l *lua.LState) int {
		schema_uuid := l.ToString(2)
		fields := l.ToString(3)
		if err := queryDataCenterLast(schema_uuid, fields); err != nil {
			l.Push(lua.LString(err.Error()))
		} else {
			l.Push(lua.LNil)
		}
		return 1
	}
}
func queryDataCenterLast(schema_uuid string, fields string) error {
	glogger.GLogger.Debug(schema_uuid, fields)
	return nil
}
