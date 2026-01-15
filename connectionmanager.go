package nets

import (
	"sync"
)

type ConnectionManager struct {
	connections        ConcurrentMap[string, IConnection] // 管理的连接信息
	connOnOpened       func(conn IConnection)             // 连接建立时的Hook函数
	connOnClosed       func(conn IConnection)             // 连接断开时的Hook函数
	connOnRateLimiting func(conn IConnection)             // 触发限流时的Hook函数
}

var instanceConnManager *ConnectionManager
var instanceConnManagerOnce = sync.Once{}

// 连接管理器
func GetInstanceConnManager() *ConnectionManager {
	instanceConnManagerOnce.Do(func() {
		manager := &ConnectionManager{
			connections: NewConcurrentMap[IConnection](),
		}
		instanceConnManager = manager
	})
	return instanceConnManager
}

func (c *ConnectionManager) RangeConnections(handler func(conn IConnection)) {
	for _, v := range c.connections.Items() {
		handler(v)
	}
}

func (c *ConnectionManager) Add(conn IConnection) {
	c.connections.Set(conn.GetConnId(), conn)

	go conn.Start()
}

func (c *ConnectionManager) Remove(conn IConnection) {
	conn.Stop()
	c.connections.Remove(conn.GetConnId())
}

func (c *ConnectionManager) Get(connId string) (IConnection, bool) {
	return c.connections.Get(connId)
}

func (c *ConnectionManager) Len() int {
	return c.connections.Count()
}

func (c *ConnectionManager) ClearConn() {
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
