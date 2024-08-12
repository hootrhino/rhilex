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

type ModbusMasterPointVo struct {
	UUID          string   `json:"uuid,omitempty"`
	DeviceUUID    string   `json:"device_uuid"`
	Tag           string   `json:"tag"`
	Alias         string   `json:"alias"`
	Function      *int     `json:"function"`
	SlaverId      *byte    `json:"slaverId"`
	Address       *uint16  `json:"address"`
	Frequency     *int64   `json:"frequency"`
	Quantity      *uint16  `json:"quantity"`
	DataType      string   `json:"dataType"`      // 数据类型
	DataOrder     string   `json:"dataOrder"`     // 字节序
	Weight        *float64 `json:"weight"`        // 权重
	Status        int      `json:"status"`        // 运行时数据
	LastFetchTime uint64   `json:"lastFetchTime"` // 运行时数据
	Value         string   `json:"value"`         // 运行时数据
	ErrMsg        string   `json:"errMsg"`        // 运行时数据

}

type ModbusMasterValidator struct {
}

func (v ModbusMasterValidator) Convert(dto dto.DataPointCreateOrUpdateDTO) (model.MDataPoint, error) {
	point := model.MDataPoint{}
	point.Tag = dto.Tag
	point.Alias = dto.Alias
	point.Frequency = dto.Frequency
	config := device.ModbusMasterDataPointConfig{}
	err := mapstructure.Decode(dto.Config, &config)
	if err != nil {
		return point, err
	}
	validate := validator.New()
	err = validate.Struct(&config)
	if err != nil {
		return point, err
	}
	point.SetConfig(dto.Config)
	return point, nil
}

func (v ModbusMasterValidator) ParseImportFile(excelFile *excelize.File) ([]model.MDataPoint, error) {
	sheetName := "sheet1"
	// 读取表格
	rows, err := excelFile.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	// 判断首行标头
	// tag, alias, function, frequency, slaverId, address, quality
	err = errors.New(" Invalid Sheet Header")
	if len(rows[0]) < 10 {
		return nil, err
	}

	// 严格检查表结构
	if rows[0][0] != "tag" ||
		rows[0][1] != "alias" ||
		rows[0][2] != "function" ||
		rows[0][3] != "frequency" ||
		rows[0][4] != "slaverId" ||
		rows[0][5] != "address" ||
		rows[0][6] != "quality" ||
		rows[0][7] != "type" ||
		rows[0][8] != "order" ||
		rows[0][9] != "weight" {
		return nil, err
	}

	list := make([]model.MDataPoint, 0)
	// tag, alias, function, frequency, slaverId, address, quality
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		tag := row[0]
		alias := row[1]
		function, _ := strconv.ParseUint(row[2], 10, 8)
		frequency, _ := strconv.ParseUint(row[3], 10, 64)
		slaverId, _ := strconv.ParseUint(row[4], 10, 8)
		address, _ := strconv.ParseUint(row[5], 10, 16)
		quantity, _ := strconv.ParseUint(row[6], 10, 16)
		Type := row[7]
		Order := row[8]
		Weight, _ := strconv.ParseFloat(row[9], 32)
		if Weight == 0 {
			Weight = 1 // 防止解析异常的时候系数0
		}
		Function := int(function)
		SlaverId := byte(slaverId)
		Address := uint16(address)
		Frequency := int64(frequency)
		Quantity := uint16(quantity)

		if err := checkModbusMasterDataPoints(ModbusMasterPointVo{
			Tag:       tag,
			Alias:     alias,
			Function:  &Function,
			SlaverId:  &SlaverId,
			Address:   &Address,
			Frequency: &Frequency, //ms
			Quantity:  &Quantity,
			DataType:  Type,
			DataOrder: utils.GetDefaultDataOrder(Type, Order),
			Weight:    &Weight,
		}); err != nil {
			return nil, err
		}
		//
		config := device.ModbusMasterDataPointConfig{
			Function:  Function,
			SlaverId:  SlaverId,
			Address:   Address,
			Quantity:  Quantity,
			DataType:  Type,
			DataOrder: utils.GetDefaultDataOrder(Type, Order),
			Weight:    Weight,
		}
		configJson, _ := json.Marshal(config)
		dataPoint := model.MDataPoint{
			UUID:      utils.ModbusPointUUID(),
			Tag:       tag,
			Alias:     alias,
			Frequency: int(frequency),
			Config:    string(configJson),
		}
		list = append(list, dataPoint)
	}
	return list, nil
}

func (v ModbusMasterValidator) Export(file *excelize.File, list []model.MDataPoint) error {
	Headers := []string{
		"tag", "alias",
		"function", "frequency",
		"slaverId", "address",
		"quality", "type",
		"order", "weight",
	}
	cell, _ := excelize.CoordinatesToCellName(1, 1)
	file.SetSheetRow("Sheet1", cell, &Headers)
	for idx, record := range list[0:] {
		config := device.ModbusMasterDataPointConfig{}
		err := json.Unmarshal([]byte(record.Config), &config)
		if err != nil {
			return err
		}
		Row := []string{
			record.Tag,
			record.Alias,
			fmt.Sprintf("%d", config.Function),
			strconv.Itoa(record.Frequency),
			fmt.Sprintf("%d", config.SlaverId),
			fmt.Sprintf("%d", config.Address),
			fmt.Sprintf("%d", config.Quantity),
			config.DataType,
			config.DataOrder,
			fmt.Sprintf("%f", config.Weight),
		}
		cell, _ = excelize.CoordinatesToCellName(1, idx+2)
		file.SetSheetRow("Sheet1", cell, &Row)
	}
	return nil
}

/*
*
* 检查点位合法性
*
 */
func checkModbusMasterDataPoints(M ModbusMasterPointVo) error {
	if M.Quantity == nil {
		return fmt.Errorf("'Missing required param 'quantity'")
	}
	switch M.DataType {
	case "UTF8":
		if (*M.Quantity * uint16(2)) > 255 {
			return fmt.Errorf("'Invalid 'UTF8' Length '%d'", (*M.Quantity * uint16(2)))
		}
		if !utils.SContains([]string{"BIG_ENDIAN", "LITTLE_ENDIAN"}, M.DataOrder) {
			return fmt.Errorf("'Invalid '%s' order '%s'", M.DataType, M.DataOrder)
		}
	case "I", "Q", "BYTE":
		if M.DataOrder != "A" {
			return fmt.Errorf("'Invalid '%s' order '%s'", M.DataType, M.DataOrder)
		}
	case "SHORT", "USHORT", "INT16", "UINT16":
		if !utils.SContains([]string{"AB", "BA"}, M.DataOrder) {
			return fmt.Errorf("'Invalid '%s' order '%s'", M.DataType, M.DataOrder)
		}
	case "RAW", "INT", "INT32", "UINT", "UINT32", "FLOAT", "UFLOAT":
		if !utils.SContains([]string{"ABCD", "DCBA", "CDAB"}, M.DataOrder) {
			return fmt.Errorf("'Invalid '%s' order '%s'", M.DataType, M.DataOrder)
		}
	default:
		return fmt.Errorf("'Invalid '%s' order '%s'", M.DataType, M.DataOrder)
	}
	if M.Weight == nil {
		return fmt.Errorf("'Invalid Weight value:%d", M.Weight)
	}
	return nil
}
