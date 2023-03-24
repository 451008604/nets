package main

import (
	"fmt"
	"github.com/451008604/socketServerFrame/api"
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logic"
	"github.com/451008604/socketServerFrame/logs"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"runtime"
	"time"
)

func main() {
	// 捕获异常
	defer func() {
		if err := recover(); err != nil {
			logs.PrintLogPanic(fmt.Errorf("%v", err))
			// 阻塞防止主线程退出中断异步打印日志
			select {}
		}
	}()

	// 注册模块
	logic.RegisterModule()
	// 注册路由
	api.RegisterRouter(logic.Module.Server())

	// 连接建立时
	logic.Module.Server().SetOnConnStart(func(conn iface.IConnection) {
		conn.SetProperty("Client", conn.RemoteAddr())
	})
	// 连接断开后
	logic.Module.Server().SetOnConnStop(func(conn iface.IConnection) {
		logs.PrintLogInfo(fmt.Sprintf("客户端%v下线", conn.GetProperty("Client")))
	})

	go func(s iface.IServer) {
		for range time.Tick(time.Second * 3) {
			logs.PrintLogInfo(fmt.Sprint("当前线程数：", runtime.NumGoroutine(), "\t当前连接数量：", s.GetConnMgr().Len()))
		}
	}(logic.Module.Server())

	// 开始监听服务
	logic.Module.Server().Listen()
}
