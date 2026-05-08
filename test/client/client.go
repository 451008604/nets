package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/xtaci/kcp-go"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	serverIP = flag.String("server", "127.0.0.1", "server IP")
	tcpPort  = flag.Int("tcp", 17001, "TCP port")
	wsPort   = flag.Int("ws", 17002, "WebSocket port")
	httpPort = flag.Int("http", 17003, "HTTP port")
	kcpPort  = flag.Int("kcp", 17004, "KCP port")
	proto    = flag.String("proto", "kcp", "protocol: tcp, ws, http, kcp")
	msgId    = flag.Int("msgid", 1001, "message ID")
	msgData  = flag.String("data", `{"Message":"hello world"}`, "message data")
	connNum  = flag.Int("conn", 20000, "number of connections")
)

type Message struct {
	Id      uint16 `protobuf:"bytes,1,opt,name=msg_id,proto3" json:"msg_id"` // 消息Id
	Data    string `protobuf:"bytes,2,opt,name=data,proto3" json:"data"`     // 消息内容
	DataLen uint16 `json:"-"`                                                // 消息长度
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
		Id:      binary.LittleEndian.Uint16(data[0:2]),
		DataLen: binary.LittleEndian.Uint16(data[2:4]),
	}
	if len(data) >= 4+int(msg.DataLen) {
		msg.Data = string(data[4 : 4+msg.DataLen])
	}
	return msg
}

type TCPClient struct {
	conn net.Conn
}

func NewTCPClient(addr string) (*TCPClient, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("TCP NewTCPClient failed: %v\n", err)
		return nil, err
	}
	return &TCPClient{conn: conn}, nil
}

func (c *TCPClient) Write(msgId int16, data []byte) error {
	_, err := c.conn.Write(pack(msgId, data))
	if err != nil {
		fmt.Printf("TCP Write failed: %v\n", err)
		return err
	}
	return nil
}

func (c *TCPClient) Read(buf []byte) (int, error) {
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
		fmt.Printf("WS NewWSClient failed: %v\n", err)
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
		fmt.Printf("WS Read failed: %v\n", err)
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
		fmt.Printf("KCP NewKCPClient failed: %v\n", err)
		return nil, err
	}
	conn.SetNoDelay(1, 10, 2, 1)
	conn.SetWindowSize(128, 128)
	return &KCPClient{conn: conn}, nil
}

func (c *KCPClient) Write(msgId int16, data []byte) error {
	_, err := c.conn.Write(pack(msgId, data))
	if err != nil {
		fmt.Printf("KCP Write Err: %v\n", err)
		return err
	}
	return nil
}

func (c *KCPClient) Read(buf []byte) (int, error) {
	return c.conn.Read(buf)
}

func (c *KCPClient) Close() error {
	return c.conn.Close()
}

type HTTPClient struct {
	url     string
	resp    *http.Response
	client  *http.Client
	respBuf chan string
}

func NewHTTPClient(addr string) *HTTPClient {
	return &HTTPClient{url: addr, client: &http.Client{}, respBuf: make(chan string)}
}

func (c *HTTPClient) Write(msgId int16, data []byte) error {
	marshal, _ := json.Marshal(Message{Id: uint16(msgId), DataLen: uint16(len(data)), Data: string(data)})
	resp, err := c.client.Post(c.url, "application/json", strings.NewReader(string(marshal)))
	if err != nil {
		fmt.Printf("HTTP Write Err: %v\n", err)
		return err
	}
	c.resp = resp
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		fmt.Printf("HTTP Write Err: %v\n", err)
		return err
	}
	c.respBuf <- string(body)
	return nil
}

func (c *HTTPClient) Read(buf []byte) (int, error) {
	str := []byte(<-c.respBuf)
	n := copy(buf, str)
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

func main() {
	flag.Parse()

	// 解析并标准化协议列表，支持多协议混合（如 "http,tcp,ws,kcp"）
	protoList := strings.Split(*proto, ",")
	var protocols []string
	for _, p := range protoList {
		p = strings.TrimSpace(p)
		if p != "" {
			protocols = append(protocols, p)
		}
	}
	fmt.Printf("混合测试协议包含：%v\n", protocols)

	// 并发创建连接，多个协议时均匀分布
	var (
		clients   []Client
		clientsMu sync.Mutex
		wg        sync.WaitGroup
		sem       = make(chan struct{}, 10000) // 并发限制，避免文件描述符耗尽
	)
	for i := 0; i < *connNum; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()
			protoc := protocols[idx%len(protocols)]
			c, err := newClient(protoc, getAddrByProto(protoc))
			if err != nil {
				fmt.Printf("Connection %d failed: %v\n", idx, err)
				return
			}
			clientsMu.Lock()
			clients = append(clients, c)
			clientsMu.Unlock()
		}(i)
	}
	wg.Wait()
	fmt.Printf("%d 个客户端初始化完成\n", len(clients))

	sendCount := int32(0)
	recCount := int32(0)
	for _, c := range clients {
		go func(client Client) {
			buf := make([]byte, 4096)
			for {
				n, err := client.Read(buf)
				if err != nil {
					fmt.Printf("Read error: %v\n", err)
					return // 连接关闭或出错
				}
				if n > 0 {
					if d := unpack(buf[:n]); d != nil {
						addInt32 := int(atomic.AddInt32(&recCount, 1))
						if len(clients) < 5 || addInt32%(len(clients)/5) == 0 {
							fmt.Printf("idx: %v, %v -> %s\n", addInt32, d.Id, d.Data)
						}
						_ = client.Close()
						return // 收到有效响应，退出
					}
				}
				time.Sleep(time.Microsecond)
			}
		}(c)
	}

	// 发送消息
	data := []byte(*msgData)
	for _, c := range clients {
		if err := c.Write(int16(*msgId), data); err != nil {
			fmt.Printf("Write error: %v\n", err)
			continue
		}
		sendCount++
	}

	// 等待全部响应或超时（30秒）
	func() {
		for {
			select {
			case <-time.After(30 * time.Second):
				fmt.Printf("等待响应超时，已收到 %d/%d 个响应\n", recCount, sendCount)
				return
			default:
				if atomic.LoadInt32(&recCount) >= sendCount {
					return
				}
			}
		}
	}()

	fmt.Printf("sendCount: %v, recCount: %v\n", sendCount, recCount)
	fmt.Printf("👉 测试完成🎉\n")
}
