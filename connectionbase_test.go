package nets

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestConnectionBase_ProtocolToByte_JSON(t *testing.T) {
	// 设置 JSON 模式
	defaultServer.AppConf.ProtocolIsJson = true

	conn := &ConnectionBase{}

	// 使用 Message 结构体进行测试
	msg := &Message{
		Id:   1,
		Data: "test message",
	}
	result := conn.ProtocolToByte(msg)

	if len(result) == 0 {
		t.Error("ProtocolToByte returned empty byte array")
	}

	// 验证 JSON 格式
	t.Logf("JSON result: %s", string(result))
}

func TestConnectionBase_ByteToProtocol_JSON(t *testing.T) {
	defaultServer.AppConf.ProtocolIsJson = true

	conn := &ConnectionBase{}

	target := &Message{}

	err := conn.ByteToProtocol([]byte(`{"msg_id":1,"data":"test"}`), target)

	if err != nil {
		t.Errorf("ByteToProtocol error: %v", err)
	}

	if target.Id != 1 {
		t.Errorf("Expected Id=1, got Id=%d", target.Id)
	}

	if target.Data != "test" {
		t.Errorf("Expected Data=test, got Data=%s", target.Data)
	}

	t.Logf("Parsed message: Id=%d, Data=%s", target.Id, target.Data)
}

