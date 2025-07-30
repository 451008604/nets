package main

import (
	"encoding/json"
	"fmt"
	"github.com/451008604/nets/network"
	pb "github.com/451008604/nets/proto/bin"
	"github.com/xtaci/kcp-go"
	"net"
	"time"
)

/*
测试1
*/

func main() {
	msg, _ := json.Marshal(&pb.EchoRequest{Message: "hello"})
	data := network.NewDataPack().Pack(network.NewMsgPackage(int32(pb.MsgId_Echo), msg))

	sendWebSocketMessage := func(data []byte) {
		conn, err := kcp.Dial("127.0.0.1:17004")
		if err != nil {
			println(err.Error())
			return
		}
		defer conn.Close()

		go func(c net.Conn) {
			for {
				buf := make([]byte, 4096)
				if message, e := c.Read(buf); e == nil {
					pack := network.NewDataPack().UnPack(buf[:message])
					fmt.Printf("服务器：%v - %s\n", pack.GetMsgId(), pack.GetData())
				} else {
					break
				}
			}
		}(conn)

		// 发送消息
		_, _ = conn.Write(append(append(append(append(data, data...), data...), data...), data...))
		// _, _ = conn.Write(append(data, data...))
		// _, _ = conn.Write(append(data, data...))
		// _, _ = conn.Write(append(data, data...))
		// _, _ = conn.Write(append(data, data...))

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
