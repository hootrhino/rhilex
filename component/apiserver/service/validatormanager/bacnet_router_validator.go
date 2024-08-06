package validatormanager

import (
	"encoding/json"
	"errors"
	"github.com/hootrhino/rhilex/component/apiserver/dto"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/device"
	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
	"github.com/xuri/excelize/v2"
)

type BacnetRouterValidator struct {
}

func (b BacnetRouterValidator) Convert(pointDTO dto.DataPointCreateOrUpdateDTO) (model.MDataPoint, error) {
	point := model.MDataPoint{}
	config := device.BacnetRouterDataPointConfig{}
	err := mapstructure.Decode(pointDTO.Config, &config)
	if err != nil {
		return point, err
	}
	point.UUID = pointDTO.UUID
	point.Tag = pointDTO.Tag
	pointDTO.Alias = pointDTO.Alias
	pointDTO.Frequency = pointDTO.Frequency
	marshal, err := json.Marshal(pointDTO.Config)
	if err != nil {
		return point, err
	}
	point.Config = string(marshal)
	return point, nil
}

func (b BacnetRouterValidator) ParseImportFile(file *excelize.File) ([]model.MDataPoint, error) {
	rows, err := file.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}
	// 判断首行标头
	// tag,alias,objectType,objectId
	err1 := errors.New("'Invalid Sheet Header, must follow fixed format: 【tag,alias,objectType,objectId】")

	const MIN_LEN = 4
	if len(rows[0]) < MIN_LEN {
		return nil, err1
	}
	// 严格检查表结构 tag,alias,objectType,objectId
	if rows[0][0] != "tag" ||
		rows[0][1] != "alias" ||
		rows[0][2] != "objectType" ||
		rows[0][3] != "objectId" {
		return nil, err1
	}

	list := make([]model.MDataPoint, 0)
}

func (b BacnetRouterValidator) Export(file *excelize.File, list []model.MDataPoint) error {
	Headers := []string{
		"tag", "alias", "bacnetDeviceId", "objectType", "objectId",
	}
	cell, _ := excelize.CoordinatesToCellName(1, 1)
	file.SetSheetRow("Sheet1", cell, &Headers)
	if len(list) > 0 {
		for idx, record := range list[0:] {
			config := device.BacnetDataPointConfig{}
			err := json.Unmarshal([]byte(record.Config), &config)
			if err != nil {
				return err
			}
			Row := []any{
				record.Tag,
				record.Alias,
				config.BacnetDeviceId,
				config.ObjectType,
				config.ObjectId,
			}
			cell, _ = excelize.CoordinatesToCellName(1, idx+2)
			file.SetSheetRow("Sheet1", cell, &Row)
		}
	}
	return nil
}

func checkBacnetRouterPoint(point device.BacnetDataPointConfig) error {
	contains := lo.Contains(dto.ValidBacnetObjectType, point.ObjectType)
	if !contains {
		return errors.New("illegal objectType")
	}
	return nil
}
