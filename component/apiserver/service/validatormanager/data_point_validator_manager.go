package validatormanager

import (
	"errors"
	"github.com/hootrhino/rhilex/component/apiserver/dto"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/typex"
	"github.com/xuri/excelize/v2"
)

type Validator interface {
	Convert(dto dto.DataPointCreateOrUpdateDTO) (model.MDataPoint, error)
	ParseImportFile(file *excelize.File) ([]model.MDataPoint, error)
	Export(file *excelize.File, list []model.MDataPoint) error
}

func GetByType(protocol string) (Validator, error) {
	dt := typex.DeviceType(protocol)
	switch dt {
	case typex.GENERIC_MODBUS_MASTER:
		return ModbusMasterValidator{}, nil
	case typex.GENERIC_BACNET_IP:
		return BacnetIpValidator{}, nil
	case typex.BACNET_ROUTER_GW:
		return BacnetRouterValidator{}, nil
	case typex.SIEMENS_PLC:
		return SiemensPLCValidator{}, nil
	case typex.GENERIC_SNMP:
		return SnmpValidator{}, nil
	default:
		return nil, errors.New("valid protocol data point validator not found")
	}
}
