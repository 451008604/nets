package nets

import (
	"github.com/gorilla/websocket"
	"sync"
	"testing"
)

/*
测试1
*/
func ClientWs(t *testing.T, wg *sync.WaitGroup, data []byte) {
	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:17002", nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer func(conn *websocket.Conn) {
		_ = conn.Close()
		wg.Done()
	}(conn)

	go func(c *websocket.Conn) {
		if _, message, _ := c.ReadMessage(); len(message) != 0 {
			if pack := NewDataPack().UnPack(message); pack != nil {
				// t.Logf("服务器：%v - %s\n", pack.GetMsgId(), pack.GetData())
			}
		}
	}(conn)

	// 发送消息
	_ = conn.WriteMessage(websocket.BinaryMessage, append(append(append(append(data, data...), data...), data...), data...))
}
