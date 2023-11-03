package main

import (
	"fmt"
	_ "github.com/451008604/socketServerFrame/api"
	"github.com/451008604/socketServerFrame/common"
	"github.com/451008604/socketServerFrame/logic"
	"github.com/451008604/socketServerFrame/logs"
	"github.com/451008604/socketServerFrame/network"
	"runtime"
	"time"
)

func main() {
	go listenChannelStatus()

	// 注册hook函数
	network.GetInstanceConnManager().OnConnOpen(logic.OnConnectionOpen)
	network.GetInstanceConnManager().OnConnClose(logic.OnConnectionClose)

	// 开始监听服务
	common.GetServerTCP().Listen()
	common.GetServerWS().Listen()

	// 阻塞主进程
	network.ServerWaitFlag.Wait()
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
