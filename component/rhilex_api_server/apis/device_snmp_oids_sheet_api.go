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

package apis

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/hootrhino/rhilex/glogger"

	"github.com/gin-gonic/gin"
	"github.com/hootrhino/rhilex/component/intercache"
	"github.com/hootrhino/rhilex/component/interdb"

	common "github.com/hootrhino/rhilex/component/rhilex_api_server/common"
	"github.com/hootrhino/rhilex/component/rhilex_api_server/model"
	"github.com/hootrhino/rhilex/component/rhilex_api_server/server"
	"github.com/hootrhino/rhilex/component/rhilex_api_server/service"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
	"github.com/xuri/excelize/v2"
)

func InitSnmpRoute() {
	// 华中数控 点位表
	Route := server.RouteGroup(server.ContextUrl("/snmp_oids_sheet"))
	{
		Route.POST(("/sheetImport"), server.AddRoute(SnmpSheetImport))
		Route.GET(("/sheetExport"), server.AddRoute(SnmpOidsExport))
		Route.GET(("/list"), server.AddRoute(SnmpSheetPageList))
		Route.POST(("/update"), server.AddRoute(SnmpSheetUpdate))
		Route.DELETE(("/delAll"), server.AddRoute(SnmpSheetDeleteAll))
		Route.DELETE(("/delIds"), server.AddRoute(SnmpSheetDeleteByUUIDs))
	}
}

type SnmpOidVo struct {
	UUID          string  `json:"uuid,omitempty"`
	DeviceUUID    string  `json:"device_uuid"`
	Oid           string  `json:"oid"`
	Tag           string  `json:"tag"`
	Alias         string  `json:"alias"`
	Frequency     *uint64 `json:"frequency"`
	ErrMsg        string  `json:"errMsg"`        // 运行时数据
	Status        int     `json:"status"`        // 运行时数据
	LastFetchTime uint64  `json:"lastFetchTime"` // 运行时数据
	Value         string  `json:"value"`         // 运行时数据
}

/*
*
* 特殊设备需要和外界交互，这里主要就是一些设备的点位表导入导出等支持
*  http://127.0.0.1:2580/api/v1/Snmp_data_sheet/export
 */

// SnmpOids 获取Snmp_excel类型的点位数据
func SnmpOidsExport(c *gin.Context, ruleEngine typex.Rhilex) {
	deviceUuid, _ := c.GetQuery("device_uuid")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment;filename=%v.xlsx",
		time.Now().UnixMilli()))
	var records []model.MSnmpOid
	result := interdb.DB().Order("created_at DESC").Find(&records,
		&model.MSnmpOid{DeviceUuid: deviceUuid})
	if result.Error != nil {
		c.JSON(common.HTTP_OK, common.Error400(result.Error))
		return
	}
	// oid	tag	alias	frequency
	Headers := []string{
		"oid", "tag", "alias", "frequency",
	}
	xlsx := excelize.NewFile()
	defer func() {
		if err := xlsx.Close(); err != nil {
			glogger.GLogger.Errorf("close excel file, err=%v", err)
		}
	}()
	cell, _ := excelize.CoordinatesToCellName(1, 1)
	xlsx.SetSheetRow("Sheet1", cell, &Headers)
	if len(records) > 1 {
		for idx, record := range records[0:] {
			Row := []string{
				record.Oid, record.Tag, record.Alias, fmt.Sprintf("%d", record.Frequency),
			}
			cell, _ = excelize.CoordinatesToCellName(1, idx+2)
			xlsx.SetSheetRow("Sheet1", cell, &Row)
		}
	}
	xlsx.WriteTo(c.Writer)
}

