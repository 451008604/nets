package network

import (
	"fmt"
	"github.com/451008604/nets/iface"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
)

type connManager struct {
	connId      int64                                     // 用于客户端连接的自增Id
	connections ConcurrentMap[Integer, iface.IConnection] // 管理的连接信息
	signalCh    chan os.Signal                            // 处理系统信号
	closeConnId chan int                                  // 已关闭的连接Id集合
	onConnOpen  func(connection iface.IConnection)        // 该Server连接创建时的Hook函数
	onConnClose func(connection iface.IConnection)        // 该Server连接断开时的Hook函数
}

var instanceConnManager iface.IConnManager
var instanceConnManagerOnce = sync.Once{}

// 全局唯一连接管理器
func GetInstanceConnManager() iface.IConnManager {
	instanceConnManagerOnce.Do(func() {
		instanceConnManager = &connManager{
			connections: NewConcurrentStringer[Integer, iface.IConnection](),
			closeConnId: make(chan int, defaultServer.AppConf.MaxConn),
		}
		instanceConnManager.OperatingSystemSignalHandler()
	})
	return instanceConnManager
}

func (c *connManager) NewConnId() int {
	if connId := c.getClosingConn(); connId != 0 {
		return connId
	}
	// 回收列表为空时递增Id
	atomic.AddInt64(&c.connId, 1)
	return int(c.connId)
}

func (c *connManager) Add(conn iface.IConnection) {
	c.connections.Set(Integer(conn.GetConnId()), conn)

	go conn.Start(conn.StartReader, conn.StartWriter)

	// 调用打开连接hook函数
	if c.onConnOpen != nil {
		c.onConnOpen(conn)
	}
}

func (c *connManager) Remove(conn iface.IConnection) {
	value, ok := c.connections.Get(Integer(conn.GetConnId()))
	// 如果不存在，或者指针不同
	if !ok || value != conn {
		return
	}
	// 删除连接
	c.connections.Remove(Integer(conn.GetConnId()))
	// 回收连接Id
	c.setClosingConn(conn.GetConnId())
	// 关闭连接
	conn.Stop()

	// 调用关闭连接hook函数
	if c.onConnClose != nil {
		c.onConnClose(conn)
	}
}

func (c *connManager) Get(connId int) (iface.IConnection, bool) {
	value, ok := c.connections.Get(Integer(connId))
	return value, ok
}

func (c *connManager) Len() int {
	return c.connections.Count()
}

func (c *connManager) ClearConn() {
	// 清理全部的connections信息
	for _, v := range c.connections.Items() {
		c.Remove(v)
	}
}

func (c *connManager) OnConnOpen(fun func(conn iface.IConnection)) {
	c.onConnOpen = fun
}

func (c *connManager) OnConnClose(fun func(conn iface.IConnection)) {
	c.onConnClose = fun
}

func (c *connManager) setClosingConn(connId int) {
	// 存入回收列表
	c.closeConnId <- connId
}

func (c *connManager) getClosingConn() int {
	// 回收列表中存在则取出使用
	select {
	case connId := <-c.closeConnId:
		return connId
	default:
		return 0
	}
}

func (c *connManager) OperatingSystemSignalHandler() {
	if c.signalCh != nil {
		return
	}
	c.signalCh = make(chan os.Signal, 1)
	signal.Notify(c.signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGKILL)
	go func() {
		sig := <-c.signalCh
		fmt.Printf("Received signal: %v\n", sig)
		c.ClearConn()
		os.Exit(0)
	}()
}
