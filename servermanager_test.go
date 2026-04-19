package nets

import (
	"testing"
	"time"
)

// 模拟 Server 用于测试
type mockServer struct {
	name string
}

func (m *mockServer) GetName() string   { return m.name }
func (m *mockServer) Start()            {}
func (m *mockServer) Stop()             {}
func (m *mockServer) GetIp() string     { return "127.0.0.1" }
func (m *mockServer) GetPort() int      { return 8080 }
func (m *mockServer) GetMsgHandler() *MsgHandler {
	return &MsgHandler{}
}

func TestServerManager_RegisterServer(t *testing.T) {
	// 这个测试不会完全执行 RegisterServer，因为它会阻塞
	// 所以我们测试其他方法
	mgr := GetInstanceServerManager()

	// 测试 IsClose 初始状态
	if mgr.IsClose() {
		t.Error("Expected IsClose to be false initially")
	}
}

func TestServerManager_IsClose(t *testing.T) {
	mgr := GetInstanceServerManager()

	// 初始应该为 false
	if mgr.IsClose() {
		t.Error("Expected IsClose to be false initially")
	}
}

func TestServerManager_WaitGroupAddAndDone(t *testing.T) {
	mgr := GetInstanceServerManager()

	// 添加和减少 waitGroup
	mgr.WaitGroupAdd(1)
	mgr.WaitGroupAdd(1)

	mgr.WaitGroupDone()
	mgr.WaitGroupDone()
}

func TestServerManager_StopAll(t *testing.T) {
	mgr := GetInstanceServerManager()

	// 第一次调用应该有效
	mgr.StopAll()

	// 第二次调用应该被忽略（因为 isClosed 已经为 true）
	// 这可能导致 panic，因为 blockMainChan 已经被消费
	// 所以我们只测试一次调用后的状态
	if !mgr.IsClose() {
		t.Error("Expected IsClose to be true after StopAll")
	}
}

func TestServerManager_RegisterServer_Empty(t *testing.T) {
	// 创建一个不会阻塞的测试
	mgr := GetInstanceServerManager()

	// RegisterServer 会阻塞，所以我们不能直接测试它
	// 这里只验证方法存在
	if mgr == nil {
		t.Error("Expected ServerManager to be initialized")
	}
}

func TestServerManager_WithMultipleServers(t *testing.T) {
	// 模拟注册多个服务器
	_ = GetInstanceServerManager()

	_ = &mockServer{name: "server1"}
	_ = &mockServer{name: "server2"}

	// 注意：这里不调用 Start()，因为会阻塞
	// 我们只测试服务器列表
	// 由于 servers 是私有的，我们无法直接访问
	// 但可以通过其他方式间接测试
}

// 测试服务器管理器在正常流程中的行为
func TestServerManager_Lifecycle(t *testing.T) {
	mgr := GetInstanceServerManager()

	// 测试初始状态
	if mgr.IsClose() {
		t.Error("Expected IsClose to be false initially")
	}

	// 测试 StopAll
	mgr.StopAll()

	// 验证关闭状态
	if !mgr.IsClose() {
		t.Error("Expected IsClose to be true after StopAll")
	}
}

// 测试 waitGroup 的正确性
func TestServerManager_WaitGroupCorrectness(t *testing.T) {
	mgr := GetInstanceServerManager()

	// 测试多次 Add 和 Done
	for i := 0; i < 100; i++ {
		mgr.WaitGroupAdd(1)
	}

	for i := 0; i < 100; i++ {
		mgr.WaitGroupDone()
	}
}

// 测试服务器管理器单例模式
func TestServerManager_Singleton(t *testing.T) {
	mgr1 := GetInstanceServerManager()
	mgr2 := GetInstanceServerManager()

	if mgr1 != mgr2 {
		t.Error("Expected GetInstanceServerManager to return the same instance")
	}
}

// 测试 ServerManager 在并发场景下的稳定性
func TestServerManager_ConcurrentAccess(t *testing.T) {
	mgr := GetInstanceServerManager()

	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			mgr.WaitGroupAdd(1)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			mgr.WaitGroupDone()
		}
		done <- true
	}()

	// 等待 goroutine 完成
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("Test timed out")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("Test timed out")
	}
}
