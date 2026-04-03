package nets

import (
	"github.com/451008604/shard-map"
	"sync"
	"time"
)

type ConnectionManager struct {
	connections        *shardmap.ShardMap[string, IConnection] // 管理的连接信息
	connOnOpened       func(conn IConnection)                  // 连接建立时的Hook函数
	connOnClosed       func(conn IConnection)                  // 连接断开时的Hook函数
	connOnRateLimiting func(conn IConnection)                  // 触发限流时的Hook函数
}

var instanceConnManager *ConnectionManager
var instanceConnManagerOnce = sync.Once{}

// 连接管理器
func GetInstanceConnManager() *ConnectionManager {
	instanceConnManagerOnce.Do(func() {
		instanceConnManager = &ConnectionManager{
			connections: shardmap.NewShardMap[string, IConnection](),
		}
		go instanceConnManager.connRWTimeOut()
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

// 读写超时检测
func (c *ConnectionManager) connRWTimeOut() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for t := range ticker.C {
		c.RangeConnections(func(conn IConnection) {
			if t.Unix()-conn.GetDeadTime() > int64(defaultServer.AppConf.ConnRWTimeOut) {
				conn.Close()
			}
		})
	}
}
