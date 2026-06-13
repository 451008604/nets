package nets

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
)

type ServerManager struct {
	servers   []IServer
	isClosed  int32 // Whether Service is Closed / 服务是否已关闭
	waitGroup sync.WaitGroup
}

var serverCtx, serverCtxCancel = context.WithCancel(context.Background())

var instanceServerManager *ServerManager
var instanceServerManagerOnce = sync.Once{}

// Server Manager / 服务管理器
func GetInstanceServerManager() *ServerManager {
	instanceServerManagerOnce.Do(func() {
		instanceServerManager = &ServerManager{
			servers: make([]IServer, 0),
		}
		go operatingSystemSignalHandler()
	})
	return instanceServerManager
}

func (c *ServerManager) RegisterServer(server ...IServer) {
	if len(server) == 0 {
		return
	}
	for _, iServer := range server {
		c.servers = append(c.servers, iServer)
		go iServer.Start()
	}

	// Block subsequent execution, wait for service shutdown / 阻塞后续执行，等待服务关闭
	<-serverCtx.Done()
	// Close All Connections / 关闭所有的连接
	GetInstanceConnManager().ClearConn()

	c.waitGroup.Wait()
	GetInstanceWorkerPool().Stop()
}

func (c *ServerManager) IsClose() bool {
	return atomic.LoadInt32(&c.isClosed) != 0
}

func (c *ServerManager) WaitGroupAdd(delta int) {
	c.waitGroup.Add(delta)
}

func (c *ServerManager) WaitGroupDone() {
	c.waitGroup.Done()
}

func (c *ServerManager) StopAll() {
	if len(c.servers) == 0 {
		return
	}
	if !atomic.CompareAndSwapInt32(&c.isClosed, 0, 1) {
		return
	}
	serverCtxCancel()
}

func operatingSystemSignalHandler() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer signal.Stop(signalCh)

	select {
	case sig := <-signalCh:
		fmt.Printf("Received signal: %v\n", sig)
	case <-serverCtx.Done():
	}

	// Execute Pre-exit Processing / 执行进程退出前的处理
	GetInstanceServerManager().StopAll()
}
