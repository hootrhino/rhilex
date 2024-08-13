package validatormanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator"
	"github.com/hootrhino/rhilex/component/apiserver/dto"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/device"
	"github.com/hootrhino/rhilex/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
	"github.com/xuri/excelize/v2"
	"strconv"
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
	err = checkBacnetRouterPoint(config)
	if err != nil {
		return point, err
	}
	point.UUID = pointDTO.UUID
	point.Tag = pointDTO.Tag
	point.Alias = pointDTO.Alias
	point.Frequency = pointDTO.Frequency
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
	err = errors.New("'Invalid Sheet Header, must follow fixed format: 【tag,alias,objectType,objectId】")

	const MIN_LEN = 4
	if len(rows[0]) < MIN_LEN {
		return nil, err
	}
	// 严格检查表结构 tag,alias,objectType,objectId
	if rows[0][0] != "tag" ||
		rows[0][1] != "alias" ||
		rows[0][2] != "objectType" ||
		rows[0][3] != "objectId" {
		return nil, err
	}

	list := make([]model.MDataPoint, 0)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < MIN_LEN {
			msg := fmt.Sprintf("illegal data, the data cell of row %d less than %d", i+1, MIN_LEN)
			return nil, errors.New(msg)
		}
		tag := row[0]
		alias := row[1]
		objectType := row[2]
		objectId, _ := strconv.ParseInt(row[3], 10, 32)

		config := device.BacnetRouterDataPointConfig{
			ObjectType: objectType,
			ObjectId:   uint32(objectId),
		}
		err = checkBacnetRouterPoint(config)
		if err != nil {
			return nil, err
		}
		marshal, err := json.Marshal(config)
		if err != nil {
			return nil, err
		}

		point := model.MDataPoint{
			UUID:   utils.BacnetPointUUID(),
			Tag:    tag,
			Alias:  alias,
			Config: string(marshal),
		}
		list = append(list, point)
	}
	return list, nil
}

func (b BacnetRouterValidator) Export(file *excelize.File, list []model.MDataPoint) error {
	Headers := []string{
		"tag", "alias", "objectType", "objectId",
	}
	cell, _ := excelize.CoordinatesToCellName(1, 1)
	file.SetSheetRow("Sheet1", cell, &Headers)
	if len(list) > 0 {
		for idx, record := range list[0:] {
			config := device.BacnetRouterDataPointConfig{}
			err := json.Unmarshal([]byte(record.Config), &config)
			if err != nil {
				return err
			}
			Row := []any{
				record.Tag,
				record.Alias,
				config.ObjectType,
				config.ObjectId,
			}
			cell, _ = excelize.CoordinatesToCellName(1, idx+2)
			file.SetSheetRow("Sheet1", cell, &Row)
		}
	}
	return nil
}

func checkBacnetRouterPoint(point device.BacnetRouterDataPointConfig) error {
	contains := lo.Contains(dto.ValidBacnetObjectType, point.ObjectType)
	if !contains {
		return errors.New("illegal objectType")
	}
	validate := validator.New()
	err := validate.Struct(&point)
	if err != nil {
		return err
	}
	return nil
}
