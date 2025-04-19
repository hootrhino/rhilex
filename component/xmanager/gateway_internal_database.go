// Copyright (C) 2025 wwhai
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

package xmanager

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/driver/sqlite" // 使用 SQLite 作为示例
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InternalDatabase 提供数据库的生命周期管理
type InternalDatabase struct {
	db   *gorm.DB
	once sync.Once
}

// NewInternalDatabase 创建一个新的 InternalDatabase 实例
func NewInternalDatabase() *InternalDatabase {
	return &InternalDatabase{}
}

// Init 初始化数据库连接
func (idb *InternalDatabase) Init(dsn string) error {
	var err error
	idb.once.Do(func() {
		// 配置 Gorm 日志
		config := &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		}

		// 使用 SQLite 作为示例
		idb.db, err = gorm.Open(sqlite.Open(dsn), config)
		if err != nil {
			log.Printf("Failed to connect to database: %v", err)
			return
		}

		// 配置连接池
		sqlDB, err := idb.db.DB()
		if err != nil {
			log.Printf("Failed to configure database connection pool: %v", err)
			return
		}
		sqlDB.SetMaxOpenConns(10)                  // 最大打开连接数
		sqlDB.SetMaxIdleConns(5)                   // 最大空闲连接数
		sqlDB.SetConnMaxLifetime(30 * time.Minute) // 连接最大生命周期
	})
	return err
}

// GetDB 获取数据库实例
func (idb *InternalDatabase) GetDB() *gorm.DB {
	if idb.db == nil {
		log.Fatal("Database is not initialized. Call Init() first.")
	}
	return idb.db
}

// Migrate 执行数据库迁移
func (idb *InternalDatabase) Migrate(models ...interface{}) error {
	if idb.db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return idb.db.AutoMigrate(models...)
}

// Close 关闭数据库连接
func (idb *InternalDatabase) Close() error {
	if idb.db == nil {
		return fmt.Errorf("database is not initialized")
	}
	sqlDB, err := idb.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
