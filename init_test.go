package nets

import (
	"encoding/json"
	"github.com/451008604/nets/internal"
	"sync"
)

var msgStr, _ = json.Marshal(&internal.Test_EchoRequest{Message: "hello world"})

var flagSend, flagReceive, flagOpened, flagClosed, flagErrCapture int32

func initTest() {
	instanceMsgHandlerOnce = sync.Once{}
	instanceConnManagerOnce = sync.Once{}
	instanceServerManagerOnce = sync.Once{}
	instanceMsgHandler = GetInstanceMsgHandler()
	instanceConnManager = GetInstanceConnManager()
	instanceServerManager = GetInstanceServerManager()
	flagSend, flagReceive, flagOpened, flagClosed, flagErrCapture = int32(0), int32(0), int32(0), int32(0), int32(0)
}