// 分页获取
// SELECT * FROM `m_snmp_oids` WHERE
// `m_snmp_oids`.`device_uuid` = "DEVICEDQNLO8"
// ORDER BY
// created_at DESC LIMIT 2 OFFSET 10
func SnmpSheetPageList(c *gin.Context, ruleEngine typex.Rhilex) {
	pager, err := service.ReadPageRequest(c)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	deviceUuid, _ := c.GetQuery("device_uuid")
	db := interdb.DB()
	tx := db.Scopes(service.Paginate(*pager))
	var count int64
	err1 := interdb.DB().Model(&model.MSnmpOid{}).
		Where("device_uuid=?", deviceUuid).Count(&count).Error
	if err1 != nil {
		c.JSON(common.HTTP_OK, common.Error400(err1))
		return
	}
	var records []model.MSnmpOid
	result := tx.Order("created_at DESC").Find(&records,
		&model.MSnmpOid{DeviceUuid: deviceUuid})
	if result.Error != nil {
		c.JSON(common.HTTP_OK, common.Error400(result.Error))
		return
	}
	recordsVo := []SnmpOidVo{}
	Slot := intercache.GetSlot(deviceUuid)
	if Slot != nil {
		for _, record := range records {
			Value, ok := Slot[record.UUID]
			Vo := SnmpOidVo{
				UUID:       record.UUID,
				Oid:        record.Oid,
				DeviceUUID: record.DeviceUuid,
				Tag:        record.Tag,
				Alias:      record.Alias,
				Frequency:  &record.Frequency,
				ErrMsg:     Value.ErrMsg,
			}
			if ok {
				Vo.Status = func() int {
					if Value.Value == "" {
						return 0
					}
					return 1
				}() // 运行时
				Vo.LastFetchTime = Value.LastFetchTime // 运行时
				Vo.Value = Value.Value                 // 运行时
				recordsVo = append(recordsVo, Vo)
			} else {
				recordsVo = append(recordsVo, Vo)
			}
		}
	}

	Result := service.WrapPageResult(*pager, recordsVo, count)
	c.JSON(common.HTTP_OK, common.OkWithData(Result))
}

/*
*
* 删除单行
*
 */
func SnmpSheetDeleteByUUIDs(c *gin.Context, ruleEngine typex.Rhilex) {
	type Form struct {
		UUIDs      []string `json:"uuids"`
		DeviceUUID string   `json:"device_uuid"`
	}
	form := Form{}
	if Error := c.ShouldBindJSON(&form); Error != nil {
		c.JSON(common.HTTP_OK, common.Error400(Error))
		return
	}
	err := service.DeleteSnmpOidByDevice(form.UUIDs, form.DeviceUUID)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	ruleEngine.RestartDevice(form.DeviceUUID)
	c.JSON(common.HTTP_OK, common.Ok())

}

/*
*
*删除
*
 */
func SnmpSheetDeleteAll(c *gin.Context, ruleEngine typex.Rhilex) {
	type Form struct {
		DeviceUUID string `json:"device_uuid"`
	}
	form := Form{}
	if Error := c.ShouldBindJSON(&form); Error != nil {
		c.JSON(common.HTTP_OK, common.Error400(Error))
		return
	}
	err := service.DeleteAllSnmpOidByDevice(form.DeviceUUID)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	ruleEngine.RestartDevice(form.DeviceUUID)
	c.JSON(common.HTTP_OK, common.Ok())

}

/*
*
* 检查点位合法性
*
 */
func checkSnmpOids(M SnmpOidVo) error {
	if M.Tag == "" {
		return fmt.Errorf("'Missing required param 'name'")
	}
	if len(M.Tag) > 256 {
		return fmt.Errorf("'Tag length must range of 1-256")
	}
	if M.Alias == "" {
		return fmt.Errorf("'Missing required param 'alias'")
	}
	if len(M.Alias) > 256 {
		return fmt.Errorf("'Alias length must range of 1-256")
	}
	return nil
}

/*
*
* 更新点位表
*
 */
