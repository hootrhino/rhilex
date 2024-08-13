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
	"github.com/xuri/excelize/v2"
	"strconv"
)

type SnmpValidator struct {
}

func (s SnmpValidator) Convert(dto dto.DataPointCreateOrUpdateDTO) (model.MDataPoint, error) {
	point := model.MDataPoint{}
	point.UUID = dto.UUID
	point.Tag = dto.Tag
	point.Alias = dto.Alias
	point.Frequency = dto.Frequency

	config := device.SnmpDataPointConfig{}
	err := mapstructure.Decode(dto.Config, &config)
	if err != nil {
		return point, err
	}

	err = checkSnmpDataPointConfig(config)
	if err != nil {
		return point, err
	}

	marshal, err := json.Marshal(dto.Config)
	if err != nil {
		return point, err
	}
	point.Config = string(marshal)
	return point, nil
}

func (s SnmpValidator) ParseImportFile(file *excelize.File) ([]model.MDataPoint, error) {
	// 读取表格
	rows, err := file.GetRows("Sheet1")
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

	list := make([]model.MDataPoint, 0)
	// name, alias, function, group, address
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		// oid,tag,alias,frequency
		oid := row[0]
		tag := row[1]
		alias := row[2]
		frequency, _ := strconv.ParseUint(row[3], 10, 64)

		config := device.SnmpDataPointConfig{
			Oid: oid,
		}
		err := checkSnmpDataPointConfig(config)
		if err != nil {
			return nil, err
		}

		configJson, _ := json.Marshal(config)
		point := model.MDataPoint{
			UUID:      utils.SnmpOidUUID(),
			Tag:       tag,
			Alias:     alias,
			Frequency: int(frequency),
			Config:    string(configJson),
		}
		list = append(list, point)
	}
	return list, nil
}

func (s SnmpValidator) Export(file *excelize.File, list []model.MDataPoint) error {
	Headers := []string{
		"oid", "tag", "alias", "frequency",
	}
	cell, _ := excelize.CoordinatesToCellName(1, 1)
	file.SetSheetRow("Sheet1", cell, &Headers)
	if len(list) > 0 {
		for idx, record := range list[0:] {
			config := device.SnmpDataPointConfig{}
			err := json.Unmarshal([]byte(record.Config), &config)
			if err != nil {
				return err
			}
			Row := []string{
				config.Oid, record.Tag, record.Alias, fmt.Sprintf("%d", record.Frequency),
			}
			cell, _ = excelize.CoordinatesToCellName(1, idx+2)
			file.SetSheetRow("Sheet1", cell, &Row)
		}
	}
	return nil
}

func checkSnmpDataPointConfig(config device.SnmpDataPointConfig) error {
	validate := validator.New()
	err := validate.Struct(&config)
	if err != nil {
		return err
	}
	return nil
}
