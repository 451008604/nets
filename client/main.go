package main

import (
	"fmt"
	"github.com/451008604/nets/network"
	pb "github.com/451008604/nets/proto/bin"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

func main() {
	msg, _ := proto.Marshal(&pb.EchoRequest{
		Message: "hello",
	})
	data := network.NewDataPack().Pack(network.NewMsgPackage(int32(pb.MSgID_Echo_Req), msg))
	sendWebSocketMessage(data)

	select {}
}

func sendWebSocketMessage(data []byte) {
	conn, _, _ := websocket.DefaultDialer.Dial("ws://ggghq.cn:17002", nil)
	defer func(conn *websocket.Conn) {
		_ = conn.Close()
	}(conn)

	go func(c *websocket.Conn) {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if len(message) != 0 {
					fmt.Printf("服务器：%v", string(message))
				}
			}
		}
	}(conn)

	// 发送消息
	_ = conn.WriteMessage(websocket.BinaryMessage, data)
}
