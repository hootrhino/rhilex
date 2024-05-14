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

package datacenter

import (
	"database/sql"
	"fmt"
	"strings"
)

type DDLColumn struct {
	Name        string
	Type        string
	Description string
}
type SchemaDDL struct {
	SchemaUUID string      // 数据模型的UUID，用来生成数据仓库的表
	DDLColumns []DDLColumn // 包含了哪些列
}

/*
*
* 生成建仓语句
*
 */
func GenerateSQLiteCreateTableDDL(schemaDDL SchemaDDL) (string, error) {
	if schemaDDL.SchemaUUID == "" {
		return "", fmt.Errorf("SchemaUUID cannot be empty")
	}

	var columns []string
	for i, col := range schemaDDL.DDLColumns {
		columnDefine := fmt.Sprintf("%s %s", col.Name, SqliteTypeMappingSchemaType(col.Type))
		if col.Description != "" {
			switch col.Type {
			case "STRING":
				columnDefine += " NOT NULL DEFAULT ''"
			case "INTEGER":
				if col.Name == "id" {
					columnDefine += " NOT NULL PRIMARY KEY AUTOINCREMENT"
				} else {
					columnDefine += " NOT NULL DEFAULT 0"
				}
			case "FLOAT":
				columnDefine += " NOT NULL DEFAULT 0"
			case "BOOL":
				columnDefine += " NOT NULL DEFAULT 0"
			case "DATETIME":
				columnDefine += " NOT NULL DEFAULT CURRENT_TIMESTAMP"
			}
			if i != len(schemaDDL.DDLColumns)-1 {
				columnDefine += ","
			}
		}
		columns = append(columns, columnDefine)
	}

	tableName := schemaDDL.SchemaUUID
	createTableStmt := fmt.Sprintf("CREATE TABLE IF NOT EXISTS '%s' (\n%s\n)",
		tableName, strings.Join(columns, "\n"))

	return createTableStmt, nil
}

/*
*
* 删除表
*
 */

func SqliteTypeMappingSchemaType(goType string) string {
	switch goType {
	case "STRING":
		return "TEXT"
	case "INTEGER":
		return "INTEGER"
	case "FLOAT":
		return "REAL"
	case "BOOL":
		return "BOOLEAN"
	case "DATETIME":
		return "DATETIME"
	default:
		return "TEXT"
	}
}

// TableColumnInfo 表结构列信息结构体
type TableColumnInfo struct {
	Name string
	Type string
}

// GetTableSchema 获取表结构
func GetTableSchema(db *sql.DB, tableName string) ([]TableColumnInfo, error) {
	var columns []TableColumnInfo
	rows, err := db.Query("PRAGMA table_info(" + tableName + ");")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var columnName, dataType, defaultValue string
		var notNull, primaryKey, autoIncrement int
		if err := rows.Scan(&columnName, &dataType,
			&defaultValue, &notNull, &primaryKey, &autoIncrement); err != nil {
			return nil, err
		}
		columns = append(columns, TableColumnInfo{
			Name: columnName,
			Type: dataType,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return columns, nil
}
