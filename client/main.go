package main

import (
	"encoding/json"
	"fmt"
	"github.com/451008604/socketServerFrame/logs"
	"github.com/451008604/socketServerFrame/network"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"github.com/gorilla/websocket"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"
)

var waitGroup = sync.WaitGroup{}

func main() {
	logs.SetPrintMode(true)

	login, _ := json.Marshal(&pb.ReqLogin{
		UserName: "guohaoqin",
		PassWord: "1234567",
	})
	msg := network.NewDataPack().Pack(network.NewMsgPackage(pb.MessageID_Login, login))

	for i := 0; i < 100; i++ {
		waitGroup.Add(2)
		go socketClient(msg)
		go WebSocketClient(msg)
	}

	waitGroup.Wait()
}

func socketClient(msgByte []byte) {
	var err error
	conn, _ := net.Dial("tcp", "127.0.0.1:7001")
	go func(dial net.Conn) {
		for {
			dp := network.NewDataPack()
			// 获取消息头信息
			headData := make([]byte, dp.GetHeadLen())
			_, err = io.ReadFull(dial, headData)
			if err != io.EOF {
				break
			}
			// 获取消息body
			msgData := dp.Unpack(headData)
			if msgData.GetDataLen() > 0 {
				msgData.SetData(make([]byte, msgData.GetDataLen()))
				_, _ = io.ReadFull(dial, msgData.GetData())
			}

			if len(msgData.GetData()) == 0 {
				continue
			}

			// 服务器返回的消息
			logs.PrintLogInfo(fmt.Sprintf("服务器：%v", string(msgData.GetData())))
		}
	}(conn)

	// 发送消息
	_, _ = conn.Write(msgByte)

	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(5) + 10
	time.Sleep(time.Second * time.Duration(randomNumber))
	_ = conn.Close()
	waitGroup.Done()
}

func WebSocketClient(msgByte []byte) {
	conn, _, _ := websocket.DefaultDialer.Dial("ws://127.0.0.1:7002", nil)
	go func(c *websocket.Conn) {
		_, message, _ := c.ReadMessage()
		if len(message) != 0 {
			logs.PrintLogInfo(fmt.Sprintf("服务器：%v", string(message)))
		}
	}(conn)

	// 发送消息
	_ = conn.WriteMessage(websocket.BinaryMessage, msgByte)

	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(5) + 10
	time.Sleep(time.Second * time.Duration(randomNumber))
	_ = conn.Close()
	waitGroup.Done()
}
