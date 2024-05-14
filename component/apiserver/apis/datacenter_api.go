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

package apis

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	common "github.com/hootrhino/rhilex/component/apiserver/common"
	"github.com/hootrhino/rhilex/component/apiserver/server"
	"github.com/hootrhino/rhilex/component/apiserver/service"
	"github.com/hootrhino/rhilex/component/interdb"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/xuri/excelize/v2"
)

type SchemaDDLVo struct {
}

func InitDataCenterApi() {
	datacenterApi := server.RouteGroup(server.ContextUrl("/datacenter"))
	datacenterApi.GET("/listSchemaDDL", server.AddRoute(ListSchemaDDL))
	datacenterApi.GET("/schemaDDLDetail", server.AddRoute(SchemaDDLDetail))
	datacenterApi.GET("/queryDataList", server.AddRoute(QueryDDLDataList))
	datacenterApi.GET("/queryLastData", server.AddRoute(QueryDDLLastData))
	datacenterApi.GET("/exportData", server.AddRoute(ExportData))
	datacenterApi.GET("/schemaDDLDefine", server.AddRoute(GetSchemaDDLDefine))
}

/*
*
* 列表, 先获取数据模型，然后拼接Table(CREATE TABLE data_center_0002)
*
 */
func ListSchemaDDL(c *gin.Context, ruleEngine typex.Rhilex) {
	DataSchemas := []IoTSchemaVo{}
	for _, vv := range service.AllDataSchema() {
		IoTSchemaVo := IoTSchemaVo{
			UUID:        vv.UUID,
			Published:   vv.Published,
			Name:        vv.Name,
			Description: vv.Description,
		}
		DataSchemas = append(DataSchemas, IoTSchemaVo)
	}
	c.JSON(common.HTTP_OK, common.OkWithData(DataSchemas))
}

/*
*
* 详情, 先返回DDL算了
*
 */
func SchemaDDLDetail(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	TableSchemas, err := service.GetTableSchema(uuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	c.JSON(common.HTTP_OK, common.OkWithData(TableSchemas))
}

/*
*
* 导出, ids[]
*
 */
func ExportData(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	TableSchemas, err := service.GetTableSchema(uuid) // PRAGMA table_info
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	Headers := []string{}
	OneRow := make([]interface{}, len(TableSchemas))
	for i, TableSchema := range TableSchemas {
		Headers = append(Headers, TableSchema.Name)
		switch TableSchema.Type {
		case "INTEGER":
			OneRow[i] = new(int)
		case "BOOLEAN":
			OneRow[i] = new(bool)
		case "DATETIME":
			OneRow[i] = new(string)
		case "TIMESTAMP":
			OneRow[i] = new(string)
		case "TEXT":
			OneRow[i] = new(string)
		case "REAL":
			OneRow[i] = new(float32)
		default:
			OneRow[i] = new(string) // 不知道啥类型就String
		}
	}

	xlsx := excelize.NewFile()
	defer func() {
		if err := xlsx.Close(); err != nil {
			glogger.GLogger.Errorf("close excel file, err=%v", err)
		}
	}()
	cell, _ := excelize.CoordinatesToCellName(1, 1)
	xlsx.SetSheetRow("Sheet1", cell, &Headers)
	tableName := fmt.Sprintf("data_center_%s", uuid)
	rows, Error := interdb.DB().Table(tableName).Rows()
	if Error != nil {
		c.JSON(common.HTTP_OK, common.Error400(Error))
		return
	}
	idx := 0
	for rows.Next() {
		if err := rows.Scan(OneRow...); err != nil {
			c.JSON(common.HTTP_OK, common.Error400(err))
			return
		}
		cell, _ = excelize.CoordinatesToCellName(1, idx+2)
		SheetRow := []interface{}{}
		for _, Column := range OneRow {
			switch T := Column.(type) {
			case *bool:
				SheetRow = append(SheetRow, *T)
			case *int:
				SheetRow = append(SheetRow, *T)
			case *float32:
				SheetRow = append(SheetRow, *T)
			case *string:
				SheetRow = append(SheetRow, *T)
			default:
				SheetRow = append(SheetRow, "NULL") // 不支持的类型
			}
		}
		xlsx.SetSheetRow("Sheet1", cell, &SheetRow)
		idx++
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment;filename=%v.xlsx",
		time.Now().UnixMilli()))
	xlsx.WriteTo(c.Writer)
	c.JSON(common.HTTP_OK, common.Ok())
}

