package main

import (
	"encoding/json"
	"fmt"
	"github.com/451008604/nets/network"
	pb "github.com/451008604/nets/proto/bin"
	"github.com/gorilla/websocket"
	"sync/atomic"
)

func main() {
	msg, _ := json.Marshal(&pb.EchoRequest{
		Message: "hello",
	})
	data := network.NewDataPack().Pack(network.NewMsgPackage(int32(pb.MsgId_Echo_Req), msg))

	for i := 0; i < 1000; i++ {
		go sendWebSocketMessage(data)
	}

	select {}
}

var n = uint32(0)

func sendWebSocketMessage(data []byte) {
	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:17002", nil)
	if err != nil {
		return
	}

	go func(c *websocket.Conn) {
		for {
			_, message, e := c.ReadMessage()
			if e == nil {
				if len(message) != 0 {
					unPacks := network.NewDataPack().UnPack(message)
					atomic.AddUint32(&n, 1)
					for _, pack := range unPacks {
						fmt.Printf("%v - 服务器：%v - %s\n", n, pack.GetMsgId(), pack.GetData())
					}
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
	if e := conn.WriteMessage(websocket.BinaryMessage, data); e != nil {
		// break
	}
	// }
	// }(conn)

	// go func(c *websocket.Conn) {
	// 	intn := rand.Intn(10)
	// 	time.Sleep(time.Second * time.Duration(5+intn))
	// 	_ = conn.Close()
	//
	// 	// time.Sleep(time.Second * time.Duration(intn))
	// 	sendWebSocketMessage(data)
	// }(conn)
}
