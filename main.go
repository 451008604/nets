package main

import (
	"fmt"
	"github.com/451008604/nets/network"
	"runtime"
	"time"
)

func main() {
	go listenChannelStatus()

	// 开始监听服务
	serverTCP := network.NewServerTCP(nil)
	serverTCP.Listen()

	serverWS := network.NewServerWS(nil)
	serverWS.Listen()

	// 阻塞主进程
	network.ServerWaitFlag.Wait()
	network.GetInstanceConnManager().ClearConn()
}

func listenChannelStatus() {
	goroutineNum := 0
	for range time.Tick(time.Second * 1) {
		if temp := runtime.NumGoroutine(); temp != goroutineNum {
			goroutineNum = temp
			fmt.Printf("currentNumberOfThreads: %v\tcurrentNumberOfConnections: %v\n", goroutineNum, network.GetInstanceConnManager().Len())
		}
	}
}
