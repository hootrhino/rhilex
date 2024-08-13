package validatormanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hootrhino/rhilex/component/apiserver/dto"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/device"
	"github.com/hootrhino/rhilex/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/xuri/excelize/v2"
	"strconv"
	"strings"
)

type SiemensPLCValidator struct {
}

func (s SiemensPLCValidator) Convert(pointDTO dto.DataPointCreateOrUpdateDTO) (model.MDataPoint, error) {
	point := model.MDataPoint{}

	config := device.SiemensS1200DataPointConfig{}
	err := mapstructure.Decode(pointDTO.Config, &config)
	if err != nil {
		return point, err
	}

	err = checkSiemensS1200DataPointConfig(config)
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
	return point, err
}

func (s SiemensPLCValidator) ParseImportFile(file *excelize.File) ([]model.MDataPoint, error) {
	rows, err := file.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}
	// 判断首行标头
	//
	err = errors.New("invalid Sheet Header")
	if len(rows[0]) < 7 {
		return nil, err
	}
	// Address Tag Alias Type Order Frequency
	if strings.ToLower(rows[0][0]) != "address" ||
		strings.ToLower(rows[0][1]) != "tag" ||
		strings.ToLower(rows[0][2]) != "alias" ||
		strings.ToLower(rows[0][3]) != "type" ||
		strings.ToLower(rows[0][4]) != "order" ||
		strings.ToLower(rows[0][5]) != "weight" ||
		strings.ToLower(rows[0][6]) != "frequency" {
		return nil, err
	}

	list := make([]model.MDataPoint, 0)
	// Address Tag Alias Type Order Frequency
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		SiemensAddress := row[0]
		Tag := row[1]
		Alias := row[2]
		Type := row[3]
		Order := row[4]
		Weight, _ := strconv.ParseFloat(row[5], 32)
		if Weight == 0 {
			Weight = 1 // 防止解析异常的时候系数0
		}
		frequency, _ := strconv.ParseUint(row[6], 10, 64)

		config := device.SiemensS1200DataPointConfig{}
		config.SiemensAddress = SiemensAddress
		config.DataBlockType = Type
		config.DataBlockOrder = utils.GetDefaultDataOrder(Type, Order)
		config.Weight = &Weight

		err = checkSiemensS1200DataPointConfig(config)
		if err != nil {
			return nil, err
		}

		marshal, err := json.Marshal(config)
		if err != nil {
			return nil, err
		}
		point := model.MDataPoint{
			UUID:      utils.SiemensPointUUID(),
			Tag:       Tag,
			Alias:     Alias,
			Frequency: int(frequency),
			Config:    string(marshal),
		}

		list = append(list, point)
	}
	return list, nil
}

func (s SiemensPLCValidator) Export(file *excelize.File, list []model.MDataPoint) error {
	Headers := []string{
		"address", "tag", "alias", "type", "order", "weight", "frequency",
	}
	cell, _ := excelize.CoordinatesToCellName(1, 1)
	file.SetSheetRow("Sheet1", cell, &Headers)
	for idx, record := range list {
		pointConfig := device.SiemensS1200DataPointConfig{}
		err := json.Unmarshal([]byte(record.Config), &pointConfig)
		if err != nil {
			return err
		}
		Row := []string{
			pointConfig.SiemensAddress,
			record.Tag,
			record.Alias,
			pointConfig.DataBlockType,
			pointConfig.DataBlockOrder,
			fmt.Sprintf("%f", *pointConfig.Weight),
			fmt.Sprintf("%d", record.Frequency),
		}
		cell, _ = excelize.CoordinatesToCellName(1, idx+2)
		file.SetSheetRow("Sheet1", cell, &Row)
	}
	return nil
}

func checkSiemensS1200DataPointConfig(config device.SiemensS1200DataPointConfig) error {
	_, err := utils.ParseSiemensDB(config.SiemensAddress)
	if err != nil {
		return err
	}

	switch config.DataBlockType {
	case "I", "Q", "BYTE":
		if config.DataBlockOrder != "A" {
			return fmt.Errorf("invalid '%s' order '%s'", config.DataBlockType, config.DataBlockOrder)
		}
	case "SHORT", "USHORT", "INT16", "UINT16":
		if !utils.SContains([]string{"AB", "BA"}, config.DataBlockOrder) {
			return fmt.Errorf("invalid '%s' order '%s'", config.DataBlockType, config.DataBlockOrder)
		}
	case "RAW", "INT", "INT32", "UINT", "UINT32", "FLOAT", "UFLOAT":
		if !utils.SContains([]string{"ABCD", "DCBA", "CDAB"}, config.DataBlockOrder) {
			return fmt.Errorf("invalid '%s' order '%s'", config.DataBlockType, config.DataBlockOrder)
		}
	default:
		return fmt.Errorf("invalid '%s' order '%s'", config.DataBlockType, config.DataBlockOrder)
	}
	if config.Weight == nil {
		return fmt.Errorf("invalid Weight value: nil value")
	}
	return nil
}
