package main

import (
	"encoding/json"
	"fmt"
	"github.com/451008604/nets/config"
	"github.com/451008604/nets/iface"
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

	// 注册hook函数
	network.GetInstanceConnManager().OnConnOpen(func(conn iface.IConnection) {

	})
	network.GetInstanceConnManager().OnConnClose(func(conn iface.IConnection) {

	})

	// 开始监听服务
	serverTCP := network.NewServerTCP()
	serverTCP.Listen()

	serverWS := network.NewServerWS()
	serverWS.Listen()

	// 阻塞主进程
	network.ServerWaitFlag.Wait()
}

func listenChannelStatus() {
	goroutineNum := 0
	for range time.Tick(time.Second * 1) {
		if temp := runtime.NumGoroutine(); temp != goroutineNum {
			goroutineNum = temp
			fmt.Printf("当前线程数：%v\t当前连接数量：%v\n", goroutineNum, network.GetInstanceConnManager().Len())
		}
	}
}
