package benchmark

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/451008604/nets/internal"
	"google.golang.org/protobuf/proto"
)

func benchTCPEcho(b *testing.B, msgSize int) {
	b.Helper()

	req := &internal.Test_EchoRequest{Message: strings.Repeat("x", msgSize)}
	data, _ := proto.Marshal(req)
	payload := packBenchMsg(1001, data)

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", benchTCPPort), 3*time.Second)
	if err != nil {
		b.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(60 * time.Second))

	b.SetBytes(int64(msgSize))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := conn.Write(payload); err != nil {
			b.Fatalf("write failed: %v", err)
		}
		resp, err := readExact(conn, len(payload))
		if err != nil {
			if err == io.EOF {
				b.Fatalf("connection closed by server")
			}
			b.Fatalf("read failed: %v", err)
		}
		_ = resp
	}
}

func BenchmarkTCP_Echo_16B(b *testing.B)  { benchTCPEcho(b, 16) }
func BenchmarkTCP_Echo_64B(b *testing.B)  { benchTCPEcho(b, 64) }
func BenchmarkTCP_Echo_256B(b *testing.B) { benchTCPEcho(b, 256) }
func BenchmarkTCP_Echo_1KB(b *testing.B)  { benchTCPEcho(b, 1024) }
func BenchmarkTCP_Echo_4KB(b *testing.B)  { benchTCPEcho(b, 4096) }
func BenchmarkTCP_Echo_8KB(b *testing.B)  { benchTCPEcho(b, 8192) }
func BenchmarkTCP_Echo_16KB(b *testing.B) { benchTCPEcho(b, 16384) }
func BenchmarkTCP_Echo_32KB(b *testing.B) { benchTCPEcho(b, 32768) }

func packBenchMsg(msgId int16, data []byte) []byte {
	buf := make([]byte, 4+len(data))
	binary.LittleEndian.PutUint16(buf[0:2], uint16(msgId))
	binary.LittleEndian.PutUint16(buf[2:4], uint16(len(data)))
	copy(buf[4:], data)
	return buf
}

func readExact(conn net.Conn, n int) ([]byte, error) {
	buf := make([]byte, n)
	total := 0
	for total < n {
		nn, err := conn.Read(buf[total:])
		if err != nil {
			return nil, err
		}
		total += nn
	}
	return buf, nil
}
