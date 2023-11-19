package network

import (
	"errors"
	"github.com/451008604/nets/config"
	"github.com/451008604/nets/iface"
	"sync"
	"sync/atomic"
)

type ConnManager struct {
	connID      int64                              // 用于客户端连接的自增ID
	connections sync.Map                           // 管理的连接信息
	closeConnID chan int                           // 已关闭的连接ID集合
	len         uint32                             // 连接数量
	onConnOpen  func(connection iface.IConnection) // 该Server连接创建时的Hook函数
	onConnClose func(connection iface.IConnection) // 该Server连接断开时的Hook函数
}

var instanceConnManager iface.IConnManager
var instanceConnManagerOnce = sync.Once{}

func GetInstanceConnManager() iface.IConnManager {
	instanceConnManagerOnce.Do(func() {
		instanceConnManager = &ConnManager{
			connections: sync.Map{},
			closeConnID: make(chan int, config.GetServerConf().MaxConn),
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
	c.connections.Store(conn.GetConnID(), conn)
	atomic.AddUint32(&c.len, 1)

	go conn.Start(conn.StartReader, conn.StartWriter)

	// 调用打开连接hook函数
	if c.onConnOpen != nil {
		c.onConnOpen(conn)
	}
}

// 删除连接
func (c *ConnManager) Remove(conn iface.IConnection) {
	if value, ok := c.connections.Load(conn.GetConnID()); !ok || value != conn {
		return
	}
	c.connections.Delete(conn.GetConnID())
	atomic.AddUint32(&c.len, ^uint32(0))

	// 调用关闭连接hook函数
	if c.onConnClose != nil {
		c.onConnClose(conn)
	}

	conn.Stop()

	c.setClosingConn(conn.GetConnID())
}

// 根据ConnID获取连接
func (c *ConnManager) Get(connID int) (iface.IConnection, error) {
	if value, ok := c.connections.Load(connID); ok {
		return value.(iface.IConnection), nil
	} else {
		return nil, errors.New("未获取到绑定的conn")
	}
}

// 获取当前连接数量
func (c *ConnManager) Len() int {
	return int(c.len)
}

// 删除并停止所有连接
func (c *ConnManager) ClearConn() {
	// 清理全部的connections信息
	c.connections.Range(func(key, value any) bool {
		c.Remove(value.(iface.IConnection))
		return true
	})
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
	// 存入回收列表
	c.closeConnID <- connID
}

func (c *ConnManager) getClosingConn() int {
	// 回收列表中存在则取出使用
	select {
	case connID := <-c.closeConnID:
		return connID
	default:
		return 0
	}
}