func SnmpSheetUpdate(c *gin.Context, ruleEngine typex.Rhilex) {
	type Form struct {
		DeviceUUID string      `json:"device_uuid"`
		SnmpOids   []SnmpOidVo `json:"snmp_oids"`
	}
	// SnmpOids := []SnmpOidVo{}
	form := Form{}
	err := c.ShouldBindJSON(&form)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	for _, SnmpDataPoint := range form.SnmpOids {
		if err := checkSnmpOids(SnmpDataPoint); err != nil {
			c.JSON(common.HTTP_OK, common.Error400(err))
			return
		}
		if SnmpDataPoint.UUID == "" {
			NewRow := model.MSnmpOid{
				UUID:      utils.SnmpOidUUID(),
				Tag:       SnmpDataPoint.Tag,
				Alias:     SnmpDataPoint.Alias,
				Frequency: *SnmpDataPoint.Frequency,
			}
			err0 := service.InsertSnmpOid(NewRow)
			if err0 != nil {
				c.JSON(common.HTTP_OK, common.Error400(err0))
				return
			}
		} else {
			OldRow := model.MSnmpOid{
				UUID:       SnmpDataPoint.UUID,
				DeviceUuid: SnmpDataPoint.DeviceUUID,
				Tag:        SnmpDataPoint.Tag,
				Alias:      SnmpDataPoint.Alias,
				Frequency:  *SnmpDataPoint.Frequency,
			}
			err0 := service.UpdateSnmpOid(OldRow)
			if err0 != nil {
				c.JSON(common.HTTP_OK, common.Error400(err0))
				return
			}
		}
	}
	ruleEngine.RestartDevice(form.DeviceUUID)
	c.JSON(common.HTTP_OK, common.Ok())

}

// SnmpSheetImport 上传Excel文件
func SnmpSheetImport(c *gin.Context, ruleEngine typex.Rhilex) {
	// 解析 multipart/form-data 类型的请求体
	err := c.Request.ParseMultipartForm(1024 * 1024 * 10)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	defer file.Close()
	deviceUuid := c.Request.Form.Get("device_uuid")
	type DeviceDto struct {
		UUID string
		Name string
		Type string
	}
	Device := DeviceDto{}
	errDb := interdb.DB().Table("m_devices").
		Where("uuid=?", deviceUuid).Find(&Device).Error
	if errDb != nil {
		c.JSON(common.HTTP_OK, common.Error400(errDb))
		return
	}
	if Device.Type != typex.GENERIC_SNMP.String() {
		c.JSON(common.HTTP_OK,
			common.Error("Invalid Device Type, Only Support Import Snmp Device"))
		return
	}
	contentType := header.Header.Get("Content-Type")
	if contentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" &&
		contentType != "application/vnd.ms-excel" {
		c.JSON(common.HTTP_OK, common.Error("File Must be Excel Sheet"))
		return
	}
	// 判断文件大小是否符合要求（10MB）
	if header.Size > 1024*1024*10 {
		c.JSON(common.HTTP_OK, common.Error("Excel file size cannot be greater than 10MB"))
		return
	}
	list, err := parseSnmpOidExcel(file, "Sheet1", deviceUuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if err = service.InsertSnmpOids(list); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	ruleEngine.RestartDevice(deviceUuid)
	c.JSON(common.HTTP_OK, common.Ok())
}

/*
*
* 解析表格
*
 */

func parseSnmpOidExcel(r io.Reader, sheetName string,
	deviceUuid string) (list []model.MSnmpOid, err error) {
	excelFile, err := excelize.OpenReader(r)
	if err != nil {
		return nil, err
	}
	defer func() {
		excelFile.Close()
	}()
	// 读取表格
	rows, err := excelFile.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	// 判断首行标头
	// oid,tag,alias,frequency
	err1 := errors.New("'Invalid Sheet Header, must follow fixed format: 【oid,tag,alias,frequency】")
	if len(rows[0]) < 4 {
		return nil, err1
	}
	// 严格检查表结构 oid,tag,alias,frequency
	if rows[0][0] != "oid" ||
		rows[0][1] != "tag" ||
		rows[0][2] != "alias" ||
		rows[0][3] != "frequency" {
		return nil, err1
	}

	list = make([]model.MSnmpOid, 0)
	// name, alias, function, group, address
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		// oid,tag,alias,frequency
		oid := row[0]
		tag := row[1]
		alias := row[2]
		frequency, _ := strconv.ParseUint(row[3], 10, 64)
		if err := checkSnmpOids(SnmpOidVo{
			Oid:       oid,
			Tag:       tag,
			Alias:     alias,
			Frequency: &frequency,
		}); err != nil {
			return nil, err
		}
		//
		model := model.MSnmpOid{
			UUID:       utils.SnmpOidUUID(),
			DeviceUuid: deviceUuid,
			Oid:        oid,
			Tag:        tag,
			Alias:      alias,
			Frequency:  frequency,
		}
		list = append(list, model)
	}
	return list, nil
}
