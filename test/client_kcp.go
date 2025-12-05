package main

import (
	"encoding/json"
	"fmt"
	"github.com/451008604/nets/network"
	pb "github.com/451008604/nets/proto/bin"
	"github.com/xtaci/kcp-go"
	"net"
)

/*
测试1
*/

func main() {
	msg, _ := json.Marshal(&pb.EchoRequest{Message: "hello"})
	data := network.NewDataPack().Pack(network.NewMsgPackage(int32(pb.MsgId_Echo), msg))

	sendWebSocketMessage := func(data []byte) {
		conn, err := kcp.DialWithOptions("127.0.0.1:17004", nil, 0, 0)
		if err != nil {
			println(err.Error())
			return
		}
		// conn.SetNoDelay(1, 10, 2, 1)
		// defer conn.Close()

		go func(c net.Conn) {
			for {
				buf := make([]byte, 4096)
				if message, e := c.Read(buf); e == nil {
					pack := network.NewDataPack().UnPack(buf[:message])
					fmt.Printf("服务器：%v - %s\n", pack.GetMsgId(), pack.GetData())
				} else {
					fmt.Printf("%v\n", err.Error())
					break
				}
			}
		}(conn)

		// 发送消息
		if _, err := conn.Write(data); err != nil {
			println(err.Error())
		}

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
