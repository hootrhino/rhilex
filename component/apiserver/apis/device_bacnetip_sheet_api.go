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
	"github.com/hootrhino/rhilex/component/apiserver/dto"
	"github.com/samber/lo"
	"mime/multipart"
	"strconv"
	"time"

	"github.com/hootrhino/rhilex/glogger"

	"github.com/gin-gonic/gin"
	"github.com/hootrhino/rhilex/component/intercache"
	"github.com/hootrhino/rhilex/component/interdb"

	common "github.com/hootrhino/rhilex/component/apiserver/common"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/component/apiserver/server"
	"github.com/hootrhino/rhilex/component/apiserver/service"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
	"github.com/xuri/excelize/v2"
)

func InitBacnetIpRoute() {
	route := server.RouteGroup(server.ContextUrl("/bacnetip_data_sheet"))
	{
		route.POST(("/sheetImport"), server.AddRoute(BacnetIpSheetImport))
		route.GET(("/sheetExport"), server.AddRoute(BacnetIpSheetExport))
		route.GET(("/list"), server.AddRoute(BacnetIpSheetPageList))
		route.POST(("/update"), server.AddRoute(BacnetIpSheetCreateOrUpdate))
		route.DELETE(("/delIds"), server.AddRoute(BacnetIpSheetDeleteByUUIDs))
		route.DELETE(("/delAll"), server.AddRoute(BacnetIpSheetDeleteAll))
	}
}

