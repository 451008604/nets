package nets

import (
	"github.com/451008604/shard-map"
	"reflect"
	"sync"
)

type connectionEntry struct {
	id   string
	conn IConnection
}

type ConnectionManager struct {
	manager            *ServerManager
	connections        *shardmap.ShardMap[string, *connectionEntry]
	mu                 sync.Mutex
	active             uint
	closed             bool
	wg                 sync.WaitGroup
	hookMu             sync.RWMutex
	connOnOpened       func(IConnection)
	connOnClosed       func(IConnection)
	connOnRateLimiting func(IConnection)
}

func newConnectionManager(r *ServerManager) *ConnectionManager {
	return &ConnectionManager{manager: r, connections: shardmap.NewShardMap[string, *connectionEntry]()}
}
func GetInstanceConnManager() *ConnectionManager { return getDefaultServerManager().connections }
func (m *ConnectionManager) RangeConnections(handler func(IConnection)) {
	m.connections.Range(func(_ string, entry *connectionEntry) bool { handler(entry.conn); return true })
}

// TryAdd atomically admits a long connection under the same budget as HTTP requests.
func (m *ConnectionManager) TryAdd(conn IConnection) bool {
	if conn == nil {
		return false
	}
	if owner, ok := conn.(interface{ managerOwner() *ServerManager }); ok && owner.managerOwner() != m.manager {
		return false
	}
	entry := &connectionEntry{id: conn.GetConnId(), conn: conn}
	m.mu.Lock()
	if m.closed || m.manager.IsClose() || m.active >= m.manager.config.AppConf.MaxConn {
		m.mu.Unlock()
		conn.Close()
		return false
	}
	if _, exists := m.connections.Get(entry.id); exists {
		m.mu.Unlock()
		conn.Close()
		return false
	}
	m.active++
	m.connections.Set(entry.id, entry)
	m.wg.Add(1)
	m.mu.Unlock()
	go func() {
		defer m.wg.Done()
		defer m.removeEntry(entry)
		defer conn.Close()
		defer m.manager.messages.GetErrCapture(conn)
		conn.Open()
	}()
	return true
}
func (m *ConnectionManager) Add(conn IConnection) { m.TryAdd(conn) }
func (m *ConnectionManager) Remove(conn IConnection) {
	if conn == nil {
		return
	}
	conn.Close()
	m.mu.Lock()
	defer m.mu.Unlock()
	entry, exists := m.connections.Get(conn.GetConnId())
	// A custom connection may contain slices or maps. Never compare such interface
	// values; its own Open goroutine will release the captured registration token.
	if exists && reflect.ValueOf(conn).Comparable() && reflect.ValueOf(entry.conn).Comparable() && entry.conn == conn {
		m.connections.Delete(entry.id)
		m.active--
	}
}
func (m *ConnectionManager) removeEntry(entry *connectionEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if current, exists := m.connections.Get(entry.id); exists && current == entry {
		m.connections.Delete(entry.id)
		m.active--
	}
}
func (m *ConnectionManager) Get(id string) (IConnection, bool) {
	entry, exists := m.connections.Get(id)
	if !exists {
		return nil, false
	}
	return entry.conn, true
}
func (m *ConnectionManager) Len() int { return m.connections.Len() }
func (m *ConnectionManager) acquireRequest() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed || m.manager.IsClose() || m.active >= m.manager.config.AppConf.MaxConn {
		return false
	}
	m.active++
	m.wg.Add(1)
	return true
}
func (m *ConnectionManager) releaseRequest() {
	m.mu.Lock()
	m.active--
	m.mu.Unlock()
	m.wg.Done()
}
func (m *ConnectionManager) ClearConn() {
	m.mu.Lock()
	m.closed = true
	m.mu.Unlock()
	var entries []*connectionEntry
	m.connections.Range(func(_ string, entry *connectionEntry) bool { entries = append(entries, entry); return true })
	for _, entry := range entries {
		entry.conn.Close()
		m.removeEntry(entry)
	}
}
func (m *ConnectionManager) Wait() { m.wg.Wait() }
func (m *ConnectionManager) SetConnOpened(hook func(IConnection)) {
	m.hookMu.Lock()
	m.connOnOpened = hook
	m.hookMu.Unlock()
}
func (m *ConnectionManager) SetConnClosed(hook func(IConnection)) {
	m.hookMu.Lock()
	m.connOnClosed = hook
	m.hookMu.Unlock()
}
func (m *ConnectionManager) SetConnOnRateLimiting(hook func(IConnection)) {
	m.hookMu.Lock()
	m.connOnRateLimiting = hook
	m.hookMu.Unlock()
}
func (m *ConnectionManager) GetConnOpened(conn IConnection) {
	m.hookMu.RLock()
	hook := m.connOnOpened
	m.hookMu.RUnlock()
	if conn != nil && hook != nil {
		defer m.manager.messages.GetErrCapture(conn)
		hook(conn)
	}
}
func (m *ConnectionManager) GetConnClosed(conn IConnection) {
	m.hookMu.RLock()
	hook := m.connOnClosed
	m.hookMu.RUnlock()
	if conn != nil && hook != nil {
		defer m.manager.messages.GetErrCapture(conn)
		hook(conn)
	}
}
func (m *ConnectionManager) ConnRateLimiting(conn IConnection) {
	m.hookMu.RLock()
	hook := m.connOnRateLimiting
	m.hookMu.RUnlock()
	if conn != nil && hook != nil {
		defer m.manager.messages.GetErrCapture(conn)
		hook(conn)
	}
}
