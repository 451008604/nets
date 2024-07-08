package network

import (
	"github.com/451008604/nets/iface"
	"sync"
	"sync/atomic"
)

type connectionManager struct {
	connId      int64                                     // 用于客户端连接的自增Id
	connections ConcurrentMap[Integer, iface.IConnection] // 管理的连接信息
	closeConnId chan int                                  // 已关闭的连接Id集合
	onConnOpen  func(connection iface.IConnection)        // 连接建立时的Hook函数
	onConnClose func(connection iface.IConnection)        // 连接断开时的Hook函数
	removeList  chan iface.IConnection                    // 等待关闭的连接列表
}

var instanceConnManager iface.IConnectionManager
var instanceConnManagerOnce = sync.Once{}

// 连接管理器
func GetInstanceConnManager() iface.IConnectionManager {
	instanceConnManagerOnce.Do(func() {
		manager := &connectionManager{
			connections: NewConcurrentStringer[Integer, iface.IConnection](),
			closeConnId: make(chan int, defaultServer.AppConf.MaxConn),
			removeList:  make(chan iface.IConnection, defaultServer.AppConf.MaxConn),
		}
		go onConnRemoveList(manager)
		instanceConnManager = manager
	})
	return instanceConnManager
}

func (c *connectionManager) NewConnId() int {
	if connId := c.getClosingConn(); connId != 0 {
		return connId
	}
	// 回收列表为空时递增Id
	atomic.AddInt64(&c.connId, 1)
	return int(c.connId)
}

func (c *connectionManager) Add(conn iface.IConnection) {
	c.connections.Set(Integer(conn.GetConnId()), conn)

	conn.SetProperty(SysPropertyConnOpened, c.onConnOpen)
	conn.SetProperty(SysPropertyConnClosed, c.onConnClose)
	go conn.Start(conn.StartReader, conn.StartWriter)
}

func (c *connectionManager) Remove(conn iface.IConnection) {
	c.removeList <- conn
}

func (c *connectionManager) Get(connId int) (iface.IConnection, bool) {
	value, ok := c.connections.Get(Integer(connId))
	return value, ok
}

func (c *connectionManager) Len() int {
	return c.connections.Count()
}

func (c *connectionManager) ClearConn() {
	// 清理全部的connections信息
	for _, v := range c.connections.Items() {
		c.Remove(v)
	}
}

func (c *connectionManager) OnConnOpen(fun func(conn iface.IConnection)) {
	c.onConnOpen = fun
}

func (c *connectionManager) OnConnClose(fun func(conn iface.IConnection)) {
	c.onConnClose = fun
}

func (c *connectionManager) setClosingConn(connId int) {
	// 存入回收列表
	c.closeConnId <- connId
}

func (c *connectionManager) getClosingConn() int {
	// 回收列表中存在则取出使用
	select {
	case connId := <-c.closeConnId:
		return connId
	default:
		return 0
	}
}

func onConnRemoveList(c *connectionManager) {
	for conn := range c.removeList {
		if conn.IsClose() {
			continue
		}
		// 关闭连接
		conn.Stop()
		// 删除连接
		c.connections.Remove(Integer(conn.GetConnId()))
		// 回收连接Id
		c.setClosingConn(conn.GetConnId())
	}
}
