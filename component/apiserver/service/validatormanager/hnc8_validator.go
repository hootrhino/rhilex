package validatormanager

import (
	"github.com/hootrhino/rhilex/component/apiserver/dto"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/xuri/excelize/v2"
)

type Hnc8Validator struct {
}

func (h Hnc8Validator) Convert(dto dto.DataPointCreateOrUpdateDTO) (model.MDataPoint, error) {
	//TODO implement me
	panic("implement me")
}

func (h Hnc8Validator) ParseImportFile(file *excelize.File) ([]model.MDataPoint, error) {
	//TODO implement me
	panic("implement me")
}

func (h Hnc8Validator) Export(file *excelize.File, list []model.MDataPoint) error {
	//TODO implement me
	panic("implement me")
}
