package service

import (
	"fmt"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/component/interdb"
)

func BatchInsertBacnetDataPoint(list []model.MBacnetDataPoint) error {
	m := model.MBacnetDataPoint{}
	return interdb.DB().Model(m).Create(list).Error
}

func InsertBacnetDataPoint(dataPoint model.MBacnetDataPoint) error {
	IgnoreUUID := dataPoint.UUID
	Count := int64(0)
	dataPoint.UUID = ""
	interdb.DB().Model(dataPoint).Where(dataPoint).Count(&Count)
	if Count > 0 {
		return fmt.Errorf("already exists same record:%s", IgnoreUUID)
	}
	dataPoint.UUID = IgnoreUUID
	return interdb.DB().Model(dataPoint).Create(&dataPoint).Error
}

func UpdateBacnetDataPoint(dataPoint model.MBacnetDataPoint) error {
	return interdb.DB().Model(&model.MBacnetDataPoint{}).
		Where("device_uuid=? AND uuid=?",
			dataPoint.DeviceUuid, dataPoint.UUID).
		Updates(dataPoint).Error
}

func BatchDeleteBacnetDataPoint(uuids []string, deviceUuid string) error {
	return interdb.DB().
		Where("uuid IN ? AND device_uuid=?", uuids, deviceUuid).
		Delete(&model.MBacnetDataPoint{}).Error
}

func DeleteAllBacnetDataPointByDeviceUuid(deviceUuid string) error {
	return interdb.DB().
		Where("device_uuid=?", deviceUuid).
		Delete(&model.MBacnetDataPoint{}).Error
}
