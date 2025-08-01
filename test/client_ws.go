package main

import (
	"encoding/json"
	"fmt"
	"github.com/451008604/nets/network"
	pb "github.com/451008604/nets/proto/bin"
	"github.com/gorilla/websocket"
	"time"
)

/*
测试1
*/

func main() {
	msg, _ := json.Marshal(&pb.EchoRequest{Message: "hello"})
	data := network.NewDataPack().Pack(network.NewMsgPackage(int32(pb.MsgId_Echo), msg))

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
						pack := network.NewDataPack().UnPack(message)
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

		time.Sleep(time.Second * 10)
	}

	for i := 0; i < 1; i++ {
		sendWebSocketMessage(data)
	}
}
