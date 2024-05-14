package datacenter

import (
	core "github.com/hootrhino/rhilex/config"
	"github.com/hootrhino/rhilex/typex"

	"github.com/hootrhino/rhilex/glogger"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const __DEFAULT_DB_PATH string = "./rhilex_internal_datacenter.db"

/*
*
* Sqlite 数据持久层
*
 */
type SqliteDb struct {
	engine typex.Rhilex
	name   string   // 框架可以根据名称来选择不同的数据库驱动,为以后扩展准备
	db     *gorm.DB // Sqlite 驱动
}

/*
*
* 初始化DAO
*
 */
func InitSqliteDb(engine typex.Rhilex) *SqliteDb {
	__Sqlite := &SqliteDb{name: "Sqlite3", engine: engine}

	var err error
	if core.GlobalConfig.AppDebugMode {
		__Sqlite.db, err = gorm.Open(sqlite.Open(__DEFAULT_DB_PATH), &gorm.Config{
			Logger:                 logger.Default.LogMode(logger.Info),
			SkipDefaultTransaction: false,
		})
	} else {
		__Sqlite.db, err = gorm.Open(sqlite.Open(__DEFAULT_DB_PATH), &gorm.Config{
			Logger:                 logger.Default.LogMode(logger.Error),
			SkipDefaultTransaction: false,
		})
	}
	if err != nil {
		glogger.GLogger.Fatal(err)
	}
	return __Sqlite
}
