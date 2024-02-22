package main

import (
	"encoding/json"
	"fmt"
	"github.com/451008604/nets/config"
	"github.com/451008604/nets/network"
	"os"
	"runtime"
	"time"
)

func main() {
	go listenChannelStatus()

	readFile, err := os.ReadFile("conf.json")
	if err != nil {
		return
	}
	conf := config.AppConf{}
	_ = json.Unmarshal(readFile, &conf)
	config.SetServerConf(conf)

	// 开始监听服务
	serverTCP := network.NewServerTCP()
	serverTCP.Listen()

	serverWS := network.NewServerWS()
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
