package network

import (
	"github.com/451008604/nets/iface"
	"sync"
	"sync/atomic"
)

type connectionManager struct {
	connId             int64                                     // 用于客户端连接的自增Id
	connections        ConcurrentMap[Integer, iface.IConnection] // 管理的连接信息
	closeConnId        chan int                                  // 已关闭的连接Id集合
	removeList         chan iface.IConnection                    // 等待关闭的连接列表
	connOnOpened       func(conn iface.IConnection)              // 连接建立时的Hook函数
	connOnClosed       func(conn iface.IConnection)              // 连接断开时的Hook函数
	connOnRateLimiting func(conn iface.IConnection)              // 触发限流时的Hook函数
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

func (c *connectionManager) RangeConnections(handler func(conn iface.IConnection)) {
	for _, v := range c.connections.Items() {
		handler(v)
	}
}

func (c *connectionManager) Add(conn iface.IConnection) {
	c.connections.Set(Integer(conn.GetConnId()), conn)

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
	c.RangeConnections(c.Remove)
}

func (c *connectionManager) SetConnOnOpened(connOpenCallBack func(conn iface.IConnection)) {
	c.connOnOpened = connOpenCallBack
}

func (c *connectionManager) ConnOnOpened(conn iface.IConnection) {
	if c.connOnOpened == nil {
		return
	}
	c.connOnOpened(conn)
}

func (c *connectionManager) SetConnOnClosed(connCloseCallBack func(conn iface.IConnection)) {
	c.connOnClosed = connCloseCallBack
}

func (c *connectionManager) ConnOnClosed(conn iface.IConnection) {
	if c.connOnClosed == nil {
		return
	}
	c.connOnClosed(conn)
}

func (c *connectionManager) SetConnOnRateLimiting(limitCallBack func(conn iface.IConnection)) {
	c.connOnRateLimiting = limitCallBack
}

func (c *connectionManager) ConnRateLimiting(conn iface.IConnection) {
	if c.connOnRateLimiting == nil {
		return
	}
	c.connOnRateLimiting(conn)
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

		if !c.connections.Has(Integer(conn.GetConnId())) {
			continue
		}

		// 删除连接
		c.connections.Remove(Integer(conn.GetConnId()))
		// 回收连接Id
		c.setClosingConn(conn.GetConnId())
	}
}
