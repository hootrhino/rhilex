package validatormanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hootrhino/rhilex/component/apiserver/dto"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/device"
	"github.com/hootrhino/rhilex/utils"
	"github.com/xuri/excelize/v2"
	"strconv"
)

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

type SnmpValidator struct {
}

func (s SnmpValidator) Validate(dto dto.DataPointCreateOrUpdateDTO) (model.MDataPoint, error) {
	point := model.MDataPoint{}
	if dto.Tag == "" {
		return point, fmt.Errorf("'Missing required param 'name'")
	}
	if len(dto.Tag) > 256 {
		return point, fmt.Errorf("'Tag length must range of 1-256")
	}
	if dto.Alias == "" {
		return point, fmt.Errorf("'Missing required param 'alias'")
	}
	if len(dto.Alias) > 256 {
		return point, fmt.Errorf("'Alias length must range of 1-256")
	}
	point.UUID = dto.UUID
	point.Tag = dto.Tag
	point.Alias = dto.Alias
	point.Frequency = dto.Frequency
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
		if err := checkSnmpOids(SnmpOidVo{
			Oid:       oid,
			Tag:       tag,
			Alias:     alias,
			Frequency: &frequency,
		}); err != nil {
			return nil, err
		}

		config := device.SnmpDataPointConfig{
			Oid: oid,
		}
		configJson, _ := json.Marshal(config)
		model := model.MDataPoint{
			UUID:      utils.SnmpOidUUID(),
			Tag:       tag,
			Alias:     alias,
			Frequency: int(frequency),
			Config:    string(configJson),
		}
		list = append(list, model)
	}
	return list, nil
}

func (s SnmpValidator) Export(file *excelize.File, list []model.MDataPoint) error {
	Headers := []string{
		"oid", "tag", "alias", "frequency",
	}
	cell, _ := excelize.CoordinatesToCellName(1, 1)
	file.SetSheetRow("Sheet1", cell, &Headers)
	if len(list) > 1 {
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
