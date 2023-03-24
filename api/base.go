package api

import (
	"encoding/json"
	"github.com/451008604/socketServerFrame/config"
	"github.com/451008604/socketServerFrame/logs"
	"google.golang.org/protobuf/proto"
)

func ProtocolToByte(str proto.Message) []byte {
	var err error
	var marshal []byte

	if config.GetGlobalObject().ProtocolIsJson {
		marshal, err = json.Marshal(str)
	} else {
		marshal, err = proto.Marshal(str)
	}

	if err != nil {
		logs.PrintLogErr(err)
		return []byte{}
	}
	return marshal
}

func ByteToProtocol(byte []byte, target proto.Message) {
	var err error

	if config.GetGlobalObject().ProtocolIsJson {
		err = json.Unmarshal(byte, target)
	} else {
		err = proto.Unmarshal(byte, target)
	}

	if err != nil {
		logs.PrintLogErr(err)
		return
	}
}
