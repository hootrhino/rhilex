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

package interdb

import (
	"runtime"

	core "github.com/hootrhino/rhilex/config"
	"github.com/hootrhino/rhilex/typex"

	"github.com/hootrhino/rhilex/glogger"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

const __INTERNAL_DB_PATH string = "./rhilex.db?cache=shared&mode=rwc"

var __InternalSqlite *SqliteDAO

/*
*
* 初始化DAO
*
 */
func InitInterDb(engine typex.Rhilex) error {
	__InternalSqlite = &SqliteDAO{name: "Sqlite3", engine: engine}

	var err error
	if core.GlobalConfig.DebugMode {
		__InternalSqlite.db, err = gorm.Open(sqlite.Open(__INTERNAL_DB_PATH), &gorm.Config{
			Logger:                 logger.Default.LogMode(logger.Info),
			SkipDefaultTransaction: false,
		})
	} else {
		__InternalSqlite.db, err = gorm.Open(sqlite.Open(__INTERNAL_DB_PATH), &gorm.Config{
			Logger:                 logger.Default.LogMode(logger.Error),
			SkipDefaultTransaction: false,
		})
	}
	if err != nil {
		glogger.GLogger.Fatal(err)
	}
	schema.RegisterSerializer("custom", CustomTypeSerializer{})
	__InternalSqlite.db.Exec("VACUUM;")
	return err
}

/*
*
* 停止
*
 */
func StopInterDb() {
	__InternalSqlite.db = nil
	runtime.GC()
}

/*
*
* 返回数据库查询句柄
*
 */
func InterDb() *gorm.DB {
	return __InternalSqlite.db
}

/*
*
* 注册数据模型
*
 */
func InterDbRegisterModel(dist ...any) {
	__InternalSqlite.db.AutoMigrate(dist...)
}

// GetDB 获取数据库连接实例
func GetDB() *gorm.DB {
	return __InternalSqlite.db
}

// Create 创建数据
func Create(value any) error {
	return __InternalSqlite.db.Create(value).Error
}

// Read 查询数据
func Read(dest any, conditions ...any) error {
	return __InternalSqlite.db.Find(dest, conditions...).Error
}

// ReadWithPagination 分页查询数据
func ReadWithPagination(dest any, page, pageSize int, conditions ...any) error {
	offset := (page - 1) * pageSize
	return __InternalSqlite.db.Where(conditions).Offset(offset).Limit(pageSize).Find(dest).Error
}

// Update 更新数据
func Update(value any, conditions ...any) error {
	return __InternalSqlite.db.Where(conditions).Updates(value).Error
}

// Delete 删除数据
func Delete(value any, conditions ...any) error {
	return __InternalSqlite.db.Where(conditions).Delete(value).Error
}

// CreateTable 创建表
func CreateTable(models ...any) error {
	return __InternalSqlite.db.AutoMigrate(models...)
}

// DropTable 删除表
func DropTable(models ...any) error {
	return __InternalSqlite.db.Migrator().DropTable(models...)
}

// CreateDatabase 创建数据库
func CreateDatabase(dbName string) error {
	return __InternalSqlite.db.Exec("CREATE DATABASE IF NOT EXISTS " + dbName).Error
}

// DropDatabase 删除数据库
func DropDatabase(dbName string) error {
	return __InternalSqlite.db.Exec("DROP DATABASE IF EXISTS " + dbName).Error
}
