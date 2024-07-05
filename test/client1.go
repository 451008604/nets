package main

import (
	"github.com/gorilla/websocket"
	"math/rand"
	"time"
)

/*
测试1
*/

func main() {
	// msg, _ := json.Marshal(&pb.EchoRequest{Message: "hello"})
	// data := network.NewDataPack().Pack(network.NewMsgPackage(int32(pb.MsgId_Echo_Req), msg))

	for i := 0; i < 1000; i++ {
		sendWebSocketMessage(nil)
	}

	select {}
}

func sendWebSocketMessage(data []byte) {
	conn, _, err := websocket.DefaultDialer.Dial("wss://dev.hz751.com:25875", nil)
	if err != nil {
		println(err.Error())
		return
	}

	go func(c *websocket.Conn) {
		for {
			if _, message, e := c.ReadMessage(); e == nil {
				if len(message) != 0 {
					// unPacks := network.NewDataPack().UnPack(message)
					// for _, pack := range unPacks {
					// fmt.Printf("服务器：%v - %s\n", pack.GetMsgId(), pack.GetData())
					// }
				}
			} else {
				_ = c.Close()
				break
			}
		}
	}(conn)

	// 发送消息
	// go func(c *websocket.Conn) {
	// for {
	// if e := conn.WriteMessage(websocket.BinaryMessage, data); e != nil {
	// break
	// }
	// }
	// }(conn)

	// go func(c *websocket.Conn) {
	// 	time.Sleep(time.Second * time.Duration(3))
	// _ = conn.Close()

	time.Sleep(time.Second * time.Duration(3+rand.Intn(2)))
	sendWebSocketMessage(data)
	// }(conn)
}
