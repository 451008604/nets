package network

import (
	"container/list"
	"errors"
	"github.com/451008604/socketServerFrame/iface"
	"sync"
	"sync/atomic"
)

type ConnManager struct {
	connID          int64                              // 用于客户端连接的自增ID
	connections     map[int]iface.IConnection          // 管理的连接信息
	connectionsLock sync.RWMutex                       // 连接的读写锁
	closeConnID     list.List                          // 已关闭的连接ID集合
	closeConnIDLock sync.Mutex                         // 存储关闭连接的读写锁
	onConnOpen      func(connection iface.IConnection) // 该Server连接创建时的Hook函数
	onConnClose     func(connection iface.IConnection) // 该Server连接断开时的Hook函数
}

var instanceConnManager iface.IConnManager
var instanceConnManagerOnce = sync.Once{}

func GetInstanceConnManager() iface.IConnManager {
	instanceConnManagerOnce.Do(func() {
		instanceConnManager = &ConnManager{
			connections:     map[int]iface.IConnection{},
			connectionsLock: sync.RWMutex{},
			closeConnID:     list.List{},
			closeConnIDLock: sync.Mutex{},
		}
	})
	return instanceConnManager
}

func (c *ConnManager) NewConnID() int {
	if connID := c.getClosingConn(); connID != 0 {
		return connID
	}
	// 回收列表为空时递增ID
	atomic.AddInt64(&c.connID, 1)
	return int(c.connID)
}

// 添加连接
func (c *ConnManager) Add(conn iface.IConnection) {
	c.connectionsLock.Lock()
	defer c.connectionsLock.Unlock()

	c.connections[conn.GetConnID()] = conn

	// 调用打开连接hook函数
	if c.onConnOpen != nil {
		c.onConnOpen(conn)
	}
}

// 删除连接
func (c *ConnManager) Remove(conn iface.IConnection) {
	c.setClosingConn(conn.GetConnID())

	c.connectionsLock.Lock()
	defer c.connectionsLock.Unlock()

	delete(c.connections, conn.GetConnID())

	// 调用关闭连接hook函数
	if c.onConnClose != nil {
		c.onConnClose(conn)
	}
}

// 根据ConnID获取连接
func (c *ConnManager) Get(connID int) (iface.IConnection, error) {
	c.connectionsLock.Lock()
	defer c.connectionsLock.Unlock()

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
	c.connectionsLock.Lock()
	defer c.connectionsLock.Unlock()

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

func (c *ConnManager) setClosingConn(connID int) {
	c.closeConnIDLock.Lock()
	defer c.closeConnIDLock.Unlock()

	// 存入回收列表
	c.closeConnID.PushBack(connID)
}

func (c *ConnManager) getClosingConn() int {
	c.closeConnIDLock.Lock()
	defer c.closeConnIDLock.Unlock()

	// 回收列表中存在则取出使用
	if c.closeConnID.Len() > 0 {
		connID := c.closeConnID.Remove(c.closeConnID.Front())
		return connID.(int)
	}
	return 0
}
