package nets

import (
	"encoding/json"
	"fmt"
	"github.com/451008604/nets/internal"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// 测试 http 请求兼容性
func TestGetServerHTTP(t *testing.T) {
	connNum := 10000
	for i := 0; i < connNum; i++ {
		// 发送Restful API 模式请求
		reqData := "testpoint"
		request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf(`http://127.0.0.1:%v/testpoint`, defaultServer.AppConf.ServerHTTP.Port), strings.NewReader(reqData))
		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Errorf("%v\n", err)
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if string(body) != `{"msg_id":0,"data":"`+reqData+`"}` {
			t.Error("TestGetServerHTTP", string(body))
		}
		atomic.AddInt32(&flag.flagSend, 1)

		// 发送路由模式请求
		marshal, _ := json.Marshal(NewMsgPackage(int32(internal.Test_MsgId_Test_Echo), msgStr)) // {"msg_id":1001,"data":"{\"Message\":\"hello world\"}"}
		request, _ = http.NewRequest(http.MethodPost, fmt.Sprintf(`http://127.0.0.1:%v/`, defaultServer.AppConf.ServerHTTP.Port), strings.NewReader(string(marshal)))
		client2 := &http.Client{}
		resp, _ = client2.Do(request)
		body, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if string(body) != string(msgStr) {
			t.Error("TestGetServerHTTP", string(body))
		}
		atomic.AddInt32(&flag.flagSend, 1)
	}

	t.Cleanup(func() {
		time.Sleep(time.Second * 3)
		fmt.Printf("%+v\n", flag)
	})
}
