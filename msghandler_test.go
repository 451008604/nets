package nets

import (
	"github.com/451008604/nets/internal"
	"google.golang.org/protobuf/proto"
	"sync/atomic"
	"testing"
	"time"
)

// 测试 taskFun 发生panic 时异常捕获
func TestMsgHandler_SetErrCapture(t *testing.T) {
	initTest()

	instanceMsgHandler.SetErrCapture(func(conn IConnection, panicInfo string) {
		atomic.AddInt32(&flagErrCapture, 1)
	})

	connection := NewConnectionTest()
	instanceConnManager.Add(connection)

	// panicInfo Test_MsgId_Test_Echo panic
	instanceMsgHandler.AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, func(conn IConnection, message proto.Message) {
		panic("Test_MsgId_Test_Echo panic")
	})
	connection.SetProperty("msgReq", defaultServer.DataPack.Pack(NewMsgPackage(int32(internal.Test_MsgId_Test_Echo), msgStr)))
	// panicInfo runtime error: integer divide by zero
	connection.DoTask(func() {
		n := int32(0)
		_ = 3 / int32(n)
	})
	// panicInfo runtime error: invalid memory address or nil pointer dereference
	connection.DoTask(func() {
		type testStruct struct {
			n int32
		}
		var test *testStruct
		test.n = 10
	})

	time.Sleep(time.Second)
	if flagErrCapture != int32(3) {
		t.Error("TestMsgHandler_SetErrCapture", flagErrCapture)
		return
	}
}

func TestMsgHandler_SetFilter(t *testing.T) {
	initTest()

	instanceMsgHandler.SetFilter(func(conn IConnection, msg IMessage) bool {
		conn.SetProperty("filterKey", "filterValue")
		return true
	})

	connection := NewConnectionTest()
	instanceConnManager.Add(connection)
	instanceMsgHandler.AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, func(conn IConnection, message proto.Message) {
		if v, ok := conn.GetProperty("filterKey").(string); ok {
			if v != "filterValue" {
				t.Error("TestMsgHandler_SetFilter", v)
			}
		}
	})
	connection.SetProperty("msgReq", defaultServer.DataPack.Pack(NewMsgPackage(int32(internal.Test_MsgId_Test_Echo), msgStr)))
}

func TestMsgHandler_AddRouter(t *testing.T) {
	initTest()
	instanceMsgHandler.AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, func(conn IConnection, message proto.Message) {})
	instanceMsgHandler.AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, func(conn IConnection, message proto.Message) {})
}
