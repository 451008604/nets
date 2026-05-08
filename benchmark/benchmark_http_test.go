package benchmark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

type benchHTTPMsg struct {
	MsgId   uint16 `json:"msg_id"`
	DataLen uint16 `json:"-"`
	Data    string `json:"data"`
}

func benchHTTPEcho(b *testing.B, msgSize int) {
	b.Helper()

	data := make([]byte, msgSize)
	for i := range data {
		data[i] = byte(i % 256)
	}
	msg := benchHTTPMsg{MsgId: 1001, Data: string(data)}
	body, _ := json.Marshal(msg)

	client := &http.Client{Timeout: 60 * time.Second}
	url := fmt.Sprintf("http://127.0.0.1:%d/echo", benchHTTPPort)

	b.SetBytes(int64(msgSize))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := client.Post(url, "application/json", bytes.NewReader(body))
		if err != nil {
			b.Fatalf("post failed: %v", err)
		}
		if _, err := io.ReadAll(resp.Body); err != nil {
			b.Fatalf("read body failed: %v", err)
		}
		resp.Body.Close()
	}
}

func BenchmarkHTTP_Echo_16B(b *testing.B)  { benchHTTPEcho(b, 16) }
func BenchmarkHTTP_Echo_64B(b *testing.B)  { benchHTTPEcho(b, 64) }
func BenchmarkHTTP_Echo_256B(b *testing.B) { benchHTTPEcho(b, 256) }
func BenchmarkHTTP_Echo_1KB(b *testing.B)  { benchHTTPEcho(b, 1024) }
func BenchmarkHTTP_Echo_4KB(b *testing.B)  { benchHTTPEcho(b, 4096) }
func BenchmarkHTTP_Echo_8KB(b *testing.B)  { benchHTTPEcho(b, 8192) }
func BenchmarkHTTP_Echo_16KB(b *testing.B) { benchHTTPEcho(b, 16384) }
func BenchmarkHTTP_Echo_32KB(b *testing.B) { benchHTTPEcho(b, 32768) }
