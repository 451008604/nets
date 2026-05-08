package benchmark

import (
	"fmt"
	"github.com/451008604/nets"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/451008604/nets/internal"
	"google.golang.org/protobuf/proto"
)

const (
	benchTCPPort  = 18001
	benchWSPort   = 18002
	benchHTTPPort = 18003
	benchKCPPort  = 18004
)

func TestMain(m *testing.M) {
	conf := nets.GetServerConf()
	conf.ConnRWTimeOut = 60
	conf.MaxMsgChanLen = 1000
	conf.WorkerTaskMaxLen = 1000
	conf.MaxPackSize = 2 * 1024 * 1024
	conf.ProtocolIsJson = false
	conf.ServerTCP = nets.ServerConf{Port: benchTCPPort}
	conf.ServerWS = nets.ServerConf{Port: benchWSPort}
	conf.ServerHTTP = nets.ServerConf{Port: benchHTTPPort}
	conf.ServerKCP = nets.ServerConf{Port: benchKCPPort}

	nets.GetInstanceMsgHandler().AddRouter(
		int32(internal.Test_MsgId_Test_Echo),
		func() proto.Message { return &internal.Test_EchoRequest{} },
		func(conn nets.IConnection, message proto.Message) {
			req, ok := message.(*internal.Test_EchoRequest)
			if !ok || req == nil {
				return
			}
			conn.SendMsg(int32(internal.Test_MsgId_Test_Echo), &internal.Test_EchoResponse{Message: req.Message})
		},
	)

	go nets.GetServerTCP().Start()
	go nets.GetServerWS().Start()
	go nets.GetServerKCP().Start()

	echoMux := http.NewServeMux()
	echoMux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Write(body)
	})
	go http.ListenAndServe(fmt.Sprintf(":%d", benchHTTPPort), echoMux)

	time.Sleep(500 * time.Millisecond)

	code := m.Run()

	nets.GetInstanceServerManager().StopAll()
	fmt.Println("benchmarks done, exiting...")
	os.Exit(code)
}
