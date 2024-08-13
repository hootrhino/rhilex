package apis

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	common "github.com/hootrhino/rhilex/component/apiserver/common"
	"github.com/hootrhino/rhilex/component/apiserver/dto"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/component/apiserver/server"
	"github.com/hootrhino/rhilex/component/apiserver/service"
	"github.com/hootrhino/rhilex/component/apiserver/service/validatormanager"
	"github.com/hootrhino/rhilex/component/intercache"
	"github.com/hootrhino/rhilex/component/interdb"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
	"github.com/samber/lo"
	"github.com/xuri/excelize/v2"
	"time"
)

func InitDataPointRoute() {
	route := server.RouteGroup(server.ContextUrl("/datapoint"))
	route.POST(("/sheetImport"), server.AddRoute(DataPointSheetImport))
	route.GET(("/sheetExport"), server.AddRoute(DataPointSheetExport))
	route.GET(("/list"), server.AddRoute(DataPointSheetPageList))
	route.POST(("/update"), server.AddRouteV2(DataPointSheetCreateOrUpdate))
	route.DELETE(("/delIds"), server.AddRoute(DataPointSheetDeleteByUUIDs))
	route.DELETE(("/delAll"), server.AddRoute(DataPointSheetDeleteAll))
}

func DataPointSheetImport(c *gin.Context, ruleEngine typex.Rhilex) {
	// 解析 multipart/form-data 类型的请求体
	err := c.Request.ParseMultipartForm(1024 * 1024 * 10)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	deviceUuid := c.Request.Form.Get("device_uuid")
	if deviceUuid == "" {
		err = errors.New("device_uuid is not allow empty")
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	device, err := service.GetMDeviceWithUUID(deviceUuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	validator, err := validatormanager.GetByType(device.Type)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	defer file.Close()
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

	// parse
	excelFile, err := excelize.OpenReader(file)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error("read file failed"))
		return
	}
	defer func() {
		excelFile.Close()
	}()
	points, err := validator.ParseImportFile(excelFile)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	for i := range points {
		points[i].DeviceUuid = deviceUuid
		points[i].UUID = utils.UnionPointUUID()
		points[i].Frequency = lo.Clamp(points[i].Frequency, 50, 100000)
	}
	err = service.BatchDataPointCreate(points)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	ruleEngine.RestartDevice(deviceUuid)
	c.JSON(common.HTTP_OK, common.Ok())
}

func DataPointSheetExport(c *gin.Context, ruleEngine typex.Rhilex) {
	deviceUuid, _ := c.GetQuery("device_uuid")
	device, err := service.GetMDeviceWithUUID(deviceUuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	validator, err := validatormanager.GetByType(device.Type)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	points, err := service.ListDataPointByUuid(deviceUuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	xlsx := excelize.NewFile()
	defer func() {
		if err := xlsx.Close(); err != nil {
			glogger.GLogger.Errorf("close excel file, err=%v", err)
		}
	}()

	err = validator.Export(xlsx, points)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment;filename=%v.xlsx",
		time.Now().UnixMilli()))

	xlsx.WriteTo(c.Writer)
}

func DataPointSheetPageList(c *gin.Context, ruleEngine typex.Rhilex) {
	pager, err := service.ReadPageRequest(c)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	deviceUuid, _ := c.GetQuery("device_uuid")
	db := interdb.DB()
	tx := db.Scopes(service.Paginate(*pager))
	var count int64
	err = interdb.DB().Model(&model.MDataPoint{}).
		Where("device_uuid=?", deviceUuid).Count(&count).Error
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	var records []model.MDataPoint
	result := tx.Order("created_at DESC").Find(&records,
		&model.MDataPoint{DeviceUuid: deviceUuid})
	if result.Error != nil {
		c.JSON(common.HTTP_OK, common.Error400(result.Error))
		return
	}

	var recordsVo []dto.DataPointVO
	Slot := intercache.GetSlot(deviceUuid)
	if Slot != nil {
		for _, record := range records {
			value, ok := Slot[record.UUID]
			pointVo := dto.DataPointVO{
				UUID:       record.UUID,
				DeviceUUID: record.DeviceUuid,
				Tag:        record.Tag,
				Alias:      record.Alias,
				Config:     record.GetConfig(),
				ErrMsg:     value.ErrMsg,
			}
			if ok {
				pointVo.Status = func() uint32 {
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

func DataPointSheetCreateOrUpdate(c *gin.Context, ruleEngine typex.Rhilex) (any, error) {
	type Form struct {
		DeviceUUID string                           `json:"device_uuid"`
		Points     []dto.DataPointCreateOrUpdateDTO `json:"points"`
	}
	form := Form{}
	err := c.ShouldBindJSON(&form)
	if err != nil {
		return nil, err
	}
	device, err := service.GetMDeviceWithUUID(form.DeviceUUID)
	if err != nil {
		return nil, err
	}

	validator, err := validatormanager.GetByType(device.Type)
	if err != nil {
		return nil, err
	}

	creates := make([]model.MDataPoint, 0, len(form.Points))
	updates := make([]model.MDataPoint, 0, len(form.Points))
	for i := range form.Points {
		pointDTO := form.Points[i]
		if !utils.IsValidColumnName(pointDTO.Tag) {
			return nil, fmt.Errorf("'Invalid Tag Name:%s", pointDTO.Tag)
		}
		point, err := validator.Convert(form.Points[i])
		if err != nil {
			return nil, err
		}
		point.UUID = pointDTO.UUID
		point.Tag = pointDTO.Tag
		point.Alias = pointDTO.Alias
		point.DeviceUuid = form.DeviceUUID
		if point.UUID == "" ||
			point.UUID == "new" ||
			point.UUID == "copy" {
			point.UUID = utils.UnionPointUUID()
			creates = append(creates, point)
		} else {
			updates = append(updates, point)
		}
	}
	if len(creates) > 0 {
		err := service.BatchDataPointCreate(creates)
		if err != nil {
			return nil, err
		}
	}

	if len(updates) > 0 {
		err := service.BatchDataPointUpdate(updates)
		if err != nil {
			return nil, err
		}
	}

	ruleEngine.RestartDevice(form.DeviceUUID)
	return nil, nil
}

func DataPointSheetDeleteByUUIDs(c *gin.Context, ruleEngine typex.Rhilex) {
	type Form struct {
		UUIDs      []string `json:"uuids"`
		DeviceUUID string   `json:"device_uuid"`
	}
	form := Form{}
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	err := service.BatchDeleteDataPointByUuids(form.DeviceUUID, form.UUIDs)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	ruleEngine.RestartDevice(form.DeviceUUID)
	c.JSON(common.HTTP_OK, common.Ok())
}

func DataPointSheetDeleteAll(c *gin.Context, ruleEngine typex.Rhilex) {
	type Form struct {
		DeviceUUID string `json:"device_uuid"`
	}
	form := Form{}
	if Error := c.ShouldBindJSON(&form); Error != nil {
		c.JSON(common.HTTP_OK, common.Error400(Error))
		return
	}
	err := service.BatchDeleteDataPointByDeviceUuid(form.DeviceUUID)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	ruleEngine.RestartDevice(form.DeviceUUID)
	c.JSON(common.HTTP_OK, common.Ok())
}
