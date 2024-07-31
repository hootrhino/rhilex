package validatormanager

import (
	"github.com/hootrhino/rhilex/component/apiserver/dto"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/xuri/excelize/v2"
)

type BacnetIpValidator struct {
}

func (b BacnetIpValidator) Validate(dto dto.DataPointCreateOrUpdateDTO) (model.MDataPoint, error) {
	//TODO implement me
	panic("implement me")
}

func (b BacnetIpValidator) ParseImportFile(file *excelize.File) ([]model.MDataPoint, error) {
	//TODO implement me
	panic("implement me")
}

func (b BacnetIpValidator) Export(file *excelize.File, list []model.MDataPoint) error {
	//TODO implement me
	panic("implement me")
}
