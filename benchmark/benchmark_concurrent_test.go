package benchmark

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/451008604/nets/internal"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

func BenchmarkTCP_ConcurrentConnections(b *testing.B) {
	connCounts := []int{10, 50, 100, 500}
	for _, connCount := range connCounts {
		b.Run(fmt.Sprintf("conns_%d", connCount), func(b *testing.B) {
			req := &internal.Test_EchoRequest{Message: strings.Repeat("x", 64)}
			data, _ := proto.Marshal(req)
			payload := packBenchMsg(1001, data)

			conns := make([]net.Conn, connCount)
			for i := 0; i < connCount; i++ {
				conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", benchTCPPort), 3*time.Second)
				if err != nil {
					b.Fatalf("dial conn %d failed: %v", i, err)
				}
				defer conn.Close()
				conn.SetDeadline(time.Now().Add(120 * time.Second))
				conns[i] = conn
			}

			b.SetBytes(64)
			b.ReportAllocs()
			b.ResetTimer()

			var ops atomic.Int64
			var wg sync.WaitGroup
			perConn := b.N / connCount
			if perConn < 1 {
				perConn = 1
			}

			for i := 0; i < connCount; i++ {
				wg.Add(1)
				go func(c net.Conn) {
					defer wg.Done()
					for j := 0; j < perConn; j++ {
						if _, err := c.Write(payload); err != nil {
							return
						}
						if _, err := readExact(c, len(payload)); err != nil {
							return
						}
						ops.Add(1)
					}
				}(conns[i])
			}
			wg.Wait()

			totalOps := ops.Load()
			b.ReportMetric(float64(totalOps)/b.Elapsed().Seconds(), "ops/sec")
		})
	}
}

func BenchmarkWS_ConcurrentConnections(b *testing.B) {
	connCounts := []int{10, 50, 100}
	for _, connCount := range connCounts {
		b.Run(fmt.Sprintf("conns_%d", connCount), func(b *testing.B) {
			req := &internal.Test_EchoRequest{Message: strings.Repeat("x", 64)}
			data, _ := proto.Marshal(req)
			payload := packBenchMsg(1001, data)

			conns := make([]*websocket.Conn, connCount)
			for i := 0; i < connCount; i++ {
				c, _, err := websocket.DefaultDialer.Dial(
					fmt.Sprintf("ws://127.0.0.1:%d", benchWSPort), nil,
				)
				if err != nil {
					b.Fatalf("ws dial conn %d failed: %v", i, err)
				}
				defer c.Close()
				c.SetReadDeadline(time.Now().Add(120 * time.Second))
				c.SetWriteDeadline(time.Now().Add(120 * time.Second))
				conns[i] = c
			}

			b.SetBytes(64)
			b.ReportAllocs()
			b.ResetTimer()

			var ops atomic.Int64
			var wg sync.WaitGroup
			perConn := b.N / connCount
			if perConn < 1 {
				perConn = 1
			}

			for i := 0; i < connCount; i++ {
				wg.Add(1)
				go func(c *websocket.Conn) {
					defer wg.Done()
					for j := 0; j < perConn; j++ {
						if err := c.WriteMessage(websocket.BinaryMessage, payload); err != nil {
							return
						}
						if _, _, err := c.ReadMessage(); err != nil {
							return
						}
						ops.Add(1)
					}
				}(conns[i])
			}
			wg.Wait()

			totalOps := ops.Load()
			b.ReportMetric(float64(totalOps)/b.Elapsed().Seconds(), "ops/sec")
		})
	}
}
