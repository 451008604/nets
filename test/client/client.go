package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xtaci/kcp-go"
)

var (
	serverIP = flag.String("server", "127.0.0.1", "server IP")
	tcpPort  = flag.Int("tcp", 17001, "TCP port")
	wsPort   = flag.Int("ws", 17002, "WebSocket port")
	httpPort = flag.Int("http", 17003, "HTTP port")
	kcpPort  = flag.Int("kcp", 17004, "KCP port")
	proto    = flag.String("proto", "tcp", "protocol: tcp, ws, http, kcp")
	connNum  = flag.Int("conn", 10, "number of connections")
	msgId    = flag.Int("msgid", 1001, "message ID")
	msgData  = flag.String("data", `{"Message":"hello world"}`, "message data")
	interval = flag.Int("interval", 1000, "interval millisecond per message")
	duration = flag.Int("duration", 0, "run duration in seconds (0 = infinite)")
	memtest  = flag.Bool("memtest", false, "run memory leak test mode")
	mixProto = flag.Bool("mix", true, "mix protocols evenly in memtest mode")
)

const STOP_SEND_BEFORE = 30

type MemStats struct {
	Alloc        uint64 `json:"alloc"`
	TotalAlloc   uint64 `json:"total_alloc"`
	Sys          uint64 `json:"sys"`
	NumGC        uint32 `json:"num_gc"`
	NumGoroutine int    `json:"num_goroutine"`
	Connections  int    `json:"connections"`
}

type Message struct {
	Id      int16
	DataLen int16
	Data    []byte
}

func pack(msgId int16, data []byte) []byte {
	buf := make([]byte, 4+len(data))
	binary.LittleEndian.PutUint16(buf[0:2], uint16(msgId))
	binary.LittleEndian.PutUint16(buf[2:4], uint16(len(data)))
	copy(buf[4:], data)
	return buf
}

func unpack(data []byte) *Message {
	if len(data) < 4 {
		return nil
	}
	msg := &Message{
		Id:      int16(binary.LittleEndian.Uint16(data[0:2])),
		DataLen: int16(binary.LittleEndian.Uint16(data[2:4])),
	}
	if len(data) >= 4+int(msg.DataLen) {
		msg.Data = data[4 : 4+msg.DataLen]
	}
	return msg
}

type TCPClient struct {
	conn net.Conn
}

func NewTCPClient(addr string) (*TCPClient, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &TCPClient{conn: conn}, nil
}

func (c *TCPClient) Write(msgId int16, data []byte) error {
	_, err := c.conn.Write(pack(msgId, data))
	return err
}

func (c *TCPClient) Read(buf []byte) (int, error) {
	_ = c.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 10))
	return c.conn.Read(buf)
}

func (c *TCPClient) Close() error {
	return c.conn.Close()
}

type WSClient struct {
	conn *websocket.Conn
}

func NewWSClient(addr string) (*WSClient, error) {
	conn, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		return nil, err
	}
	return &WSClient{conn: conn}, nil
}

func (c *WSClient) Write(msgId int16, data []byte) error {
	return c.conn.WriteMessage(websocket.BinaryMessage, pack(msgId, data))
}

func (c *WSClient) Read(buf []byte) (int, error) {
	_, message, err := c.conn.ReadMessage()
	if err != nil || len(message) == 0 {
		return 0, err
	}
	if len(message) > len(buf) {
		message = message[:len(buf)]
	}
	copy(buf, message)
	return len(message), nil
}

func (c *WSClient) Close() error {
	return c.conn.Close()
}

type KCPClient struct {
	conn *kcp.UDPSession
}

func NewKCPClient(addr string) (*KCPClient, error) {
	conn, err := kcp.DialWithOptions(addr, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	conn.SetNoDelay(1, 10, 2, 1)
	conn.SetWindowSize(128, 128)
	return &KCPClient{conn: conn}, nil
}

func (c *KCPClient) Write(msgId int16, data []byte) error {
	_, err := c.conn.Write(pack(msgId, data))
	return err
}

func (c *KCPClient) Read(buf []byte) (int, error) {
	_ = c.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 10))
	return c.conn.Read(buf)
}