func TestConnectionBase_ByteToProtocol_InvalidJSON(t *testing.T) {
	defaultServer.AppConf.ProtocolIsJson = true

	conn := &ConnectionBase{}

	target := &Message{}
	err := conn.ByteToProtocol([]byte(`invalid json`), target)

	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestConnectionBase_GetProperty(t *testing.T) {
	conn := &ConnectionBase{}
	conn.property = make(map[string]any)

	conn.SetProperty("key1", "value1")

	result := conn.GetProperty("key1")
	if result != "value1" {
		t.Errorf("Expected value1, got %v", result)
	}

	nonExistent := conn.GetProperty("nonExistent")
	if nonExistent != nil {
		t.Error("Expected nil for non-existent key")
	}
}

func TestConnectionBase_SetProperty(t *testing.T) {
	conn := &ConnectionBase{}
	conn.property = make(map[string]any)

	conn.SetProperty("key1", "value1")
	conn.SetProperty("key2", 123)
	conn.SetProperty("key3", true)

	if conn.GetProperty("key1") != "value1" {
		t.Error("Property key1 not set correctly")
	}
	if conn.GetProperty("key2") != 123 {
		t.Error("Property key2 not set correctly")
	}
	if conn.GetProperty("key3") != true {
		t.Error("Property key3 not set correctly")
	}
}

func TestConnectionBase_RemoveProperty(t *testing.T) {
	conn := &ConnectionBase{}
	conn.property = make(map[string]any)

	conn.SetProperty("key1", "value1")

	if conn.GetProperty("key1") == nil {
		t.Error("Property should exist before remove")
	}

	conn.RemoveProperty("key1")

	if conn.GetProperty("key1") != nil {
		t.Error("Property should be removed")
	}
}

func TestConnectionBase_IsClose(t *testing.T) {
	conn := &ConnectionBase{}

	if conn.IsClose() {
		t.Error("Expected IsClose to be false initially")
	}
}

func TestConnectionBase_GetDeadTime(t *testing.T) {
	conn := &ConnectionBase{}

	deadTime := conn.GetDeadTime()
	if deadTime == 0 {
		t.Log("DeadTime is 0 initially (expected)")
	}
}

func TestConnectionBase_SendMsg(t *testing.T) {
	conn := &ConnectionBase{
		msgBuffChan: make(chan []byte, 10),
	}
	conn.property = make(map[string]any)

	// 使用 Message 结构体
	msg := &Message{
		Id:   1,
		Data: "test",
	}

	// 如果连接已关闭，SendMsg 应该返回
	conn.isClosed = 1

	// 这个调用应该安全返回，不 panic
	conn.SendMsg(1, msg)
}

func TestConnectionBase_FlowControl(t *testing.T) {
	conn := &ConnectionBase{}
	conn.property = make(map[string]any)
	conn.limitingTimer = time.Now().UnixMilli()
	conn.limitingCount = 0

	// 默认配置下，MaxFlowSecond 应该是 -1 或较大值
	// 测试限流逻辑

	// 设置限流参数
	defaultServer.AppConf.MaxFlowSecond = 5

	// 发送 6 个请求，第 6 个应该触发限流
	for i := 0; i < 5; i++ {
		result := conn.FlowControl()
		if result {
			t.Errorf("Request %d should not be rate limited", i+1)
		}
	}

	// 第 6 个请求应该触发限流
	result := conn.FlowControl()
	if !result {
		t.Error("Expected 6th request to be rate limited")
	}

	// 重置限流参数
	defaultServer.AppConf.MaxFlowSecond = -1
}

func TestConnectionBase_FlowControl_NoLimit(t *testing.T) {
	conn := &ConnectionBase{}
	conn.property = make(map[string]any)
	conn.limitingTimer = 0
	conn.limitingCount = 0

	// MaxFlowSecond = -1 表示不限流
	defaultServer.AppConf.MaxFlowSecond = -1

	for i := 0; i < 100; i++ {
		result := conn.FlowControl()
		if result {
			t.Errorf("Request %d should not be rate limited", i+1)
		}
	}

	// 重置
	defaultServer.AppConf.MaxFlowSecond = 5
}

func TestConnectionBase_FlowControl_TokenReset(t *testing.T) {
	conn := &ConnectionBase{}
	conn.property = make(map[string]any)
	conn.limitingTimer = time.Now().UnixMilli() - 2000 // 2 秒前
	conn.limitingCount = 100

	defaultServer.AppConf.MaxFlowSecond = 5

	// 因为时间窗口已过，计数器应该重置
	result := conn.FlowControl()
	if result {
		t.Error("Request should not be rate limited after token bucket reset")
	}

	// 重置
	defaultServer.AppConf.MaxFlowSecond = -1
}

func TestConnectionBase_GetConnId(t *testing.T) {
	conn := &ConnectionBase{connId: "test-conn-id"}

	id := conn.GetConnId()
	if id != "test-conn-id" {
		t.Errorf("Expected connId test-conn-id, got %s", id)
	}
}

func TestConnectionBase_DoTask(t *testing.T) {
	conn := &ConnectionBase{
		taskQueue: make(chan func(), 10),
	}
	conn.property = make(map[string]any)

	var executed atomic.Bool

	conn.DoTask(func() {
		executed.Store(true)
	})

	// 注意：任务是在任务协程中执行的，不会立即执行
	// 这里只是测试 DoTask 不会 panic
	time.Sleep(time.Millisecond * 10)
}

func TestConnectionBase_Close(t *testing.T) {
	conn := &ConnectionBase{
		conn: &mockConnection{id: "close-test", deadTime: time.Now().Unix()},
	}

	// 第一次关闭应该返回 true
	result := conn.Close()
	if !result {
		t.Error("Expected Close to return true on first call")
	}

	// 第二次关闭应该返回 false
	result = conn.Close()
	if result {
		t.Error("Expected Close to return false on second call")
	}
}

// 测试连接 ID 生成
func TestConnectionBase_ConnIdGeneration(t *testing.T) {
	// connIdSeed 是全局变量，每次调用都会递增
	id1 := atomicAddUint32(&connIdSeed, 1)
	id2 := atomicAddUint32(&connIdSeed, 1)

	if id1 >= id2 {
		t.Error("Expected connection IDs to be increasing")
	}
}

// 辅助函数：原子增加并返回新值
func atomicAddUint32(addr *uint32, delta uint32) uint32 {
	atomic.AddUint32(addr, delta)
	return *addr
}

// 测试 readerTaskHandler 的基本功能
func TestReaderTaskHandler(t *testing.T) {
	// 初始化 MsgHandler
	GetInstanceMsgHandler()

	// 创建一个模拟消息
	mockMsg := &Message{
		Id:   1,
		Data: "test",
	}

	// 创建一个模拟连接
	mockConn := &mockConnection{
		id:       "task-test",
		deadTime: time.Now().Unix(),
	}

	// 这个调用应该不 panic
	readerTaskHandler(mockConn, mockMsg)
}

func TestReaderTaskHandler_InvalidMsgId(t *testing.T) {
	GetInstanceMsgHandler()

	mockMsg := &Message{
		Id:   9999, // 未注册的 MsgId
		Data: "test",
	}

	mockConn := &mockConnection{
		id:       "invalid-msgid-test",
		deadTime: time.Now().Unix(),
	}

	// 这个调用应该安全处理未注册的 MsgId
	readerTaskHandler(mockConn, mockMsg)
}

func TestReaderTaskHandler_ClosedConnection(t *testing.T) {
	GetInstanceMsgHandler()

	mockMsg := &Message{
		Id:   0,
		Data: "test",
	}

	mockConn := &mockConnection{
		id:       "closed-conn-test",
		closed:   1, // 标记为已关闭
		deadTime: time.Now().Unix(),
	}

	// 连接已关闭，应该直接返回
	readerTaskHandler(mockConn, mockMsg)
}
