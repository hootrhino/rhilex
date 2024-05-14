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

package service

import (
	"fmt"

	"github.com/hootrhino/rhilex/component/datacenter"
)

// TableColumnInfo 表结构列信息结构体
type TableColumnInfo struct {
	Cid        int    `json:"cid"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	NotNull    bool   `json:"not_null"`
	Default    any    `json:"default"`
	PrimaryKey bool   `json:"primary_key"`
}

// GetTableSchema 使用 GORM 的 DB 对象执行 PRAGMA table_info
func GetTableSchema(tableName string) ([]TableColumnInfo, error) {
	columns := []TableColumnInfo{}
	// 注意：从datacenter取数据
	rows, err := datacenter.DB().Raw(fmt.Sprintf("PRAGMA table_info(\"data_center_%s\");", tableName)).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var column TableColumnInfo
		if err := rows.Scan(
			&column.Cid,
			&column.Name,
			&column.Type,
			&column.NotNull,
			&column.Default,
			&column.PrimaryKey); err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return columns, nil
}

func SchemaDDLDetail(tableName string) ([]TableColumnInfo, error) {
	return GetTableSchema(tableName)
}

// QueryDataList
func QueryDataList(uuid string) error {
	return nil
}

// QueryLastData
func QueryLastData(uuid string) error {
	return nil
}