func (c *KCPClient) Close() error {
	return c.conn.Close()
}

type HTTPClient struct {
	url     string
	resp    *http.Response
	client  *http.Client
	respBuf []byte
	respMu  sync.Mutex
}

func NewHTTPClient(addr string) *HTTPClient {
	return &HTTPClient{url: addr, client: &http.Client{}}
}

func (c *HTTPClient) Write(msgId int16, data []byte) error {
	msg := fmt.Sprintf(`{"msg_id":%d,"data":"%s"}`, msgId, string(data))
	resp, err := c.client.Post(c.url, "application/json", strings.NewReader(msg))
	if err != nil {
		return err
	}
	c.resp = resp
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	c.respMu.Lock()
	c.respBuf = body
	c.respMu.Unlock()
	return nil
}

func (c *HTTPClient) Read(buf []byte) (int, error) {
	c.respMu.Lock()
	defer c.respMu.Unlock()
	if len(c.respBuf) == 0 {
		return 0, nil
	}
	n := copy(buf, c.respBuf)
	c.respBuf = c.respBuf[n:]
	return n, nil
}

func (c *HTTPClient) Close() error {
	if c.resp != nil {
		_ = c.resp.Body.Close()
	}
	return nil
}

type Client interface {
	Write(msgId int16, data []byte) error
	Read(buf []byte) (int, error)
	Close() error
}

func newClient(proto, addr string) (Client, error) {
	// 支持多协议混合 tcp,kcp,http,ws
	protoList := strings.Split(proto, ",")
	if len(protoList) > 1 {
		// 返回第一个协议客户端，多协议由调用方处理
		proto = strings.TrimSpace(protoList[0])
	}
	switch proto {
	case "tcp":
		return NewTCPClient(addr)
	case "ws":
		return NewWSClient(fmt.Sprintf("ws://%s", addr))
	case "kcp":
		return NewKCPClient(addr)
	case "http":
		return NewHTTPClient(fmt.Sprintf("http://%s", addr)), nil
	default:
		return nil, fmt.Errorf("unknown protocol: %s", proto)
	}
}

func getAddrByProto(protocol string) string {
	switch protocol {
	case "tcp":
		return fmt.Sprintf("%s:%d", *serverIP, *tcpPort)
	case "ws":
		return fmt.Sprintf("%s:%d", *serverIP, *wsPort)
	case "http":
		return fmt.Sprintf("%s:%d", *serverIP, *httpPort)
	case "kcp":
		return fmt.Sprintf("%s:%d", *serverIP, *kcpPort)
	default:
		return ""
	}
}

func getMemStats() MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	connCount := 0
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/debug/connections", *serverIP, *httpPort))
	if err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Sscanf(string(body), "Active Connections = %d", &connCount)
	}

	return MemStats{
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		NumGC:        m.NumGC,
		NumGoroutine: runtime.NumGoroutine(),
		Connections:  connCount,
	}
}

func printMemStats(label string, stats MemStats) {
	fmt.Printf("[%s] Memory Stats:\n", label)
	fmt.Printf("  Alloc:        %d bytes (%.2f MB)\n", stats.Alloc, float64(stats.Alloc)/1024/1024)
	fmt.Printf("  TotalAlloc:   %d bytes (%.2f MB)\n", stats.TotalAlloc, float64(stats.TotalAlloc)/1024/1024)
	fmt.Printf("  Sys:          %d bytes (%.2f MB)\n", stats.Sys, float64(stats.Sys)/1024/1024)
	fmt.Printf("  NumGC:        %d\n", stats.NumGC)
	fmt.Printf("  NumGoroutine:  %d\n", stats.NumGoroutine)
	fmt.Printf("  Connections:  %d\n\n", stats.Connections)
}

