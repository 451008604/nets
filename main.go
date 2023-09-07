package main

import (
	"fmt"
	"github.com/451008604/socketServerFrame/api"
	"github.com/451008604/socketServerFrame/logic"
	"github.com/451008604/socketServerFrame/logs"
	"github.com/451008604/socketServerFrame/network"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
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

	// 注册路由
	api.RegisterRouter(network.GetInstanceMsgHandler())
	// 注册模块
	logic.RegisterModule()

	// 开始监听服务
	runServer()

	network.ServerWaitFlag.Wait()
}

func runServer() {
	logic.Module.ServerTCP().Listen()
	logic.Module.ServerWS().Listen()
}
