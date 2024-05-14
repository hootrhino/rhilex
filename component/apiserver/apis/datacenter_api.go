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

	"github.com/gin-gonic/gin"
	common "github.com/hootrhino/rhilex/component/apiserver/common"
	"github.com/hootrhino/rhilex/component/apiserver/server"
	"github.com/hootrhino/rhilex/component/apiserver/service"
	"github.com/hootrhino/rhilex/component/interdb"
	"github.com/hootrhino/rhilex/typex"
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
* 导出
*
 */
func ExportData(c *gin.Context, ruleEngine typex.Rhilex) {
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
	DbTx := interdb.DB().Scopes(service.Paginate(*pager))
	records := []map[string]any{}
	tableName := fmt.Sprintf("data_center_%s", uuid)
	Order := "DESC"
	if order == "DESC" || order == "ASC" {
		Order = order
	}
	result := DbTx.Select(selectFields).Table(tableName).Order("ts " + Order).Scan(&records)
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
		Order("ts DESC").Limit(1).Find(&record)
	if result.Error != nil {
		c.JSON(common.HTTP_OK, common.Error400(result.Error))
		return
	}
	c.JSON(common.HTTP_OK, common.OkWithData(record))
}
