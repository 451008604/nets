package benchmark

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/451008604/nets/internal"
	"github.com/xtaci/kcp-go"
	"google.golang.org/protobuf/proto"
)

func benchKCPEcho(b *testing.B, msgSize int) {
	b.Helper()

	req := &internal.Test_EchoRequest{Message: strings.Repeat("x", msgSize)}
	data, _ := proto.Marshal(req)
	payload := packBenchMsg(1001, data)

	conn, err := kcp.DialWithOptions(
		fmt.Sprintf("127.0.0.1:%d", benchKCPPort), nil, 0, 0,
	)
	if err != nil {
		b.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()
	conn.SetNoDelay(1, 10, 2, 1)
	conn.SetWindowSize(128, 128)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	b.SetBytes(int64(msgSize))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := conn.Write(payload); err != nil {
			b.Fatalf("write failed: %v", err)
		}
		resp, err := readExact(conn, len(payload))
		if err != nil {
			b.Fatalf("read failed: %v", err)
		}
		_ = resp
	}
}

func BenchmarkKCP_Echo_16B(b *testing.B)  { benchKCPEcho(b, 16) }
func BenchmarkKCP_Echo_64B(b *testing.B)  { benchKCPEcho(b, 64) }
func BenchmarkKCP_Echo_256B(b *testing.B) { benchKCPEcho(b, 256) }
func BenchmarkKCP_Echo_1KB(b *testing.B)  { benchKCPEcho(b, 1024) }
func BenchmarkKCP_Echo_4KB(b *testing.B)  { benchKCPEcho(b, 4096) }
func BenchmarkKCP_Echo_8KB(b *testing.B)  { benchKCPEcho(b, 8192) }
func BenchmarkKCP_Echo_16KB(b *testing.B) { benchKCPEcho(b, 16384) }
func BenchmarkKCP_Echo_32KB(b *testing.B) { benchKCPEcho(b, 32768) }
