// Copyright (C) 2023 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package service

import (
	"fmt"

	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/component/interdb"
)

/*
*
* Modbus点位表管理
*
 */
// InsertModbusPointPosition 插入modbus点位表
func InsertModbusPointPositions(list []model.MModbusDataPoint) error {
	m := model.MModbusDataPoint{}
	return interdb.InterDb().Model(m).Create(list).Error
}

// InsertModbusPointPosition 插入modbus点位表
func InsertModbusPointPosition(P model.MModbusDataPoint) error {
	IgnoreUUID := P.UUID
	Count := int64(0)
	P.UUID = ""
	interdb.InterDb().Model(P).Where(P).Count(&Count)
	if Count > 0 {
		return fmt.Errorf("already exists same record:%s", IgnoreUUID)
	}
	P.UUID = IgnoreUUID
	return interdb.InterDb().Model(P).Create(&P).Error
}

// DeleteModbusPointByDevice 删除modbus点位与设备
func DeleteModbusPointByDevice(uuids []string, deviceUuid string) error {
	return interdb.InterDb().
		Where("uuid IN ? AND device_uuid=?", uuids, deviceUuid).
		Delete(&model.MModbusDataPoint{}).Error
}

// DeleteAllModbusPointByDevice 删除modbus点位与设备
func DeleteAllModbusPointByDevice(deviceUuid string) error {
	return interdb.InterDb().
		Where("device_uuid=?", deviceUuid).
		Delete(&model.MModbusDataPoint{}).Error
}

// 更新DataSchema
func UpdateModbusPoint(MModbusDataPoint model.MModbusDataPoint) error {
	return interdb.InterDb().Model(model.MModbusDataPoint{}).
		Where("device_uuid=? AND uuid=?",
			MModbusDataPoint.DeviceUuid, MModbusDataPoint.UUID).
		Updates(MModbusDataPoint).Error
}
