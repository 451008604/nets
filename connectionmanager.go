package nets

import (
	"github.com/451008604/shard-map"
	"sync"
)

type ConnectionManager struct {
	connections        *shardmap.ShardMap[string, IConnection] // Managed Connection Info / 管理的连接信息
	connOnOpened       func(conn IConnection)                  // Hook Function on Connection Open / 连接建立时的Hook函数
	connOnClosed       func(conn IConnection)                  // Hook Function on Connection Close / 连接断开时的Hook函数
	connOnRateLimiting func(conn IConnection)                  // Hook Function on Rate Limiting / 触发限流时的Hook函数
	callbackMutex      sync.RWMutex                            // Protects all callback fields
}

var instanceConnManager *ConnectionManager
var instanceConnManagerOnce = sync.Once{}

// Connection Manager / 连接管理器
func GetInstanceConnManager() *ConnectionManager {
	instanceConnManagerOnce.Do(func() {
		instanceConnManager = &ConnectionManager{
			connections: shardmap.NewShardMap[string, IConnection](),
		}
	})
	return instanceConnManager
}

func (c *ConnectionManager) RangeConnections(handler func(conn IConnection)) {
	c.connections.Range(func(key string, value IConnection) bool {
		handler(value)
		return true
	})
}

func (c *ConnectionManager) Add(conn IConnection) {
	c.connections.Set(conn.GetConnId(), conn)

	GetInstanceServerManager().WaitGroupAdd(1)
	go conn.Open()
}

// Register registers a connection without starting Open() goroutine.
// Used for short-lived connections like HTTP.
// Returns false if registering would exceed MaxConn; the caller is
// responsible for emitting an error response and skipping the handler.
// Register 注册连接但不启动 Open() 协程，用于 HTTP 等短连接。
// 如果注册会导致超过 MaxConn 则返回 false,调用方负责返回错误响应并跳过 handler。
func (c *ConnectionManager) Register(conn IConnection) bool {
	if conn == nil {
		return false
	}
	if defaultServer.AppConf.MaxConn > 0 && c.Len() >= defaultServer.AppConf.MaxConn {
		return false
	}
	c.connections.Set(conn.GetConnId(), conn)
	if defaultServer.AppConf.MaxConn > 0 && c.Len() > defaultServer.AppConf.MaxConn {
		c.connections.Delete(conn.GetConnId())
		return false
	}
	c.GetConnOpened(conn)
	return true
}

func (c *ConnectionManager) Remove(conn IConnection) {
	if conn == nil {
		return
	}
	conn.Close()

	c.connections.Delete(conn.GetConnId())
}

func (c *ConnectionManager) Get(connId string) (IConnection, bool) {
	return c.connections.Get(connId)
}

func (c *ConnectionManager) Len() int {
	return c.connections.Len()
}

func (c *ConnectionManager) ClearConn() {
	var connIds []string
	c.connections.Range(func(key string, value IConnection) bool {
		connIds = append(connIds, key)
		return true
	})
	for _, connId := range connIds {
		if conn, ok := c.connections.Get(connId); ok {
			c.Remove(conn)
		}
	}
}

func (c *ConnectionManager) SetConnOpened(connOpenCallBack func(conn IConnection)) {
	c.callbackMutex.Lock()
	defer c.callbackMutex.Unlock()
	c.connOnOpened = connOpenCallBack
}

func (c *ConnectionManager) GetConnOpened(conn IConnection) {
	if conn == nil {
		return
	}
	c.callbackMutex.RLock()
	callback := c.connOnOpened
	c.callbackMutex.RUnlock()
	if callback == nil {
		return
	}
	callback(conn)
}

func (c *ConnectionManager) SetConnClosed(connCloseCallBack func(conn IConnection)) {
	c.callbackMutex.Lock()
	defer c.callbackMutex.Unlock()
	c.connOnClosed = connCloseCallBack
}

func (c *ConnectionManager) GetConnClosed(conn IConnection) {
	if conn == nil {
		return
	}
	c.callbackMutex.RLock()
	callback := c.connOnClosed
	c.callbackMutex.RUnlock()
	if callback == nil {
		return
	}
	callback(conn)
}

func (c *ConnectionManager) SetConnOnRateLimiting(limitCallBack func(conn IConnection)) {
	c.callbackMutex.Lock()
	defer c.callbackMutex.Unlock()
	c.connOnRateLimiting = limitCallBack
}

func (c *ConnectionManager) ConnRateLimiting(conn IConnection) {
	if conn == nil {
		return
	}
	c.callbackMutex.RLock()
	callback := c.connOnRateLimiting
	c.callbackMutex.RUnlock()
	if callback == nil {
		return
	}
	callback(conn)
}