func main() {
	flag.Parse()

	if *memtest {
		runMemtest()
		return
	}

	protoList := strings.Split(*proto, ",")
	var protocols []string
	for _, p := range protoList {
		p = strings.TrimSpace(p)
		if p != "" {
			protocols = append(protocols, p)
		}
	}
	if len(protocols) == 0 {
		protocols = []string{"tcp"}
	}

	fmt.Printf("Connecting to server with %d connections (protocols: %v)...\n", *connNum, protocols)

	data := []byte(*msgData)

	var clients []Client
	for i := 0; i < *connNum; i++ {
		proto := protocols[i%len(protocols)]
		addr := getAddrByProto(proto)
		c, err := newClient(proto, addr)
		if err != nil {
			fmt.Printf("Connection %d failed: %v\n", i, err)
			continue
		}
		clients = append(clients, c)
		if i%100 == 0 && i > 0 {
			fmt.Printf("Connected %d/%d\n", i, *connNum)
		}
	}
	fmt.Printf("Connected %d/%d connections\n", len(clients), *connNum)

	sendCount := int64(0)
	recvCount := int64(0)

	ticker := time.NewTicker(time.Duration(*interval) * time.Millisecond)
	defer ticker.Stop()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	var maxTotalSend int64 = 0
	if *duration > 0 {
		maxTotalSend = int64(*duration) * int64(*connNum) * 1000 / int64(*interval)
	}

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(1)
	var stopRequested int32
	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			hasWork := false
			for _, c := range clients {
				n, _ := c.Read(buf)
				if n > 0 {
					hasWork = true
					if pack := unpack(buf[:n]); pack != nil {
						atomic.AddInt64(&recvCount, 1)
					}
				}
			}
			if hasWork {
				time.Sleep(time.Millisecond * 10)
			} else if atomic.LoadInt32(&stopRequested) != 0 {
				return
			} else {
				time.Sleep(time.Millisecond * 50)
			}
		}
	}()

	for {
		select {
		case <-done:
			fmt.Printf("\nShutting down... sent=%d recv=%d\n", atomic.LoadInt64(&sendCount), atomic.LoadInt64(&recvCount))
			fmt.Println("Final result:", atomic.LoadInt64(&sendCount), "=", atomic.LoadInt64(&sendCount))
			atomic.StoreInt32(&stopRequested, 1)
			for _, c := range clients {
				_ = c.Close()
			}
			wg.Wait()
			fmt.Printf("Final result... sent=%d recv=%d\n", atomic.LoadInt64(&sendCount), atomic.LoadInt64(&sendCount))
			return
		case <-ticker.C:
			if maxTotalSend > 0 && atomic.LoadInt64(&sendCount) >= maxTotalSend {
				fmt.Printf("\nAll messages sent, waiting for responses...\n")
				atomic.StoreInt32(&stopRequested, 1)
				wg.Wait()
				for i := 0; i < 60; i++ {
					allDone := true
					for _, c := range clients {
						buf := make([]byte, 4096)
						n, _ := c.Read(buf)
						if n > 0 {
							allDone = false
							if unpack(buf[:n]) != nil {
								atomic.AddInt64(&recvCount, 1)
							}
						}
					}
					if allDone || atomic.LoadInt64(&sendCount) == atomic.LoadInt64(&recvCount) {
						break
					}
					time.Sleep(time.Second * 1)
				}
				for _, c := range clients {
					_ = c.Close()
				}
				fmt.Printf("Final result... sent=%d recv=%d\n", atomic.LoadInt64(&sendCount), atomic.LoadInt64(&sendCount))
				return
			}
			for _, c := range clients {
				if err := c.Write(int16(*msgId), data); err != nil {
					fmt.Printf("Write error: %v\n", err)
					continue
				}
				atomic.AddInt64(&sendCount, 1)
			}
			s := atomic.LoadInt64(&sendCount)
			r := atomic.LoadInt64(&recvCount)
			if s%1000 == 0 {
				fmt.Printf("Sent: %d, Recv: %d\n", s, r)
			}
		}
		d := time.Since(start)
		if d.Seconds() > 5 {
			s := atomic.LoadInt64(&sendCount)
			r := atomic.LoadInt64(&recvCount)
			fmt.Printf("Sent: %d, Recv: %d\n", s, r)
			start = time.Now()
		}
	}
}