/*
*
* 分页查找
*
 */
func QueryDDLDataList(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	order, _ := c.GetQuery("order")
	selectFields, _ := c.GetQueryArray("select")
	pager, err := service.ReadPageRequest(c)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if pager.Size > 1000 {
		c.JSON(common.HTTP_OK, common.Error("Query size too large, Must less than 1000"))
		return
	}
	DbTx := interdb.DB().Scopes(service.Paginate(*pager))
	records := []map[string]any{}
	tableName := fmt.Sprintf("data_center_%s", uuid)
	Order := "DESC"
	if order == "DESC" || order == "ASC" {
		Order = order
	}
	// Default order by ts desc
	result := DbTx.Select(selectFields).Table(tableName).Order("create_at " + Order).Scan(&records)
	if result.Error != nil {
		c.JSON(common.HTTP_OK, common.Error400(result.Error))
		return
	}
	var count int64
	err2 := DbTx.Table(tableName).Count(&count).Error
	if err2 != nil {
		c.JSON(common.HTTP_OK, common.Error400(err2))
		return
	}
	Result := service.WrapPageResult(*pager, records, count)
	c.JSON(common.HTTP_OK, common.OkWithData(Result))
}

/*
*
* 最新数据
*
 */
func QueryDDLLastData(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	selectFields, _ := c.GetQueryArray("select")

	tableName := fmt.Sprintf("data_center_%s", uuid)
	record := map[string]any{}
	result := interdb.DB().Select(selectFields).
		Table(tableName).
		Order("create_at DESC").Limit(1).Find(&record)
	if result.Error != nil {
		c.JSON(common.HTTP_OK, common.Error400(result.Error))
		return
	}
	c.JSON(common.HTTP_OK, common.OkWithData(record))
}

/*
*
*Lua辅助生成器
*
 */
func GenDataToLuaFunc(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	TableColumnInfos, err := service.GetTableSchema(uuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	// {k=v...}
	s := []string{}
	for _, TableColumn := range TableColumnInfos {
		if TableColumn.Name == "id" || TableColumn.Name == "create_at" {
			continue
		}
		s = append(s, fmt.Sprintf("%s=%v",
			TableColumn.Name, parse_type(TableColumn.Type)))
	}
	luaS := fmt.Sprintf("local errRdsSave = rds:Save('%s', {%s\n}\n)",
		uuid, strings.Join(s, ", "))
	c.JSON(common.HTTP_OK, common.OkWithData(luaS))
}

/*
*
* 获取定义
*
 */
func GetSchemaDDLDefine(c *gin.Context, ruleEngine typex.Rhilex) {
	type tableColumn struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	uuid, _ := c.GetQuery("uuid")
	TableColumnInfos, err := service.GetTableSchema(uuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	tableColumns := []tableColumn{}
	for _, TableColumn := range TableColumnInfos {
		if TableColumn.Name == "id" || TableColumn.Name == "create_at" {
			continue
		}
		tableColumns = append(tableColumns, tableColumn{
			Name: TableColumn.Name,
			Type: TableColumn.Type,
		})
	}
	c.JSON(common.HTTP_OK, common.OkWithData(tableColumns))

}
func parse_type(dbType string) interface{} {
	switch dbType {
	case "TEXT":
		return "''"
	case "INTEGER":
		return 0
	case "REAL":
		return 0
	case "BOOLEAN":
		return false
	default:
		return "''"
	}
}
