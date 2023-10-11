package main

import (
	"fmt"
	"github.com/451008604/socketServerFrame/api"
	"github.com/451008604/socketServerFrame/logs"
	"github.com/451008604/socketServerFrame/network"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"io"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

var waitGroup = sync.WaitGroup{}

func main() {
	logs.SetPrintMode(true)
	api.RegisterRouterClient()

	for i := 0; i < 100; i++ {
		waitGroup.Add(1)

		login, _ := proto.Marshal(&pb.PlayerLoginRequest{
			LoginType:   proto.String("quick"),
			Account:     proto.String("eric" + strconv.Itoa(i)),
			PassWord:    proto.String(""),
			ChannelType: proto.String("2"),
		})
		msg := network.NewDataPack().Pack(network.NewMsgPackage(pb.MSgID_PlayerLogin_Req, login))

		go socketClient(msg)
		// go webSocketClient(msg)
	}

	waitGroup.Wait()
}

func socketClient(msgByte []byte) {
	var err error
	conn, _ := net.Dial("tcp", "127.0.0.1:17001")
	go func(dial net.Conn) {
		for {
			dp := network.NewDataPack()
			// 获取消息头信息
			headData := make([]byte, dp.GetHeadLen())
			if _, err = io.ReadFull(dial, headData); err != nil {
				break
			}
			// 通过消息头获取dataLen和Id
			msgData := dp.Unpack(headData)
			if msgData == nil {
				break
			}
			// 通过消息头获取消息body
			if msgData.GetDataLen() > 0 {
				msgData.SetData(make([]byte, msgData.GetDataLen()))
				if _, err = io.ReadFull(dial, msgData.GetData()); logs.PrintLogErr(err) {
					break
				}
			}

			if len(msgData.GetData()) == 0 {
				continue
			}

			router := network.GetInstanceMsgHandler().Apis[pb.MSgID(msgData.GetMsgId())]
			msg := router.GetNewMsg()
			if err = proto.Unmarshal(msgData.GetData(), msg); err != nil {
				println(err.Error())
			}
			// 服务器返回的消息
			logs.PrintLogInfo(fmt.Sprintf("服务器：msgid:%v,data:%v", msgData.GetMsgId(), msg))
		}
	}(conn)

	// 发送消息
	_, _ = conn.Write(msgByte)

	// rand.Seed(time.Now().UnixNano())
	// randomNumber := rand.Intn(5) + 10
	// time.Sleep(time.Second * time.Duration(randomNumber))
	// _ = conn.Close()
	// waitGroup.Done()
}

func webSocketClient(msgByte []byte) {
	conn, _, _ := websocket.DefaultDialer.Dial("ws://127.0.0.1:17002", nil)
	go func(c *websocket.Conn) {
		for {
			msgType, message, _ := c.ReadMessage()
			if msgType == -1 {
				break
			}
			if len(message) != 0 {
				logs.PrintLogInfo(fmt.Sprintf("服务器：%v", string(message)))
			}
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