func BacnetIpSheetImport(c *gin.Context, ruleEngine typex.Rhilex) {
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
	if deviceUuid == "" {
		err := errors.New("device_uuid is not allow empty")
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
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
	if Device.Type != typex.GENERIC_BACNET_IP.String() {
		c.JSON(common.HTTP_OK,
			common.Error("Invalid Device Type, Only Support Import BacnetIp Device"))
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

	list, err := parseBacnetExcel(file, "Sheet1", deviceUuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if err = service.BatchInsertBacnetDataPoint(list); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	ruleEngine.RestartDevice(deviceUuid)
	c.JSON(common.HTTP_OK, common.Ok())
}

func BacnetIpSheetExport(c *gin.Context, ruleEngine typex.Rhilex) {
	deviceUuid, _ := c.GetQuery("device_uuid")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment;filename=%v.xlsx",
		time.Now().UnixMilli()))
	var records []model.MBacnetDataPoint
	result := interdb.DB().Model(&model.MBacnetDataPoint{}).
		Where("device_uuid=?", deviceUuid).Find(&records)
	if result.Error != nil {
		c.JSON(common.HTTP_OK, common.Error400(result.Error))
		return
	}
	// header
	Headers := []string{
		"tag", "alias", "bacnetDeviceId", "objectType", "objectId",
	}
	xlsx := excelize.NewFile()
	defer func() {
		if err := xlsx.Close(); err != nil {
			glogger.GLogger.Errorf("close excel file, err=%v", err)
		}
	}()
	cell, _ := excelize.CoordinatesToCellName(1, 1)
	xlsx.SetSheetRow("Sheet1", cell, &Headers)
	if len(records) >= 1 {
		for idx, record := range records[0:] {
			Row := []any{
				record.Tag,
				record.Alias,
				record.BacnetDeviceId,
				record.ObjectType,
				record.ObjectId,
			}
			cell, _ = excelize.CoordinatesToCellName(1, idx+2)
			xlsx.SetSheetRow("Sheet1", cell, &Row)
		}
	}
	xlsx.WriteTo(c.Writer)
}

func BacnetIpSheetPageList(c *gin.Context, ruleEngine typex.Rhilex) {
	pager, err := service.ReadPageRequest(c)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	deviceUuid, _ := c.GetQuery("device_uuid")
	db := interdb.DB()
	tx := db.Scopes(service.Paginate(*pager))
	var count int64
	err = interdb.DB().Model(&model.MBacnetDataPoint{}).
		Where("device_uuid=?", deviceUuid).Count(&count).Error
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	var records []model.MBacnetDataPoint
	result := tx.Order("created_at DESC").Find(&records,
		&model.MBacnetDataPoint{DeviceUuid: deviceUuid})
	if result.Error != nil {
		c.JSON(common.HTTP_OK, common.Error400(result.Error))
		return
	}
	var recordsVo []dto.BacnetDataPointVO
	Slot := intercache.GetSlot(deviceUuid)
	if Slot != nil {
		for _, record := range records {
			value, ok := Slot[record.UUID]
			pointVo := dto.BacnetDataPointVO{
				UUID:           record.UUID,
				DeviceUUID:     record.DeviceUuid,
				Tag:            record.Tag,
				Alias:          record.Alias,
				BacnetDeviceId: record.BacnetDeviceId,
				ObjectType:     record.ObjectType,
				ObjectId:       record.ObjectId,
				ErrMsg:         value.ErrMsg,
			}
			if ok {
				pointVo.Status = func() int {
					if value.Value == "" {
						return 0
					}
					return 1
				}()
				pointVo.LastFetchTime = value.LastFetchTime
				pointVo.Value = value.Value
				recordsVo = append(recordsVo, pointVo)
			} else {
				recordsVo = append(recordsVo, pointVo)
			}
		}
	}

	Result := service.WrapPageResult(*pager, recordsVo, count)
	c.JSON(common.HTTP_OK, common.OkWithData(Result))
}

func BacnetIpSheetDeleteByUUIDs(c *gin.Context, ruleEngine typex.Rhilex) {
	type Form struct {
		UUIDs      []string `json:"uuids"`
		DeviceUUID string   `json:"device_uuid"`
	}
	form := Form{}
	if Error := c.ShouldBindJSON(&form); Error != nil {
		c.JSON(common.HTTP_OK, common.Error400(Error))
		return
	}
	err := service.BatchDeleteBacnetDataPoint(form.UUIDs, form.DeviceUUID)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	ruleEngine.RestartDevice(form.DeviceUUID)
	c.JSON(common.HTTP_OK, common.Ok())

}

func BacnetIpSheetDeleteAll(c *gin.Context, ruleEngine typex.Rhilex) {
	type Form struct {
		DeviceUUID string `json:"device_uuid"`
	}
	form := Form{}
	if Error := c.ShouldBindJSON(&form); Error != nil {
		c.JSON(common.HTTP_OK, common.Error400(Error))
		return
	}
	err := service.DeleteAllBacnetDataPointByDeviceUuid(form.DeviceUUID)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	ruleEngine.RestartDevice(form.DeviceUUID)
	c.JSON(common.HTTP_OK, common.Ok())

}

func BacnetIpSheetCreateOrUpdate(c *gin.Context, ruleEngine typex.Rhilex) {
	type Form struct {
		DeviceUUID string                              `json:"device_uuid"`
		Points     []dto.BacnetDataPointCreateOrUpdate `json:"points"`
	}
	form := Form{}
	err := c.ShouldBindJSON(&form)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	for _, Point := range form.Points {
		if err = checkBacnetPoint(Point); err != nil {
			c.JSON(common.HTTP_OK, common.Error400(err))
			return
		}
		if Point.UUID == "" {
			NewRow := model.MBacnetDataPoint{
				UUID:           utils.BacnetPointUUID(),
				DeviceUuid:     form.DeviceUUID,
				Tag:            Point.Tag,
				Alias:          Point.Alias,
				BacnetDeviceId: Point.BacnetDeviceId,
				ObjectType:     Point.ObjectType,
				ObjectId:       Point.ObjectId,
			}
			err0 := service.InsertBacnetDataPoint(NewRow)
			if err0 != nil {
				c.JSON(common.HTTP_OK, common.Error400(err0))
				return
			}
		} else {
			OldRow := model.MBacnetDataPoint{
				UUID:           Point.UUID,
				DeviceUuid:     form.DeviceUUID,
				Tag:            Point.Tag,
				Alias:          Point.Alias,
				BacnetDeviceId: Point.BacnetDeviceId,
				ObjectType:     Point.ObjectType,
				ObjectId:       Point.ObjectId,
			}
			err0 := service.UpdateBacnetDataPoint(OldRow)
			if err0 != nil {
				c.JSON(common.HTTP_OK, common.Error400(err0))
				return
			}
		}
	}
	ruleEngine.RestartDevice(form.DeviceUUID)
	c.JSON(common.HTTP_OK, common.Ok())

}

func parseBacnetExcel(r multipart.File, sheetName string, deviceUuid string) ([]model.MBacnetDataPoint, error) {
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
	// "tag", "alias", "bacnetDeviceId", "objectType", "objectId"
	err1 := errors.New("'Invalid Sheet Header, must follow fixed format: 【tag,alias,bacnetDeviceId,objectType,objectId】")

	const MIN_LEN = 5
	if len(rows[0]) < MIN_LEN {
		return nil, err1
	}
	// 严格检查表结构 oid,tag,alias,frequency
	if rows[0][0] != "tag" ||
		rows[0][1] != "alias" ||
		rows[0][2] != "bacnetDeviceId" ||
		rows[0][3] != "objectType" ||
		rows[0][4] != "objectId" {
		return nil, err1
	}

	list := make([]model.MBacnetDataPoint, 0)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < MIN_LEN {
			msg := fmt.Sprintf("illegal data, the data cell of row %d less than %d", i+1, MIN_LEN)
			return nil, errors.New(msg)
		}
		// oid,tag,alias,frequency
		tag := row[0]
		alias := row[1]
		bacnetDeviceId, _ := strconv.ParseInt(row[2], 10, 32)
		objectType := row[3]
		objectId, _ := strconv.ParseInt(row[4], 10, 32)

		createDto := dto.BacnetDataPointCreateOrUpdate{
			Tag:            tag,
			Alias:          alias,
			BacnetDeviceId: int(bacnetDeviceId),
			ObjectType:     objectType,
			ObjectId:       int(objectId),
		}
		err = checkBacnetPoint(createDto)
		if err != nil {
			return nil, err
		}

		model := model.MBacnetDataPoint{
			UUID:           utils.BacnetPointUUID(),
			DeviceUuid:     deviceUuid,
			Tag:            createDto.Tag,
			Alias:          createDto.Alias,
			BacnetDeviceId: createDto.BacnetDeviceId,
			ObjectType:     createDto.ObjectType,
			ObjectId:       createDto.ObjectId,
		}
		list = append(list, model)
	}
	return list, nil
}

func checkBacnetPoint(point dto.BacnetDataPointCreateOrUpdate) error {
	contains := lo.Contains(dto.ValidBacnetObjectType, point.ObjectType)
	if contains == false {
		return errors.New("illegal objectType")
	}
	return nil
}
