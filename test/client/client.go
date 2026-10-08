package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/xtaci/kcp-go"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	serverIP        = flag.String("server", "127.0.0.1", "server IP")
	tcpPort         = flag.Int("tcp", 17001, "TCP port")
	wsPort          = flag.Int("ws", 17002, "WebSocket port")
	httpPort        = flag.Int("http", 17003, "HTTP port")
	kcpPort         = flag.Int("kcp", 17004, "KCP port")
	proto           = flag.String("proto", "tcp,ws,http,kcp", "protocol: tcp, ws, http, kcp")
	msgId           = flag.Int("msgid", 1001, "message ID")
	msgData         = flag.String("data", `{"Message":"hello world"}`, "message data")
	connNum         = flag.Int("conn", 1, "number of connections")
	useTLS          = flag.Bool("tls", false, "enable verified TLS for selected protocols")
	tlsCA           = flag.String("tls-ca", "", "additional trusted CA certificate PEM path (enables TLS)")
	tlsName         = flag.String("tls-server-name", "", "TLS certificate server name (defaults to server host; enables TLS)")
	clientTLSConfig *tls.Config
)

func loadClientTLSConfig(caPath, serverName string) (*tls.Config, error) {
	config := &tls.Config{ServerName: serverName, MinVersion: tls.VersionTLS12}
	if caPath != "" {
		data, err := os.ReadFile(caPath)
		if err != nil {
			return nil, err
		}
		roots, err := x509.SystemCertPool()
		if err != nil {
			roots = x509.NewCertPool()
		}
		if !roots.AppendCertsFromPEM(data) {
			return nil, fmt.Errorf("TLS CA file contains no certificates")
		}
		config.RootCAs = roots
	}
	return config, nil
}

type Message struct {
	Id      uint16 `protobuf:"bytes,1,opt,name=msg_id,proto3" json:"msg_id"` // Message ID / 消息Id
	Data    []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data"`     // Message Content / 消息内容
	DataLen uint16 `json:"-"`                                                // Message Length / 消息长度
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
		msg.Data = data[4 : 4+msg.DataLen]
	}
	return msg
}

type TCPClient struct {
	conn net.Conn
}

func NewTCPClient(addr string) (*TCPClient, error) {
	var conn net.Conn
	var err error
	if clientTLSConfig != nil {
		conn, err = tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", addr, clientTLSConfig.Clone())
	} else {
		conn, err = net.DialTimeout("tcp", addr, 5*time.Second)
	}
	if err != nil {
		return nil, err
	}
	return &TCPClient{conn: conn}, nil
}

