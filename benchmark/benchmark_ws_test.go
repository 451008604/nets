package benchmark

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/451008604/nets/internal"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

func benchWSEcho(b *testing.B, msgSize int) {
	b.Helper()

	req := &internal.Test_EchoRequest{Message: strings.Repeat("x", msgSize)}
	data, _ := proto.Marshal(req)
	payload := packBenchMsg(1001, data)

	conn, _, err := websocket.DefaultDialer.Dial(
		fmt.Sprintf("ws://127.0.0.1:%d", benchWSPort), nil,
	)
	if err != nil {
		b.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(60 * time.Second))

	b.SetBytes(int64(msgSize))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := conn.WriteMessage(websocket.BinaryMessage, payload); err != nil {
			b.Fatalf("write failed: %v", err)
		}
		if _, _, err := conn.ReadMessage(); err != nil {
			b.Fatalf("read failed: %v", err)
		}
	}
}

func BenchmarkWS_Echo_16B(b *testing.B)  { benchWSEcho(b, 16) }
func BenchmarkWS_Echo_64B(b *testing.B)  { benchWSEcho(b, 64) }
func BenchmarkWS_Echo_256B(b *testing.B) { benchWSEcho(b, 256) }
func BenchmarkWS_Echo_1KB(b *testing.B)  { benchWSEcho(b, 1024) }
func BenchmarkWS_Echo_4KB(b *testing.B)  { benchWSEcho(b, 4096) }
func BenchmarkWS_Echo_8KB(b *testing.B)  { benchWSEcho(b, 8192) }
func BenchmarkWS_Echo_16KB(b *testing.B) { benchWSEcho(b, 16384) }
func BenchmarkWS_Echo_32KB(b *testing.B) { benchWSEcho(b, 32768) }
