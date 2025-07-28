package main

import (
	"encoding/json"
	"fmt"
	"github.com/451008604/nets/network"
	pb "github.com/451008604/nets/proto/bin"
	"github.com/gorilla/websocket"
)

/*
测试1
*/

func main() {
	msg, _ := json.Marshal(&pb.EchoRequest{Message: "hello"})
	data := network.NewDataPack().Pack(network.NewMsgPackage(int32(pb.MsgId_Echo), msg))

	for i := 0; i < 100; i++ {
		sendWebSocketMessage(data)
	}

	select {}
}

func sendWebSocketMessage(data []byte) {
	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:17002", nil)
	if err != nil {
		println(err.Error())
		return
	}

	go func(c *websocket.Conn) {
		for {
			if _, message, e := c.ReadMessage(); e == nil {
				if len(message) != 0 {
					pack := network.NewDataPack().UnPack(message)
					fmt.Printf("服务器：%v - %s\n", pack.GetMsgId(), pack.GetData())
				}
			} else {
				_ = c.Close()
				break
			}
		}
	}(conn)

	// 发送消息
	go func(c *websocket.Conn) {
		if e := conn.WriteMessage(websocket.BinaryMessage, data); e != nil {
			_ = conn.Close()
			return
		}
	}(conn)

	// go func(c *websocket.Conn) {
	// 	time.Sleep(time.Second * time.Duration(3+rand.Intn(2)))
	// 	_ = conn.Close()
	// 	sendWebSocketMessage(data)
	// }(conn)
}
