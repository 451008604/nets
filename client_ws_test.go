package nets

import (
	"encoding/json"
	"fmt"
	"github.com/451008604/nets/internal"
	"github.com/gorilla/websocket"
	"testing"
)

/*
测试1
*/

func Test_Client_WS(t *testing.T) {
	msg, _ := json.Marshal(&internal.Test_EchoRequest{Message: "hello"})
	data := NewDataPack().Pack(NewMsgPackage(int32(internal.Test_MsgId_Test_Echo), msg))

	sendWebSocketMessage := func(data []byte) {
		conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:17002", nil)
		if err != nil {
			println(err.Error())
			return
		}
		defer conn.Close()

		go func(c *websocket.Conn) {
			for {
				if _, message, e := c.ReadMessage(); e == nil {
					if len(message) != 0 {
						pack := NewDataPack().UnPack(message)
						fmt.Printf("服务器：%v - %s\n", pack.GetMsgId(), pack.GetData())
					}
				} else {
					break
				}
			}
		}(conn)

		// 发送消息
		_ = conn.WriteMessage(websocket.BinaryMessage, append(append(append(append(data, data...), data...), data...), data...))

		// go func(c *websocket.Conn) {
		// 	time.Sleep(time.Second * time.Duration(3+rand.Intn(2)))
		// 	_ = conn.Close()
		// 	sendWebSocketMessage(data)
		// }(conn)

		select {}
	}

	for i := 0; i < 1; i++ {
		sendWebSocketMessage(data)
	}
}
