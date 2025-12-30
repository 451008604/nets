package nets

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type ServerManager struct {
	servers       []IServer
	isClosed      bool      // 服务是否已关闭
	blockMainChan chan bool // 服务启动后阻塞主协程
	waitGroup     sync.WaitGroup
}

var instanceServerManager *ServerManager
var instanceServerManagerOnce = sync.Once{}

// 服务管理器
func GetInstanceServerManager() *ServerManager {
	instanceServerManagerOnce.Do(func() {
		instanceServerManager = &ServerManager{
			servers:       make([]IServer, 0),
			isClosed:      false,
			blockMainChan: make(chan bool),
			waitGroup:     sync.WaitGroup{},
		}
		go operatingSystemSignalHandler()
	})
	return instanceServerManager
}

func (c *ServerManager) RegisterServer(server ...IServer) {
	for _, iServer := range server {
		c.servers = append(c.servers, iServer)
		go iServer.Start()
	}

	// 阻塞后续执行，等待服务关闭
	<-c.blockMainChan

	// 关闭所有的连接
	GetInstanceConnManager().ClearConn()

	c.waitGroup.Wait()
}

func (c *ServerManager) Servers() []IServer {
	return c.servers
}

func (c *ServerManager) IsClose() bool {
	return c.isClosed
}

func (c *ServerManager) WaitGroupAdd(delta int) {
	c.waitGroup.Add(delta)
}

func (c *ServerManager) WaitGroupDone() {
	c.waitGroup.Done()
}

func (c *ServerManager) StopAll() {
	if c.isClosed {
		return
	}
	c.isClosed = true
	c.blockMainChan <- c.isClosed
}

func operatingSystemSignalHandler() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGKILL)
	sig := <-signalCh
	fmt.Printf("Received signal: %v\n", sig)
	// 执行进程退出前的处理
	GetInstanceServerManager().StopAll()
}
