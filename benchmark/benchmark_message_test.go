package benchmark

import (
	"encoding/json"
	"fmt"
	"github.com/451008604/nets"
	"testing"

	"github.com/451008604/nets/internal"
	"google.golang.org/protobuf/proto"
)

func BenchmarkNewMsgPackage(b *testing.B) {
	sizes := []int{16, 64, 256, 1024}
	for _, size := range sizes {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i % 256)
		}
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				msg := nets.NewMsgPackage(1001, data)
				_ = msg
			}
		})
	}
}

func BenchmarkMessageGetPut(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		msg := nets.GetMessage()
		msg.Id = 1001
		msg.Data = []byte("hello")
		msg.DataLen = 5
		nets.PutMessage(msg)
	}
}

type benchProtoMsg struct {
	MsgId   int32  `json:"msg_id"`
	Message string `json:"message"`
}

func BenchmarkJSONMarshal(b *testing.B) {
	sizes := []int{16, 64, 256, 1024}
	for _, size := range sizes {
		msg := &benchProtoMsg{MsgId: 1001, Message: fmt.Sprintf("%-*s", size, "x")}
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				data, _ := json.Marshal(msg)
				_ = data
			}
		})
	}
}

func BenchmarkJSONUnmarshal(b *testing.B) {
	sizes := []int{16, 64, 256, 1024}
	for _, size := range sizes {
		msg := &benchProtoMsg{MsgId: 1001, Message: fmt.Sprintf("%-*s", size, "x")}
		data, _ := json.Marshal(msg)
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var m benchProtoMsg
				_ = json.Unmarshal(data, &m)
			}
		})
	}
}

func BenchmarkProtoMarshal(b *testing.B) {
	sizes := []int{16, 64, 256, 1024}
	for _, size := range sizes {
		payload := fmt.Sprintf("%-*s", size, "x")
		msg := &internal.Test_EchoRequest{Message: payload}
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				data, _ := proto.Marshal(msg)
				_ = data
			}
		})
	}
}

func BenchmarkProtoUnmarshal(b *testing.B) {
	sizes := []int{16, 64, 256, 1024}
	for _, size := range sizes {
		payload := fmt.Sprintf("%-*s", size, "x")
		msg := &internal.Test_EchoRequest{Message: payload}
		data, _ := proto.Marshal(msg)
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var m internal.Test_EchoRequest
				_ = proto.Unmarshal(data, &m)
			}
		})
	}
}
