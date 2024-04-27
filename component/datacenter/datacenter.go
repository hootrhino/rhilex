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

package datacenter

import (
	"fmt"

	"github.com/hootrhino/rhilex/typex"
)

var __DefaultDataCenter *DataCenter

/*
*
* 留着未来扩充数据中心的功能
*
 */
type DataCenter struct {
	LocalDb DataSource
	rhilex  typex.Rhilex
}

func InitDataCenter(rhilex typex.Rhilex) {
	__DefaultDataCenter = new(DataCenter)
	__DefaultDataCenter.rhilex = rhilex
	__DefaultDataCenter.LocalDb = InitLocalDb(rhilex)
}

/*
*
* 获取表格定义
*
 */
func SchemaList() []SchemaDetail {
	Schemas := []SchemaDetail{}
	// 本地内部数据中心
	Schemas = append(Schemas, __DefaultDataCenter.LocalDb.GetSchemaDetail("rhilex_internal_datacenter"))
	return Schemas
}

/*
*
* 表结构
*
 */

func GetSchemaDefine(uuid string) (SchemaDefine, error) {
	schemaDefine := SchemaDefine{}
	return schemaDefine, nil

}

/*
*
* 仓库列表
*
 */
func SchemaDefineList() ([]SchemaDefine, error) {
	var err error
	ColumnsMap := []SchemaDefine{}

	return ColumnsMap, err
}

/*
*
* 获取仓库详情, 现阶段写死的, 后期会在proto中实现
*
 */
func GetSchemaDetail(uuid string) SchemaDetail {
	return SchemaDetail{
		UUID:        "********",
		Name:        "Local",
		LocalPath:   ".local",
		CreateTs:    0,
		Size:        0,
		StorePath:   "test.db",
		Description: "Test Db",
	}
}

/*
*
* 查询，第一个参数是查询请求，针对Sqlite就是SQL，针对mongodb就是JS，根据具体情况而定
  TODO 未来实现：DataCenter['uuid'].Query(query string)
*
*/

func Query(uuid, query string) ([]map[string]any, error) {

	// 本地
	// Rows 来自本地Sqlite查询
	if uuid == "rhilex_internal_datacenter" {
		LocalResult, err := __DefaultDataCenter.LocalDb.Query(uuid, query)
		return LocalResult, err
	}
	// 外部
	return nil, fmt.Errorf("unsupported db type:" + uuid)

}
