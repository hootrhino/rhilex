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
	"github.com/hootrhino/rhilex/component/internotify"
)

/*
*
* InsertInternalNotifies
*
 */
func InsertInternalNotify(m internotify.MInternalNotify) error {
	return internotify.InterNotifyDb().Model(&m).Save(&m).Error
}

/*
*
* 右上角
*
 */
func AllInternalNotifiesHeader() []internotify.MInternalNotify {
	m := []internotify.MInternalNotify{}
	internotify.InterNotifyDb().Table("m_internal_notifies").Where("status=1").Limit(6).Find(&m)
	return m
}

/*
*
* 所有列表
*
 */
func AllInternalNotifies() []internotify.MInternalNotify {
	m := []internotify.MInternalNotify{}
	internotify.InterNotifyDb().Table("m_internal_notifies").Where("status=1").Limit(100).Find(&m)
	return m
}

/*
*
* 清空表
*
 */
func ClearInternalNotifies() error {
	return internotify.InterNotifyDb().Exec("DELETE FROM m_internal_notifies;VACUUM;").Error
}

/*
*
* 点击已读
*
 */
func ReadInternalNotifies(uuid string) error {
	return internotify.InterNotifyDb().Table("m_internal_notifies").
		Where("uuid=?", uuid).Delete(&internotify.MInternalNotify{}).Error
}
