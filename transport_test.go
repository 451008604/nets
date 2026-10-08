package nets

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xtaci/kcp-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func availablePort(t *testing.T, udp bool) int {
	t.Helper()
	if udp {
		conn, err := net.ListenPacket("udp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		return conn.LocalAddr().(*net.UDPAddr).Port
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func retry[T any](t *testing.T, operation func() (T, error)) T {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		value, err := operation()
		if err == nil {
			return value
		}
		if time.Now().After(deadline) {
			t.Fatal(err)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func startTestServer(t *testing.T, r *ServerManager, servers ...IServer) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- r.Run(ctx, servers...) }()
	t.Cleanup(func() {
		cancel()
		if err := await(t, result); err != nil {
			t.Error(err)
		}
	})
}

func TestProtocolsShareManagerAndShutdown(t *testing.T) {
	for _, jsonProtocol := range []bool{true, false} {
		t.Run(fmt.Sprintf("json=%v", jsonProtocol), func(t *testing.T) {
			configs := make(map[string]ServerConf)
			usedTCPPorts := make(map[int]bool)
			for _, protocol := range []string{"tcp", "kcp", "ws", "http"} {
				port := availablePort(t, protocol == "kcp")
				for protocol != "kcp" && usedTCPPorts[port] {
					port = availablePort(t, false)
				}
				if protocol != "kcp" {
					usedTCPPorts[port] = true
				}
				configs[protocol] = ServerConf{Address: "127.0.0.1", Port: port}
			}
			r := testServerManager(t, func(app *AppConf) {
				app.ProtocolIsJson = jsonProtocol
				app.ServerTCP, app.ServerKCP = configs["tcp"], configs["kcp"]
				app.ServerWS, app.ServerHTTP = configs["ws"], configs["http"]
			})
			// Register once: every protocol must use this manager's route table.
			r.MsgHandler().AddRouter(1001, func() proto.Message { return &wrapperspb.StringValue{} }, func(conn IConnection, msg proto.Message) { conn.SendMsg(1001, msg) })
			startTestServer(t, r, r.GetServerTCP(), r.GetServerKCP(), r.GetServerWS(), r.GetServerHTTP())
			for _, protocol := range []string{"tcp", "kcp", "ws", "http"} {
				t.Run(protocol, func(t *testing.T) {
					port := configs[protocol].Port
					codec := &ConnectionBase{manager: r}
					frame, err := codec.pack(1001, wrapperspb.String("echo 测试"))
					if err != nil {
						t.Fatal(err)
					}
					address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
					var response []byte
					switch protocol {
					case "tcp", "kcp":
						var conn net.Conn
						if protocol == "tcp" {
							conn = retry(t, func() (net.Conn, error) { return net.DialTimeout("tcp", address, 100*time.Millisecond) })
						} else {
							conn, err = kcp.Dial(address)
							if err != nil {
								t.Fatal(err)
							}
						}
						defer conn.Close()
						conn.SetDeadline(time.Now().Add(2 * time.Second))
						if !writeStream(conn, frame) {
							t.Fatal("write failed")
						}
						response = make([]byte, len(frame))
						if _, err := io.ReadFull(conn, response); err != nil {
							t.Fatal(err)
						}
					case "ws":
						conn := retry(t, func() (*websocket.Conn, error) {
							conn, response, err := websocket.DefaultDialer.Dial("ws://"+address, nil)
							if response != nil {
								response.Body.Close()
							}
							return conn, err
						})
						defer conn.Close()
						conn.SetReadDeadline(time.Now().Add(2 * time.Second))
						if err := conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
							t.Fatal(err)
						}
						_, response, err = conn.ReadMessage()
						if err != nil {
							t.Fatal(err)
						}
					case "http":
						payload, err := codec.marshal(wrapperspb.String("echo 测试"))
						if err != nil {
							t.Fatal(err)
						}
						body, err := codec.marshal(&Message{Id: 1001, Data: payload})
						if err != nil {
							t.Fatal(err)
						}
						client := &http.Client{Timeout: 2 * time.Second}
						result := retry(t, func() (*http.Response, error) {
							return client.Post("http://"+address, "application/octet-stream", bytes.NewReader(body))
						})
						defer result.Body.Close()
						if result.StatusCode != http.StatusOK {
							t.Fatalf("status %d", result.StatusCode)
						}
						response, err = io.ReadAll(result.Body)
						if err != nil {
							t.Fatal(err)
						}
					}
					if !bytes.Equal(response, frame) {
						t.Fatalf("response frame differs: %x != %x", response, frame)
					}
				})
			}
		})
	}
}

func TestHTTPAndStreamShareRouterAndLifecycle(t *testing.T) {
	for _, jsonProtocol := range []bool{true, false} {
		t.Run(fmt.Sprintf("json=%v", jsonProtocol), func(t *testing.T) {
			m := testServerManager(t, func(conf *AppConf) { conf.MaxConn = 2; conf.ProtocolIsJson = jsonProtocol })
			observed := make(chan string, 2)
			m.MsgHandler().AddSessionRouter(1001,
				func() proto.Message { return &wrapperspb.StringValue{} },
				func(conn Session, msg proto.Message) {
					observed <- msg.(*wrapperspb.StringValue).Value
					conn.SendMsg(1001, msg)
				})
			server, client := net.Pipe()
			defer client.Close()
			defer server.Close()
			client.SetDeadline(time.Now().Add(2 * time.Second))
			opened := make(chan struct{}, 1)
			m.ConnManager().SetConnOpened(func(IConnection) { opened <- struct{}{} })
			if !m.ConnManager().TryAdd(NewConnectionKCP(m.GetServerKCP(), server)) {
				t.Fatal("stream connection rejected")
			}
			await(t, opened)
			codec := &ConnectionBase{manager: m}
			frame, err := codec.pack(1001, wrapperspb.String("stream"))
			if err != nil {
				t.Fatal(err)
			}
			if !writeStream(client, frame) {
				t.Fatal("stream write failed")
			}
			response := make([]byte, len(frame))
			if _, err := io.ReadFull(client, response); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(response, frame) || await(t, observed) != "stream" {
				t.Fatal("stream did not use the shared route")
			}
			payload, err := codec.marshal(wrapperspb.String("http"))
			if err != nil {
				t.Fatal(err)
			}
			body, err := codec.marshal(&Message{Id: 1001, Data: payload})
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()
			m.GetServerHTTP().(*serverHTTP).handler().ServeHTTP(recorder, httptest.NewRequest("POST", "/", bytes.NewReader(body)))
			expected, err := codec.pack(1001, wrapperspb.String("http"))
			if err != nil {
				t.Fatal(err)
			}
			if recorder.Code != http.StatusOK || !bytes.Equal(recorder.Body.Bytes(), expected) || await(t, observed) != "http" {
				t.Fatal("HTTP did not use the shared route")
			}
			await(t, opened)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := m.Shutdown(ctx); err != nil {
				t.Fatal(err)
			}
			if m.ConnManager().Len() != 0 {
				t.Fatal("shutdown retained the stream connection")
			}
			if _, err := client.Read(make([]byte, 1)); err != io.EOF {
				t.Fatalf("shutdown did not close the stream: %v", err)
			}

		})
	}
}

func TestWebSocketRejectsOversizedFrame(t *testing.T) {
	r := testServerManager(t, func(conf *AppConf) { conf.MaxPackSize = 32 })
	server := httptest.NewServer(r.GetServerWS().(*serverWS).handler())
	defer server.Close()
	conn, response, err := websocket.DefaultDialer.Dial("ws"+server.URL[4:], nil)
	if response != nil {
		response.Body.Close()
	}
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(time.Second))
	if err := conn.WriteMessage(websocket.BinaryMessage, make([]byte, 1024)); err != nil {
		t.Fatal(err)
	}
	_, _, err = conn.ReadMessage()
	if !websocket.IsCloseError(err, websocket.CloseMessageTooBig) {
		t.Fatalf("expected size-limit close, got %v", err)
	}
}

func TestHTTPShutdownWaitsForCanceledHandler(t *testing.T) {
	port := availablePort(t, false)
	r := testServerManager(t, func(conf *AppConf) { conf.ServerHTTP = ServerConf{Address: "127.0.0.1", Port: port} })
	entered, canceled, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	r.messages.AddRouter(0, func() proto.Message { return &Message{} }, func(conn IConnection, _ proto.Message) {
		close(entered)
		<-conn.ConnCtx().Done()
		close(canceled)
		<-release
	})
	result := make(chan error, 1)
	go func() { result <- r.Run(context.Background(), r.GetServerHTTP()) }()
	response := make(chan error, 1)
	go func() {
		client := &http.Client{Timeout: 2 * time.Second}
		deadline := time.Now().Add(time.Second)
		for {
			res, err := client.Post(fmt.Sprintf("http://127.0.0.1:%d", port), "text/plain", bytes.NewBufferString("raw"))
			if err == nil {
				res.Body.Close()
				response <- nil
				return
			}
			if time.Now().After(deadline) {
				response <- err
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()
	await(t, entered)
	r.Stop()
	await(t, canceled)
	select {
	case <-result:
		t.Fatal("Run returned before HTTP handler completion")
	default:
	}
	close(release)
	if err := await(t, response); err != nil {
		t.Fatal(err)
	}
	if err := await(t, result); err != nil {
		t.Fatal(err)
	}
}