func runMemtest() {
	protocols := []string{"tcp", "ws", "http", "kcp"}
	protocolCount := len(protocols)

	if *connNum <= 0 {
		fmt.Println("conn must be > 0")
		os.Exit(1)
	}

	fmt.Printf("=== Memory Leak Test ===\n")
	fmt.Printf("Target server: %s\n", *serverIP)
	fmt.Printf("Total connections: %d\n", *connNum)
	fmt.Printf("Protocols: %v\n\n", protocols)

	beforeStats := getMemStats()
	printMemStats("BEFORE", beforeStats)

	var wg sync.WaitGroup
	var successCount int32
	var mu sync.Mutex
	var clients []Client

	testData := []byte(`{"Message":"memory_test"}`)

	fmt.Printf("Creating %d connections...\n", *connNum)

	for i := 0; i < *connNum; i++ {
		var proto string
		if *mixProto {
			proto = protocols[i%protocolCount]
		} else {
			proto = protocols[0]
		}

		addr := getAddrByProto(proto)
		c, err := newClient(proto, addr)
		if err != nil {
			fmt.Printf("Connection %d (%s) failed: %v\n", i, proto, err)
			continue
		}

		if err := c.Write(1001, testData); err != nil {
			fmt.Printf("Write %d (%s) failed: %v\n", i, proto, err)
			c.Close()
			continue
		}

		mu.Lock()
		clients = append(clients, c)
		mu.Unlock()

		atomic.AddInt32(&successCount, 1)

		if atomic.LoadInt32(&successCount)%100 == 0 {
			fmt.Printf("  Created %d/%d connections\n", atomic.LoadInt32(&successCount), *connNum)
		}

		if i%50 == 0 && i > 0 {
			time.Sleep(time.Millisecond * 10)
		}
	}

	fmt.Printf("\nTotal connections created: %d\n", len(clients))

	runtime.GC()
	time.Sleep(time.Second)

	afterStats := getMemStats()
	printMemStats("AFTER", afterStats)

	fmt.Printf("=== Memory Delta ===\n")
	allocDelta := int64(afterStats.Alloc) - int64(beforeStats.Alloc)
	totalAllocDelta := int64(afterStats.TotalAlloc) - int64(beforeStats.TotalAlloc)
	sysDelta := int64(afterStats.Sys) - int64(beforeStats.Sys)

	fmt.Printf("Alloc delta:      %d bytes (%.2f MB)\n", allocDelta, float64(allocDelta)/1024/1024)
	fmt.Printf("TotalAlloc delta: %d bytes (%.2f MB)\n", totalAllocDelta, float64(totalAllocDelta)/1024/1024)
	fmt.Printf("Sys delta:      %d bytes (%.2f MB)\n", sysDelta, float64(sysDelta)/1024/1024)
	fmt.Printf("Goroutines:    %d -> %d (delta: %d)\n", beforeStats.NumGoroutine, afterStats.NumGoroutine, afterStats.NumGoroutine-beforeStats.NumGoroutine)
	fmt.Printf("GC count delta: %d\n", afterStats.NumGC-beforeStats.NumGC)

	fmt.Printf("\n=== Closing connections ===\n")
	for i, c := range clients {
		c.Close()
		if (i+1)%200 == 0 {
			fmt.Printf("  Closed %d/%d\n", i+1, len(clients))
		}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
	}()

	time.Sleep(time.Second * 2)

	runtime.GC()
	time.Sleep(time.Second)

	closedStats := getMemStats()
	printMemStats("CLOSED", closedStats)

	closedDelta := int64(closedStats.Alloc) - int64(beforeStats.Alloc)

	fmt.Printf("=== Final Analysis ===\n")
	fmt.Printf("Memory after close: %d bytes (%.2f MB)\n", closedStats.Alloc, float64(closedStats.Alloc)/1024/1024)
	fmt.Printf("Memory leaked:  %d bytes (%.2f MB)\n", closedDelta, float64(closedDelta)/1024/1024)

	if closedDelta > 10*1024*1024 {
		fmt.Printf("\nWARNING: Potential memory leak detected (>10MB)\n")
	} else if closedDelta > 5*1024*1024 {
		fmt.Printf("\nCAUTION: Elevated memory usage (>5MB)\n")
	} else {
		fmt.Printf("\nOK Memory usage looks normal\n")
	}

	if afterStats.NumGoroutine > beforeStats.NumGoroutine+50 {
		fmt.Printf("WARNING: Goroutine leak detected\n")
	}
}
