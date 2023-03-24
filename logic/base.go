package logic

import (
	"github.com/451008604/socketServerFrame/database"
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/network"
	"github.com/gogf/gf/v2/database/gdb"
)

type StaticModule struct {
	server iface.IServer        // 服务进程模块
	notify iface.INotifyManager // 广播管理模块
	sql    gdb.DB               // 数据库模块
}

var Module *StaticModule

func RegisterModule() {
	Module = &StaticModule{}
	Module.server = network.NewServer()
	Module.notify = network.NewNotifyManager()
	Module.sql = database.NewSqlDBModel()
}

func (s *StaticModule) Server() iface.IServer {
	return s.server
}

func (s *StaticModule) Notify() iface.INotifyManager {
	return s.notify
}

func (s *StaticModule) Sql() gdb.DB {
	return s.sql
}
