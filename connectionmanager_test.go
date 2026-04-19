package nets

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
)

// 模拟连接用于测试
type mockConnection struct {
	id       string
	closed   int32
	deadTime int64
}

func (m *mockConnection) GetConnId() string  { return m.id }
func (m *mockConnection) IsClose() bool      { return atomic.LoadInt32(&m.closed) != 0 }
func (m *mockConnection) GetDeadTime() int64 { return m.deadTime }
func (m *mockConnection) Open()              {}
func (m *mockConnection) Close() bool {
	if atomic.AddInt32(&m.closed, 1) == 1 {
		return true
	}
	return false
}
func (m *mockConnection) DoTask(task func())                                     {}
func (m *mockConnection) StartReader() bool                                      { return true }
func (m *mockConnection) StartWriter(data []byte) bool                           { return true }
func (m *mockConnection) SetProperty(key string, value any)                      {}
func (m *mockConnection) GetProperty(key string) any                             { return nil }
func (m *mockConnection) RemoveProperty(key string)                              {}
func (m *mockConnection) SendMsg(msgId int32, msgData proto.Message)             {}
func (m *mockConnection) FlowControl() bool                                      { return false }
func (m *mockConnection) ProtocolToByte(msg proto.Message) []byte                { return nil }
func (m *mockConnection) ByteToProtocol(data []byte, target proto.Message) error { return nil }
func (m *mockConnection) RemoteAddrStr() string                                  { return "" }

func TestConnectionManager_AddAndGet(t *testing.T) {
	mgr := GetInstanceConnManager()
	mockConn := &mockConnection{id: "test-conn-1", deadTime: time.Now().Unix()}

	mgr.Add(mockConn)

	retrieved, exists := mgr.Get("test-conn-1")
	if !exists {
		t.Error("Expected connection to exist after Add")
	}
	if retrieved.GetConnId() != "test-conn-1" {
		t.Errorf("Expected connId test-conn-1, got %s", retrieved.GetConnId())
	}
}

func TestConnectionManager_Remove(t *testing.T) {
	mgr := GetInstanceConnManager()
	mockConn := &mockConnection{id: "test-conn-2", deadTime: time.Now().Unix()}

	mgr.Add(mockConn)

	_, exists := mgr.Get("test-conn-2")
	if !exists {
		t.Error("Expected connection to exist before Remove")
	}

	mgr.Remove(mockConn)

	_, exists = mgr.Get("test-conn-2")
	if exists {
		t.Error("Expected connection to be removed")
	}
}

func TestConnectionManager_Len(t *testing.T) {
	mgr := GetInstanceConnManager()

	mockConn1 := &mockConnection{id: "len-test-1", deadTime: time.Now().Unix()}
	mockConn2 := &mockConnection{id: "len-test-2", deadTime: time.Now().Unix()}

	mgr.Add(mockConn1)
	mgr.Add(mockConn2)

	if mgr.Len() < 2 {
		t.Errorf("Expected len >= 2, got %d", mgr.Len())
	}
}

func TestConnectionManager_RangeConnections(t *testing.T) {
	mgr := GetInstanceConnManager()
	count := 0

	mockConn1 := &mockConnection{id: "range-test-1", deadTime: time.Now().Unix()}
	mockConn2 := &mockConnection{id: "range-test-2", deadTime: time.Now().Unix()}

	mgr.Add(mockConn1)
	mgr.Add(mockConn2)

	mgr.RangeConnections(func(conn IConnection) {
		count++
		if conn.GetConnId() == "" {
			t.Error("Expected non-empty connId in RangeConnections")
		}
	})

	if count == 0 {
		t.Error("Expected RangeConnections to be called at least once")
	}
}

func TestConnectionManager_ClearConn(t *testing.T) {
	mgr := GetInstanceConnManager()

	mockConn1 := &mockConnection{id: "clear-test-1", deadTime: time.Now().Unix()}
	mockConn2 := &mockConnection{id: "clear-test-2", deadTime: time.Now().Unix()}

	mgr.Add(mockConn1)
	mgr.Add(mockConn2)

	mgr.ClearConn()

	_, exists1 := mgr.Get("clear-test-1")
	_, exists2 := mgr.Get("clear-test-2")

	if exists1 || exists2 {
		t.Error("Expected all connections to be cleared")
	}
}

func TestConnectionManager_SetAndGetConnOpened(t *testing.T) {
	mgr := GetInstanceConnManager()
	called := false

	mgr.SetConnOpened(func(conn IConnection) {
		called = true
		if conn.GetConnId() != "opened-test" {
			t.Errorf("Expected connId opened-test, got %s", conn.GetConnId())
		}
	})

	mockConn := &mockConnection{id: "opened-test", deadTime: time.Now().Unix()}
	mgr.GetConnOpened(mockConn)

	if !called {
		t.Error("Expected GetConnOpened to call the callback")
	}
}

func TestConnectionManager_GetConnOpened_NoCallback(t *testing.T) {
	instanceConnManager = nil; instanceConnManagerOnce = sync.Once{}
	mgr := GetInstanceConnManager()

	// 不设置回调
	mgr.GetConnOpened(&mockConnection{id: "no-callback", deadTime: time.Now().Unix()})
	// 不应 panic
}

func TestConnectionManager_SetAndGetConnClosed(t *testing.T) {
	mgr := GetInstanceConnManager()
	called := false

	mgr.SetConnClosed(func(conn IConnection) {
		called = true
		if conn.GetConnId() != "closed-test" {
			t.Errorf("Expected connId closed-test, got %s", conn.GetConnId())
		}
	})

	mockConn := &mockConnection{id: "closed-test", deadTime: time.Now().Unix()}
	mgr.GetConnClosed(mockConn)

	if !called {
		t.Error("Expected GetConnClosed to call the callback")
	}
}

func TestConnectionManager_GetConnClosed_NoCallback(t *testing.T) {
	instanceConnManager = nil; instanceConnManagerOnce = sync.Once{}
	mgr := GetInstanceConnManager()

	// 不设置回调
	mgr.GetConnClosed(&mockConnection{id: "no-callback", deadTime: time.Now().Unix()})
	// 不应 panic
}

func TestConnectionManager_SetAndGetConnOnRateLimiting(t *testing.T) {
	mgr := GetInstanceConnManager()
	called := false

	mgr.SetConnOnRateLimiting(func(conn IConnection) {
		called = true
		if conn.GetConnId() != "rate-test" {
			t.Errorf("Expected connId rate-test, got %s", conn.GetConnId())
		}
	})

	mockConn := &mockConnection{id: "rate-test", deadTime: time.Now().Unix()}
	mgr.ConnRateLimiting(mockConn)

	if !called {
		t.Error("Expected ConnRateLimiting to call the callback")
	}
}

func TestConnectionManager_ConnRateLimiting_NoCallback(t *testing.T) {
	instanceConnManager = nil; instanceConnManagerOnce = sync.Once{}
	mgr := GetInstanceConnManager()

	// 不设置回调
	mgr.ConnRateLimiting(&mockConnection{id: "no-callback", deadTime: time.Now().Unix()})
	// 不应 panic
}
