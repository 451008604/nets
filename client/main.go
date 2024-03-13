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
	data := network.NewDataPack().Pack(network.NewMsgPackage(int32(pb.MsgId_Echo_Req), msg))

	for i := 0; i < 1; i++ {
		sendWebSocketMessage(data)
	}

	select {}
}

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
					unPack := network.NewDataPack().UnPack(message)
					body := &pb.EchoResponse{}
					_ = proto.Unmarshal(unPack.GetData(), body)
					fmt.Printf("服务器：%v - %s\n", unPack.GetMsgId(), body.Message)
				}
			} else {
				_ = c.Close()
				break
			}
		}
	}(conn)

	// 发送消息
	// go func(c *websocket.Conn) {
	for {
		if e := conn.WriteMessage(websocket.BinaryMessage, data); e != nil {
			// break
		}
	}
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
