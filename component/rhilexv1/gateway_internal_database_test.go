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

package rhilex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// 示例模型
type TestModel struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:100"`
}

func TestInternalDatabase(t *testing.T) {
	// 创建 InternalDatabase 实例
	db := NewInternalDatabase()

	// 初始化数据库（使用 SQLite 内存数据库）
	err := db.Init("file::memory:?cache=shared")
	assert.NoError(t, err, "Database initialization should not fail")
	defer db.Close()

	// 测试 GetDB 方法
	t.Run("GetDB", func(t *testing.T) {
		gormDB := db.GetDB()
		assert.NotNil(t, gormDB, "GetDB should return a valid *gorm.DB instance")
	})

	// 测试 Migrate 方法
	t.Run("Migrate", func(t *testing.T) {
		err := db.Migrate(&TestModel{})
		assert.NoError(t, err, "Migrate should not fail")
	})

	// 测试插入数据
	t.Run("Insert Data", func(t *testing.T) {
		gormDB := db.GetDB()
		testData := TestModel{Name: "Test Name"}
		result := gormDB.Create(&testData)
		assert.NoError(t, result.Error, "Insert should not fail")
		assert.NotZero(t, testData.ID, "Inserted data should have a valid ID")
	})

	// 测试查询数据
	t.Run("Query Data", func(t *testing.T) {
		gormDB := db.GetDB()
		var result TestModel
		err := gormDB.First(&result, "name = ?", "Test Name").Error
		assert.NoError(t, err, "Query should not fail")
		assert.Equal(t, "Test Name", result.Name, "Queried data should match inserted data")
	})

	// 测试关闭数据库
	t.Run("Close Database", func(t *testing.T) {
		err := db.Close()
		assert.NoError(t, err, "Close should not fail")
	})
}
