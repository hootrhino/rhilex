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

	"github.com/hootrhino/rhilex/component/interdb"
	"github.com/hootrhino/rhilex/component/rhilex_api_server/model"
)

/*
*
* Snmp点位表管理
*
 */
// InsertSnmpOid 插入Snmp点位表
func InsertSnmpOids(list []model.MSnmpOid) error {
	m := model.MSnmpOid{}
	return interdb.DB().Model(m).Create(list).Error
}

// InsertSnmpOid 插入Snmp点位表
func InsertSnmpOid(P model.MSnmpOid) error {
	IgnoreUUID := P.UUID
	Count := int64(0)
	P.UUID = ""
	interdb.DB().Model(P).Where(P).Count(&Count)
	if Count > 0 {
		return fmt.Errorf("already exists same record:%s", IgnoreUUID)
	}
	P.UUID = IgnoreUUID
	return interdb.DB().Model(P).Create(&P).Error
}

// DeleteSnmpOidByDevice 删除Snmp点位与设备
func DeleteSnmpOidByDevice(uuids []string, deviceUuid string) error {
	return interdb.DB().
		Where("uuid IN ? AND device_uuid=?", uuids, deviceUuid).
		Delete(&model.MSnmpOid{}).Error
}

// DeleteAllSnmpOidByDevice 删除Snmp点位与设备
func DeleteAllSnmpOidByDevice(deviceUuid string) error {
	return interdb.DB().
		Where("device_uuid=?", deviceUuid).
		Delete(&model.MSnmpOid{}).Error
}

// 更新DataSchema
func UpdateSnmpOid(MSnmpOid model.MSnmpOid) error {
	return interdb.DB().Model(model.MSnmpOid{}).
		Where("device_uuid=? AND uuid=?",
			MSnmpOid.DeviceUuid, MSnmpOid.UUID).
		Updates(MSnmpOid).Error
}