func (c *TCPClient) Write(msgId int16, data []byte) error {
	_, err := c.conn.Write(pack(msgId, data))
	if err != nil {
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
	dialer := *websocket.DefaultDialer
	if clientTLSConfig != nil {
		dialer.TLSClientConfig = clientTLSConfig.Clone()
		dialer.TLSClientConfig.NextProtos = []string{"http/1.1"}
	}
	conn, response, err := dialer.Dial(addr, nil)
	if response != nil {
		response.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	return &WSClient{conn: conn}, nil
}

func (c *WSClient) Write(msgId int16, data []byte) error {
	return c.conn.WriteMessage(websocket.BinaryMessage, pack(msgId, data))
}

func (c *WSClient) Read(buf []byte) (int, error) {
	msgType, message, err := c.conn.ReadMessage()
	if err != nil || msgType != websocket.BinaryMessage {
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
	conn net.Conn
}

func NewKCPClient(addr string) (*KCPClient, error) {
	conn, err := kcp.DialWithOptions(addr, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	conn.SetNoDelay(1, 10, 2, 1)
	conn.SetWindowSize(128, 128)
	conn.SetStreamMode(true)
	if clientTLSConfig != nil {
		config := clientTLSConfig.Clone()
		if config.ServerName == "" {
			config.ServerName, _, err = net.SplitHostPort(addr)
			if err != nil {
				conn.Close()
				return nil, err
			}
		}
		secure := tls.Client(conn, config)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := secure.HandshakeContext(ctx); err != nil {
			conn.Close()
			return nil, err
		}
		return &KCPClient{conn: secure}, nil
	}
	return &KCPClient{conn: conn}, nil
}

func (c *KCPClient) Write(msgId int16, data []byte) error {
	_, err := c.conn.Write(pack(msgId, data))
	if err != nil {
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
	client := &http.Client{Timeout: 5 * time.Second}
	if clientTLSConfig != nil {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = clientTLSConfig.Clone()
		client.Transport = transport
	}
	return &HTTPClient{url: addr, client: client, respBuf: make(chan string)}
}

func (c *HTTPClient) Write(msgId int16, data []byte) error {
	marshal, _ := json.Marshal(Message{Id: uint16(msgId), DataLen: uint16(len(data)), Data: data})
	resp, err := c.client.Post(c.url, "application/json", strings.NewReader(string(marshal)))
	if err != nil {
		return err
	}
	c.resp = resp
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
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
	if clientTLSConfig != nil {
		c.client.CloseIdleConnections()
	}
	return nil
}

type Client interface {
	Write(msgId int16, data []byte) error
	Read(buf []byte) (int, error)
	Close() error
}

func newClient(proto, addr string) (Client, error) {
	httpScheme, wsScheme := "http", "ws"
	if clientTLSConfig != nil {
		httpScheme, wsScheme = "https", "wss"
	}
	switch proto {
	case "tcp":
		return NewTCPClient(addr)
	case "ws":
		return NewWSClient(fmt.Sprintf("%s://%s", wsScheme, addr))
	case "kcp":
		return NewKCPClient(addr)
	case "http":
		return NewHTTPClient(fmt.Sprintf("%s://%s", httpScheme, addr)), nil
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
	if *useTLS || *tlsCA != "" || *tlsName != "" {
		var err error
		clientTLSConfig, err = loadClientTLSConfig(*tlsCA, *tlsName)
		if err != nil {
			slog.Error("invalid TLS configuration", "err", err)
			os.Exit(1)
		}
	}
	// Concurrently create connections, evenly distributed across protocols / 并发创建连接，多个协议时均匀分布
	var (
		sendCount = int32(0)
		recCount  = int32(0)
		wg        sync.WaitGroup
	)
	go func() {
		select {
		case <-time.After(30 * time.Second):
			slog.Error("测试超时", "received", atomic.LoadInt32(&recCount), "sent", atomic.LoadInt32(&sendCount))
		}
		os.Exit(1)
	}()

	// Parse and normalize protocol list, support multi-protocol mixing (e.g., "http,tcp,ws,kcp") / 解析并标准化协议列表，支持多协议混合（如 "http,tcp,ws,kcp"）
	protoList := strings.Split(*proto, ",")
	var protocols []string
	for _, p := range protoList {
		p = strings.TrimSpace(p)
		if p != "" {
			protocols = append(protocols, p)
		}
	}
	for i := 0; i < *connNum; i++ {
		wg.Add(1)
		go func(idx int) {
			protoc := protocols[idx%len(protocols)]
			c, err := newClient(protoc, getAddrByProto(protoc))
			if err != nil {
				slog.Error("connection failed", "index", idx, "err", err)
				wg.Done()
				return
			}

			go func(client Client) {
				defer wg.Done()
				buf := make([]byte, 4096)
				for {
					if n, err2 := client.Read(buf); err2 != nil {
						return // Connection closed or error / 连接关闭或出错
					} else if n > 0 {
						if d := unpack(buf[:n]); d != nil {
							atomic.AddInt32(&recCount, 1)
							// 不关闭连接，继续等待下一条消息
							_ = client.Close()
							return
						}
					}
					time.Sleep(time.Microsecond)
				}
			}(c)

			// Send Message periodically / 周期性发送消息
			// for {
			if err3 := c.Write(int16(*msgId), []byte(*msgData)); err3 != nil {
				slog.Error("write failed", "index", idx, "err", err3)
				_ = c.Close()
				return
			}
			atomic.AddInt32(&sendCount, 1)
			// time.Sleep(time.Second) // 每秒发送一次
			// }
		}(i)
	}
	wg.Wait()
	if atomic.LoadInt32(&recCount) != int32(*connNum) {
		slog.Error("incomplete responses", "received", atomic.LoadInt32(&recCount), "expected", *connNum)
		os.Exit(1)
	}
}
