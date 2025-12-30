package main

import (
	"sync"
	"sync/atomic"
)

type ConnectionManager struct {
	connId             int64                               // 用于客户端连接的自增Id
	connections        ConcurrentMap[Integer, IConnection] // 管理的连接信息
	closeConnId        chan int                            // 已关闭的连接Id集合
	removeList         chan IConnection                    // 等待关闭的连接列表
	connOnOpened       func(conn IConnection)              // 连接建立时的Hook函数
	connOnClosed       func(conn IConnection)              // 连接断开时的Hook函数
	connOnRateLimiting func(conn IConnection)              // 触发限流时的Hook函数
}

var instanceConnManager *ConnectionManager
var instanceConnManagerOnce = sync.Once{}

// 连接管理器
func GetInstanceConnManager() *ConnectionManager {
	instanceConnManagerOnce.Do(func() {
		manager := &ConnectionManager{
			connections: NewConcurrentStringer[Integer, IConnection](),
			closeConnId: make(chan int, defaultServer.AppConf.MaxConn),
			removeList:  make(chan IConnection, defaultServer.AppConf.MaxConn),
		}
		instanceConnManager = manager

		go func(c *ConnectionManager) {
			for conn := range c.removeList {
				if conn.GetConnId() == 0 || conn.IsClose() {
					continue
				}
				// 关闭连接
				conn.Stop()

				if c.connections.Has(Integer(conn.GetConnId())) {
					// 删除连接
					c.connections.Remove(Integer(conn.GetConnId()))
					// 回收连接Id
					c.setClosingConn(conn.GetConnId())
				}
			}
		}(manager)
	})
	return instanceConnManager
}

func (c *ConnectionManager) NewConnId() int {
	select {
	case connId := <-c.closeConnId:
		// 回收列表中存在则取出使用
		return connId
	default:
		// 回收列表为空时递增Id
		atomic.AddInt64(&c.connId, 1)
		return int(c.connId)
	}
}

func (c *ConnectionManager) RangeConnections(handler func(conn IConnection)) {
	for _, v := range c.connections.Items() {
		handler(v)
	}
}

func (c *ConnectionManager) Add(conn IConnection) {
	c.connections.Set(Integer(conn.GetConnId()), conn)

	go conn.Start(conn.StartReader, conn.StartWriter)
}

func (c *ConnectionManager) Remove(conn IConnection) {
	c.removeList <- conn
}

func (c *ConnectionManager) Get(connId int) (IConnection, bool) {
	value, ok := c.connections.Get(Integer(connId))
	return value, ok
}

func (c *ConnectionManager) Len() int {
	return c.connections.Count()
}

func (c *ConnectionManager) ClearConn() {
	// 清理全部的connections信息
	c.RangeConnections(c.Remove)
}

func (c *ConnectionManager) SetConnOnOpened(connOpenCallBack func(conn IConnection)) {
	c.connOnOpened = connOpenCallBack
}

func (c *ConnectionManager) ConnOnOpened(conn IConnection) {
	if c.connOnOpened == nil {
		return
	}
	c.connOnOpened(conn)
}

func (c *ConnectionManager) SetConnOnClosed(connCloseCallBack func(conn IConnection)) {
	c.connOnClosed = connCloseCallBack
}

func (c *ConnectionManager) ConnOnClosed(conn IConnection) {
	if c.connOnClosed == nil {
		return
	}
	c.connOnClosed(conn)
}

func (c *ConnectionManager) SetConnOnRateLimiting(limitCallBack func(conn IConnection)) {
	c.connOnRateLimiting = limitCallBack
}

func (c *ConnectionManager) ConnRateLimiting(conn IConnection) {
	if c.connOnRateLimiting == nil {
		return
	}
	c.connOnRateLimiting(conn)
}

func (c *ConnectionManager) setClosingConn(connId int) {
	// 存入回收列表
	c.closeConnId <- connId
}
