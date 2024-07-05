package network

import (
	"fmt"
	"github.com/451008604/nets/iface"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type serverManager struct {
	servers []iface.IServer
}

var instanceServerManager iface.IServerManager
var instanceServerManagerOnce = sync.Once{}

// 服务管理器
func GetInstanceServerManager() iface.IServerManager {
	instanceServerManagerOnce.Do(func() {
		manager := &serverManager{
			servers: make([]iface.IServer, 0),
		}
		operatingSystemSignalHandler(manager)
		instanceServerManager = manager
	})

	return instanceServerManager
}

func (c *serverManager) RegisterServer(server ...iface.IServer) {
	for _, iServer := range server {
		c.servers = append(c.servers, iServer)
		go iServer.Start()
	}
}

func operatingSystemSignalHandler(c *serverManager) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGKILL)
	go func() {
		sig := <-signalCh
		fmt.Printf("Received signal: %v\n", sig)

		GetServerWS().Stop()
		GetServerTCP().Stop()
	}()
}
