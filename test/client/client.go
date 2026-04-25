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
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"github.com/gorilla/websocket"
	"github.com/xtaci/kcp-go"
)

var (
	serverIP = flag.String("server", "127.0.0.1", "server IP")
	tcpPort  = flag.Int("tcp", 17001, "TCP port")
	wsPort   = flag.Int("ws", 17002, "WebSocket port")
	httpPort = flag.Int("http", 17003, "HTTP port")
	kcpPort  = flag.Int("kcp", 17004, "KCP port")
	proto   = flag.String("proto", "tcp", "protocol: tcp, ws, http, kcp")
	connNum  = flag.Int("conn", 10, "number of connections")
	msgId    = flag.Int("msgid", 1001, "message ID")
	msgData  = flag.String("data", `{"Message":"hello world"}`, "message data")
	interval = flag.Int("interval", 1000, "interval millisecond per message")
	duration = flag.Int("duration", 0, "run duration in seconds (0 = infinite)")
)

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

func getAddr() string {
	switch *proto {
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

func main() {
	flag.Parse()

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

	sendCount := 0
	recvCount := 0

	ticker := time.NewTicker(time.Duration(*interval) * time.Millisecond)
	defer ticker.Stop()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	if *duration > 0 {
		go func() {
			time.Sleep(time.Duration(*duration) * time.Second)
			done <- syscall.SIGTERM
		}()
	}

start := time.Now()
	go func() {
		buf := make([]byte, 4096)
		for {
			for _, c := range clients {
				n, err := c.Read(buf)
				if err != nil || n == 0 {
					continue
				}
				if pack := unpack(buf[:n]); pack != nil {
					atomic.AddInt32((*int32)(unsafe.Pointer(&recvCount)), 1)
				}
			}
			time.Sleep(time.Millisecond * 10)
		}
	}()

	for {
		select {
		case <-done:
			fmt.Printf("\nShutting down... sent=%d recv=%d\n", sendCount, recvCount)
			for _, c := range clients {
				_ = c.Close()
			}
			os.Exit(0)
		case <-ticker.C:
			for _, c := range clients {
				if err := c.Write(int16(*msgId), data); err != nil {
					fmt.Printf("Write error: %v\n", err)
					continue
				}
				sendCount++
			}
			if sendCount%1000 == 0 {
				fmt.Printf("Sent: %d, Recv: %d\n", sendCount, recvCount)
			}
		}
		d := time.Since(start)
		if d.Seconds() > 5 {
			fmt.Printf("Sent: %d, Recv: %d\n", sendCount, recvCount)
			start = time.Now()
		}
	}
}
