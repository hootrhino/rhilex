// Copyright (C) 2024 wwhai
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
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package lostcache

import (
	"fmt"

	"gorm.io/gorm"
)

func CreateLostDataTable(deviceId string) {
	LostCacheDb().Transaction(func(tx *gorm.DB) error {
		sql1 := `
		CREATE TABLE IF NOT EXISTS "lost_data_%s" (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			target_id TEXT NOT NULL,
			data TEXT NOT NULL,
			UNIQUE (target_id, created_at)
		);`
		tx.Exec(fmt.Sprintf(sql1, deviceId))

		sql2 := `
CREATE TRIGGER IF NOT EXISTS limit_lost_data_%s
AFTER INSERT ON "lost_data_%s"
WHEN ((SELECT COUNT(*) FROM "lost_data_%s") / 100) * 100 = (SELECT COUNT(*) FROM "lost_data_%s")
AND (SELECT COUNT(*) FROM "lost_data_%s") > 1000
BEGIN
    DELETE FROM "lost_data_%s"
    WHERE id IN (
        SELECT id FROM "lost_data_%s"
        ORDER BY id ASC
        LIMIT 100
    );
END;

`
		tx.Exec(fmt.Sprintf(sql2, deviceId, deviceId, deviceId,
			deviceId, deviceId, deviceId, deviceId))
		return tx.Error
	})

}
func DeleteLostDataTable(deviceId string) {
	sql := `DROP TABLE IF EXISTS "lost_data_%s";`
	LostCacheDb().Exec(fmt.Sprintf(sql, deviceId))
}

/**
 * Save
 *
 */
func SaveLostCacheData(deviceId string, data CacheDataDto) error {
	LostCacheDb().Table(fmt.Sprintf("lost_data_%s", deviceId)).Create(&CacheData{
		TargetId: data.TargetId,
		Data:     data.Data,
	})
	return LostCacheDb().Error
}

/**
 * Get
 *
 */
func GetLostCacheData(deviceId string) ([]CacheDataDto, error) {
	dataDto := []CacheDataDto{}
	LostCacheDb().Table(fmt.Sprintf("lost_data_%s", deviceId)).Where("target_id=?", deviceId).Find(&dataDto)
	return dataDto, LostCacheDb().Error

}

/**
 * Delete Lost Cache Data
 *
 */
func DeleteLostCacheData(deviceId string, dbId uint) {
	LostCacheDb().Table(fmt.Sprintf("lost_data_%s", deviceId)).Where("id=?", dbId).Delete(&CacheData{})
}

/**
 * Clear
 *
 */
func ClearLostCacheData(deviceId string) {
	LostCacheDb().Table(fmt.Sprintf("lost_data_%s", deviceId)).Where("target_id=?", deviceId).Delete(&CacheData{})
}
