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
	connId      int64                              // 用于客户端连接的自增Id
	connections sync.Map                           // 管理的连接信息
	signalCh    chan os.Signal                     // 处理系统信号
	closeConnId chan int                           // 已关闭的连接Id集合
	len         uint32                             // 连接数量
	onConnOpen  func(connection iface.IConnection) // 该Server连接创建时的Hook函数
	onConnClose func(connection iface.IConnection) // 该Server连接断开时的Hook函数
}

var instanceConnManager iface.IConnManager
var instanceConnManagerOnce = sync.Once{}

// 全局唯一连接管理器
func GetInstanceConnManager() iface.IConnManager {
	instanceConnManagerOnce.Do(func() {
		instanceConnManager = &connManager{
			connections: sync.Map{},
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
	c.connections.Store(conn.GetConnId(), conn)
	atomic.AddUint32(&c.len, 1)

	go conn.Start(conn.StartReader, conn.StartWriter)

	// 调用打开连接hook函数
	if c.onConnOpen != nil {
		c.onConnOpen(conn)
	}
}

func (c *connManager) Remove(conn iface.IConnection) {
	value, ok := c.connections.LoadAndDelete(conn.GetConnId())
	if !ok {
		return
	}
	if value != conn {
		c.connections.Store(conn.GetConnId(), value)
		return
	}
	atomic.AddUint32(&c.len, ^uint32(0))
	c.setClosingConn(conn.GetConnId())

	conn.Stop()

	// 调用关闭连接hook函数
	if c.onConnClose != nil {
		c.onConnClose(conn)
	}
}

func (c *connManager) Get(connId int) (iface.IConnection, bool) {
	value, ok := c.connections.Load(connId)
	return value.(iface.IConnection), ok
}

func (c *connManager) Len() int {
	return int(c.len)
}

func (c *connManager) ClearConn() {
	// 清理全部的connections信息
	c.connections.Range(func(key, value any) bool {
		c.Remove(value.(iface.IConnection))
		return true
	})
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
