package sql

import (
	"context"
	"github.com/451008604/socketServerFrame/logs"
	"github.com/gogf/gf/v2/database/gdb"
)

type Module struct {
	gdb.DB
	Ctx context.Context
}

var DB *Module

func NewSqlDBModel() *Module {
	DB = &Module{Ctx: context.Background()}
	dsn := "mysql:root:Guohaoqin123.@tcp(ggghq.cn:6606)/game_library?charset=utf8mb4&parseTime=true&loc=Local"
	var err error
	DB.DB, err = gdb.New(gdb.ConfigNode{
		Link:                 dsn,
		CreatedAt:            "created_at", // 自动创建时间字段名称
		UpdatedAt:            "updated_at", // 自动更新时间字段名称
		DeletedAt:            "deleted_at", // 软删除时间字段名称
		TimeMaintainDisabled: false,        // 是否完全关闭时间更新特性，true时CreatedAt/UpdatedAt/DeletedAt都将失效
	})
	logs.PrintLogPanic(err)
	// 开启调试模式
	DB.SetDebug(true)

	return DB
}
