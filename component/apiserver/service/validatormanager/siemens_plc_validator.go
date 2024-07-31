package validatormanager

import (
	"github.com/hootrhino/rhilex/component/apiserver/dto"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/xuri/excelize/v2"
)

type SiemensPLCValidator struct {
}

func (s SiemensPLCValidator) Validate(dto dto.DataPointCreateOrUpdateDTO) (model.MDataPoint, error) {
	//TODO implement me
	panic("implement me")
}

func (s SiemensPLCValidator) ParseImportFile(file *excelize.File) ([]model.MDataPoint, error) {
	//TODO implement me
	panic("implement me")
}

func (s SiemensPLCValidator) Export(file *excelize.File, list []model.MDataPoint) error {
	//TODO implement me
	panic("implement me")
}
