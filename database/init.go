package database

import (
	"github.com/451008604/socketServerFrame/logs"
	"github.com/gogf/gf/v2/database/gdb"
)

func NewSqlDBModel() gdb.DB {
	dsn := ""
	sqlDB, err := gdb.New(gdb.ConfigNode{
		Link:                 dsn,
		CreatedAt:            "created_at", // 自动创建时间字段名称
		UpdatedAt:            "updated_at", // 自动更新时间字段名称
		DeletedAt:            "deleted_at", // 软删除时间字段名称
		TimeMaintainDisabled: false,        // 是否完全关闭时间更新特性，true时CreatedAt/UpdatedAt/DeletedAt都将失效
	})
	if err != nil {
		logs.PrintLogPanic(err)
	}
	// 开启调试模式
	sqlDB.SetDebug(true)

	return sqlDB
}

func NewRedisModel() {

}
