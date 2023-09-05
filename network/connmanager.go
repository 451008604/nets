package network

import (
	"errors"
	"github.com/451008604/socketServerFrame/iface"
	"sync"
	"sync/atomic"
)

type ConnManager struct {
	connID      int64                              // 客户端连接自增ID
	connLock    sync.RWMutex                       // 连接的读写锁
	connections map[int]iface.IConnection          // 管理的连接信息
	onConnOpen  func(connection iface.IConnection) // 该Server连接创建时的Hook函数
	onConnClose func(connection iface.IConnection) // 该Server连接断开时的Hook函数
}

var instanceConnManager *ConnManager
var instanceConnManagerOnce = sync.Once{}

func GetInstanceConnManager() *ConnManager {
	instanceConnManagerOnce.Do(func() {
		instanceConnManager = &ConnManager{
			connections: map[int]iface.IConnection{},
			connLock:    sync.RWMutex{},
		}
	})
	return instanceConnManager
}

func (c *ConnManager) NewConnID() int64 {
	atomic.AddInt64(&c.connID, 1)
	return c.connID
}

// 添加连接
func (c *ConnManager) Add(conn iface.IConnection) {
	c.connLock.Lock()
	defer c.connLock.Unlock()

	c.connections[conn.GetConnID()] = conn

	// 调用打开连接hook函数
	if c.onConnOpen != nil {
		c.onConnOpen(conn)
	}
}

// 删除连接
func (c *ConnManager) Remove(conn iface.IConnection) {
	c.connLock.Lock()
	defer c.connLock.Unlock()

	delete(c.connections, conn.GetConnID())

	// 调用关闭连接hook函数
	if c.onConnClose != nil {
		c.onConnClose(conn)
	}
}

// 根据ConnID获取连接
func (c *ConnManager) Get(connID int) (iface.IConnection, error) {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	if conn, ok := c.connections[connID]; ok {
		return conn, nil
	} else {
		return nil, errors.New("未获取到绑定的conn")
	}
}

// 获取当前连接数量
func (c *ConnManager) Len() int {
	return len(c.connections)
}

// 删除并停止所有连接
func (c *ConnManager) ClearConn() {
	c.connLock.Lock()
	defer c.connLock.Unlock()

	// 清理全部的connections信息
	for connID, conn := range c.connections {
		conn.Stop()
		delete(c.connections, connID)
	}
}

// 连接创建时的Hook函数
func (c *ConnManager) OnConnOpen(fun func(conn iface.IConnection)) {
	c.onConnOpen = fun
}

// 连接断开时的Hook函数
func (c *ConnManager) OnConnClose(fun func(conn iface.IConnection)) {
	c.onConnClose = fun
}
