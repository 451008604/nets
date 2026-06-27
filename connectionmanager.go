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
	go conn.Open()
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
	c.connOnOpened = connOpenCallBack
}

func (c *ConnectionManager) GetConnOpened(conn IConnection) {
	if conn == nil || c.connOnOpened == nil {
		return
	}
	defer GetInstanceMsgHandler().GetErrCapture(conn)
	c.connOnOpened(conn)
}

func (c *ConnectionManager) SetConnClosed(connCloseCallBack func(conn IConnection)) {
	c.connOnClosed = connCloseCallBack
}

func (c *ConnectionManager) GetConnClosed(conn IConnection) {
	if conn == nil || c.connOnClosed == nil {
		return
	}
	defer GetInstanceMsgHandler().GetErrCapture(conn)
	c.connOnClosed(conn)
}

func (c *ConnectionManager) SetConnOnRateLimiting(limitCallBack func(conn IConnection)) {
	c.connOnRateLimiting = limitCallBack
}

func (c *ConnectionManager) ConnRateLimiting(conn IConnection) {
	if conn == nil || c.connOnRateLimiting == nil {
		return
	}
	defer GetInstanceMsgHandler().GetErrCapture(conn)
	c.connOnRateLimiting(conn)
}
