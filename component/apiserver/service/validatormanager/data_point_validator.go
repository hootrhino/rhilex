package validatormanager

import (
	"errors"
	"github.com/hootrhino/rhilex/component/apiserver/dto"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/component/apiserver/service/validatormanager/modbus"
	"github.com/hootrhino/rhilex/typex"
	"github.com/xuri/excelize/v2"
)

type Validator interface {
	Validate(dto dto.DataPointCreateOrUpdateDTO) (model.MDataPoint, error)
	ParseImportFile(file *excelize.File) ([]model.MDataPoint, error)
	Export(file *excelize.File, list []model.MDataPoint) error
}

func GetByType(protocol string) (Validator, error) {
	dt := typex.DeviceType(protocol)
	switch dt {
	case typex.GENERIC_MODBUS_MASTER:
		return modbus.ModbusValidator{}, nil
	case typex.GENERIC_MODBUS_SLAVER:
		return modbus.ModbusValidator{}, nil
	default:
		return nil, errors.New("valid protocol data point validator not found")
	}
}
