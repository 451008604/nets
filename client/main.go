package main

import (
	"encoding/json"
	"github.com/451008604/socketServerFrame/client/base"
	"github.com/451008604/socketServerFrame/logs"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
)

func main() {
	logs.SetPrintMode(true)

	// socketClient()
	WebSocketClient()
}

func socketClient() {
	conn := &base.CustomConnect{}
	conn.NewConnection("127.0.0.1", "7777")
	defer conn.SetBlocking()

	// login
	login, _ := json.Marshal(&pb.ReqLogin{
		UserName: "guohaoqin",
		PassWord: "1234567",
	})
	conn.SendMsg(pb.MessageID_Login, login)
}

func WebSocketClient() {
	u := url.URL{Scheme: "ws", Host: "localhost:8086", Path: "/"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer func(c *websocket.Conn) {
		_ = c.Close()
	}(c)
	// login
	login, _ := json.Marshal(&pb.ReqLogin{
		UserName: "guohaoqin",
		PassWord: "1234567",
	})
	err = c.WriteMessage(websocket.TextMessage, login)
	if err != nil {
		log.Fatalln("write:", err)
	}

	_, message, err := c.ReadMessage()
	if err != nil {
		log.Fatalln("read:", err)
	}

	log.Printf("Received: %s\n", message)
}
