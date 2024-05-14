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
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hootrhino/rhilex/component/interdb"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"

	lua "github.com/hootrhino/gopher-lua"
)

// GenInsertSql 生成 INSERT SQL 语句
func GenInsertSql(tableName string, rowList [][2]interface{}) (string, error) {
	if len(rowList) == 0 {
		return "", fmt.Errorf("no rows to insert")
	}
	columnStr := strings.Join(buildColumns(rowList), ", ")
	values := make([]string, len(rowList))
	for i, row := range rowList {
		value := row[1]
		valueStr, err := formatValue(value)
		if err != nil {
			return "", err
		}
		values[i] = valueStr
	}
	insertSql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, columnStr, strings.Join(values, ", "))
	return insertSql, nil
}

// buildColumns 构建一个包含列名的字符串列表
func buildColumns(rowList [][2]interface{}) []string {
	var columnList []string
	for _, row := range rowList {
		column := row[0]
		columnList = append(columnList, fmt.Sprintf("`%s`", column))
	}
	return columnList
}

// formatValue 根据值的类型格式化 SQL 值
func formatValue(val interface{}) (string, error) {
	switch v := val.(type) {
	case string:
		return fmt.Sprintf("'%s'", v), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%f", v), nil
	case bool:
		if v {
			return "1", nil
		} else {
			return "0", nil
		}
	default:
		return "", fmt.Errorf("unsupported type: %s", reflect.TypeOf(val))
	}
}

/*
*
 */
func InsertToDataCenterTable(rx typex.Rhilex, uuid string) func(L *lua.LState) int {
	return func(l *lua.LState) int {
		schema_uuid := l.ToString(2)
		kvs := l.ToTable(3)
		RowList := [][2]interface{}{}
		kvs.ForEach(func(k, v lua.LValue) {
			Row := [2]interface{}{}
			// K 只能String
			if k.Type() == lua.LTString {
				// create_at 不允许用户填写
				if Row[0] != "create_at" {
					switch v.Type() {
					case lua.LTString:
						Row[0] = lua.LVAsString(k)
						Row[1] = lua.LVAsString(v)
					case lua.LTNumber:
						Row[0] = lua.LVAsString(k)
						Row[1] = float64(lua.LVAsNumber(v))
					case lua.LTBool:
						Row[0] = lua.LVAsString(k)
						Row[1] = bool(lua.LVAsBool(v))
					case lua.LTNil:
						Row[0] = lua.LVAsString(k)
						Row[1] = nil
					default:
						Row[0] = lua.LVAsString(k)
						Row[1] = nil // 不支持其他类型
					}
					RowList = append(RowList, Row)
				}
			}
		})
		// create_at 最后追加
		RowList = append(RowList, [2]interface{}{
			"create_at", time.Now(),
		})
		if err := saveToDataCenter(fmt.Sprintf("data_center_%s", schema_uuid), RowList); err != nil {
			l.Push(lua.LString(err.Error()))
		} else {
			l.Push(lua.LNil)
		}
		return 1
	}
}

// Save to local DataCenter
func saveToDataCenter(schema_uuid string, RowList [][2]interface{}) error {
	Sql, err0 := GenInsertSql(schema_uuid, RowList)
	glogger.GLogger.Debug(Sql)
	if err0 != nil {
		return err0
	}
	err1 := interdb.DB().Exec(Sql).Error
	if err1 != nil {
		return err1
	}
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
		if err := queryDataCenterList(fmt.Sprintf("data_center_%s", schema_uuid),
			int(page), int(size), fields); err != nil {
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
