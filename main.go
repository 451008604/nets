package main

import (
	"fmt"
	_ "github.com/451008604/socketServerFrame/api"
	"github.com/451008604/socketServerFrame/common"
	"github.com/451008604/socketServerFrame/config"
	"github.com/451008604/socketServerFrame/logic"
	"github.com/451008604/socketServerFrame/logs"
	"github.com/451008604/socketServerFrame/network"
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
	config.SetRemoteConfigAddress("http://101.43.0.205:6001")
	config.InitServerConfig()

	// 注册模块
	common.Module.SetServerTCP(network.NewServerTCP())
	common.Module.SetServerWS(network.NewServerWS())
	common.Module.SetNotify(network.NewNotifyManager())
	// 注册hook函数
	network.GetInstanceConnManager().OnConnOpen(logic.OnConnectionOpen)
	network.GetInstanceConnManager().OnConnClose(logic.OnConnectionClose)

	// 开始监听服务
	runServer()

	// 阻塞主进程
	network.ServerWaitFlag.Wait()
}

func runServer() {
	common.Module.ServerTCP().Listen()
	common.Module.ServerWS().Listen()
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
