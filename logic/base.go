package logic

import (
	"fmt"
	"github.com/451008604/socketServerFrame/api"
	"github.com/451008604/socketServerFrame/database"
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logs"
	"github.com/451008604/socketServerFrame/modules"
	"github.com/451008604/socketServerFrame/network"
	"github.com/gogf/gf/v2/database/gdb"
	"runtime"
	"time"
)

type StaticModule struct {
	serverTCP iface.IServer        // 服务进程模块
	serverWS  iface.IServer        // 服务进程模块
	notify    iface.INotifyManager // 广播管理模块
	sql       gdb.DB               // 数据库模块
}

var Module *StaticModule

func RegisterModule() {
	Module = &StaticModule{}
	Module.serverTCP = network.NewServerTCP()
	Module.serverWS = network.NewServerWS()
	Module.notify = network.NewNotifyManager()
	Module.sql = database.NewSqlDBModel()

	// 注册路由
	api.RegisterRouter(network.GetInstanceMsgHandler())

	// 连接建立时
	network.GetInstanceConnManager().OnConnOpen(onConnectionOpen)
	// 连接断开后
	network.GetInstanceConnManager().OnConnClose(onConnectionClose)

	go func() {
		goroutineNum := 0
		for range time.Tick(time.Second * 1) {
			if temp := runtime.NumGoroutine(); temp != goroutineNum {
				goroutineNum = temp
				logs.PrintLogInfo(fmt.Sprintf("当前线程数：%v\t当前连接数量：%v", goroutineNum, network.GetInstanceConnManager().Len()))
			}
		}
	}()

	// 启动计时器
	modules.StartTicker()
}

func (s *StaticModule) ServerTCP() iface.IServer {
	return s.serverTCP
}

func (s *StaticModule) ServerWS() iface.IServer {
	return s.serverWS
}

func (s *StaticModule) Notify() iface.INotifyManager {
	return s.notify
}

func (s *StaticModule) Sql() gdb.DB {
	return s.sql
}

// 建立连接时
func onConnectionOpen(conn iface.IConnection) {
	conn.SetProperty("Client", conn.RemoteAddrStr())
}

// 断开连接时
func onConnectionClose(conn iface.IConnection) {
	logs.PrintLogInfo(fmt.Sprintf("客户端%v下线", conn.GetProperty("Client")))
}
