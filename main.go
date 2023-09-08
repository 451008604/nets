package main

import (
	"fmt"
	"github.com/451008604/socketServerFrame/database/redis"
	"github.com/451008604/socketServerFrame/database/sql"
	"github.com/451008604/socketServerFrame/logic"
	"github.com/451008604/socketServerFrame/logs"
	"github.com/451008604/socketServerFrame/modules"
	"github.com/451008604/socketServerFrame/network"
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
	go listenChannelStatus()

	// 注册模块
	modules.Module.SetServerTCP(network.NewServerTCP())
	modules.Module.SetServerWS(network.NewServerWS())
	modules.Module.SetNotify(network.NewNotifyManager())
	modules.Module.SetSql(sql.NewSqlDBModel())
	modules.Module.SetRedis(redis.NewRedisModel())
	// 注册hook函数
	network.GetInstanceConnManager().OnConnOpen(logic.OnConnectionOpen)
	network.GetInstanceConnManager().OnConnClose(logic.OnConnectionClose)

	// 开始监听服务
	runServer()

	// 阻塞主进程
	network.ServerWaitFlag.Wait()
}

func runServer() {
	modules.Module.ServerTCP().Listen()
	modules.Module.ServerWS().Listen()
}

func listenChannelStatus() {
	goroutineNum := 0
	for range time.Tick(time.Second * 1) {
		if temp := runtime.NumGoroutine(); temp != goroutineNum {
			goroutineNum = temp
			logs.PrintLogInfo(fmt.Sprintf("当前线程数：%v\t当前连接数量：%v", goroutineNum, network.GetInstanceConnManager().Len()))
		}
	}
}
