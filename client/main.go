package main

import (
	"encoding/json"
	"github.com/451008604/socketServerFrame/client/base"
	"github.com/451008604/socketServerFrame/logs"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"sync"
)

func main() {
	logs.SetPrintMode(true)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	for n := 0; n < 1; n++ {
		go func(n int) {
			conn := &base.CustomConnect{}
			conn.NewConnection("127.0.0.1", "7777")
			defer conn.SetBlocking()
			// go func(n int) {
			//	i := 0
			//	for {
			//		i++
			//
			//		data := &pb.Ping{TimeStamp: time.Now().UnixMicro()}
			//		marshal, err := proto.Marshal(data)
			//		if err != nil {
			//			return
			//		}
			//		conn.SendMsg(pb.MessageID_PING, marshal)
			//		logs.PrintLogInfo(data.String())
			//		time.Sleep(5 * time.Second)
			//	}
			// }(n)

			// login
			login, _ := json.Marshal(&pb.ReqLogin{
				UserName: "guohaoqin",
				PassWord: "1234567",
			})
			conn.SendMsg(pb.MessageID_Login, login)
		}(n)
	}
	wg.Wait()
}
