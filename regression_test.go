package nets

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func testServerManager(t *testing.T, change func(*AppConf)) *ServerManager {
	t.Helper()
	conf := GetServerConf()
	conf.WorkerPoolSize, conf.WorkerTaskMaxLen, conf.MaxMsgChanLen = 2, 4, 4
	if change != nil {
		change(&conf)
	}
	r, err := NewServerManager(&CustomServer{AppConf: conf, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := r.Shutdown(ctx); err != nil {
			t.Error(err)
		}
	})
	return r
}

func await[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(2 * time.Second):
		t.Fatal("operation timed out")
		var zero T
		return zero
	}
}

func TestMessageProtobufRoundTrip(t *testing.T) {
	var id uint16 = 1001
	message := &Message{Id: id, Data: []byte("payload"), DataLen: 99}
	wire, err := proto.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	decoded := &Message{}
	if err := proto.Unmarshal(wire, decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Id != id || !bytes.Equal(decoded.Data, message.Data) || decoded.DataLen != 0 {
		t.Fatalf("unexpected message: %#v", decoded)
	}
	clone, ok := proto.Clone(message).(*Message)
	if !ok || !proto.Equal(message, clone) {
		t.Fatal("clone lost the concrete message type or content")
	}
	clone.Data[0] = 'X'
	if bytes.Equal(message.Data, clone.Data) {
		t.Fatal("clone aliases source payload")
	}
	reflection := message.ProtoReflect()
	message.Id = 42
	if reflection.Get(messageDescriptor.Fields().ByNumber(1)).Uint() != 42 {
		t.Fatal("reflection reads a stale snapshot")
	}
	unknown := protowire.AppendTag(nil, 10, protowire.VarintType)
	unknown = protowire.AppendVarint(unknown, 7)
	wire = append(wire, unknown...)
	if err := proto.Unmarshal(wire, decoded); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decoded.ProtoReflect().GetUnknown(), unknown) {
		t.Fatal("unknown fields were lost")
	}
	proto.Reset(decoded)
	if decoded.Id != 0 || decoded.Data != nil || len(decoded.ProtoReflect().GetUnknown()) != 0 {
		t.Fatal("reset retained message data")
	}
}

func TestHTTPJSONAndProtobufIsolation(t *testing.T) {
	var group sync.WaitGroup
	for _, jsonProtocol := range []bool{true, false} {
		r := testServerManager(t, func(conf *AppConf) { conf.ProtocolIsJson = jsonProtocol })
		r.messages.AddSessionRouter(1001, func() proto.Message { return &wrapperspb.StringValue{} }, func(conn Session, msg proto.Message) {
			conn.SendMsg(1001, msg)
		})
		server := r.GetServerHTTP().(*serverHTTP)
		group.Add(1)
		go func() {
			defer group.Done()
			for i := 0; i < 30; i++ {
				value := fmt.Sprintf("protocol=%v-request=%d", jsonProtocol, i)
				codec := &ConnectionBase{manager: r}
				payload, err := codec.marshal(wrapperspb.String(value))
				if err != nil {
					t.Error(err)
					return
				}
				body, err := codec.marshal(&Message{Id: 1001, Data: payload})
				if err != nil {
					t.Error(err)
					return
				}
				response := httptest.NewRecorder()
				server.handler().ServeHTTP(response, httptest.NewRequest("POST", "/", bytes.NewReader(body)))
				if response.Code != http.StatusOK {
					t.Errorf("status %d: %s", response.Code, response.Body)
					return
				}
				envelope := r.config.DataPack.UnPack(response.Body.Bytes())
				if envelope == nil {
					t.Error("invalid response frame")
					return
				}
				decoded := &wrapperspb.StringValue{}
				err = codec.ByteToProtocol(envelope.GetData(), decoded)
				PutMessage(envelope)
				if err != nil || decoded.Value != value {
					t.Errorf("response corrupted: %q, %v", decoded.Value, err)
					return
				}
			}
		}()
	}
	group.Wait()
}

func TestHTTPMessageIsReleasedOnce(t *testing.T) {
	r := testServerManager(t, nil)
	r.messages.AddRouter(0, func() proto.Message { return &Message{} }, func(IConnection, proto.Message) {})
	server := r.GetServerHTTP()
	var opened, closed atomic.Int32
	r.connections.SetConnOpened(func(IConnection) { opened.Add(1) })
	r.connections.SetConnClosed(func(conn IConnection) {
		if !conn.IsClose() {
			t.Error("closed hook saw a live connection")
		}
		closed.Add(1)
	})
	for i := 0; i < 20; i++ {
		conn := NewConnectionHTTP(server, httptest.NewRecorder(), httptest.NewRequest("POST", "/", bytes.NewBufferString("raw body")))
		if !conn.StartReader() {
			t.Fatal("raw request rejected")
		}
		borrowed := make([]*Message, 16)
		seen := make(map[*Message]bool)
		for j := range borrowed {
			borrowed[j] = GetMessage()
			if seen[borrowed[j]] {
				t.Fatal("object pool lent one Message to two borrowers")
			}
			seen[borrowed[j]] = true
		}
		for _, msg := range borrowed {
			PutMessage(msg)
		}
	}
	if opened.Load() != 20 || closed.Load() != 20 {
		t.Fatalf("hook counts %d/%d", opened.Load(), closed.Load())
	}
}

func TestHTTPRequestContextAndInvalidEnvelope(t *testing.T) {
	r := testServerManager(t, func(conf *AppConf) { conf.ProtocolIsJson = false })
	ctx, cancel := context.WithCancel(context.Background())
	conn := NewConnectionHTTP(r.GetServerHTTP(), httptest.NewRecorder(), httptest.NewRequest("POST", "/", nil).WithContext(ctx))
	cancel()
	if conn.ConnCtx().Err() != context.Canceled || !conn.IsClose() {
		t.Fatal("HTTP cancellation was not propagated")
	}
	wire := protowire.AppendTag(nil, 1, protowire.VarintType)
	wire = protowire.AppendVarint(wire, 65536)
	if err := conn.ByteToProtocol(wire, &Message{}); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("out-of-range wire ID accepted: %v", err)
	}
}

func tcpPair(t *testing.T) (*net.TCPConn, net.Conn) {
	t.Helper()
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	client, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	server, err := listener.AcceptTCP()
	if err != nil {
		client.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close(); server.Close() })
	server.SetDeadline(time.Now().Add(2 * time.Second))
	client.SetDeadline(time.Now().Add(2 * time.Second))
	return server, client
}

func TestTCPFragmentedHeaderAndOrderedMessages(t *testing.T) {
	r := testServerManager(t, nil)
	server, client := tcpPair(t)
	conn := NewConnectionTCP(r.GetServerTCP(), server)
	observed := make(chan string, 2)
	r.messages.AddRouter(1001, func() proto.Message { return &wrapperspb.StringValue{} }, func(_ IConnection, msg proto.Message) { observed <- msg.(*wrapperspb.StringValue).Value })
	first, err := conn.(*connectionTCP).pack(1001, wrapperspb.String("first"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := conn.(*connectionTCP).pack(1001, wrapperspb.String("second"))
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan bool, 1)
	go func() { result <- conn.StartReader() }()
	if _, err := client.Write(first[:1]); err != nil {
		t.Fatal(err)
	}
	select {
	case <-result:
		t.Fatal("reader rejected an incomplete header")
	case <-time.After(25 * time.Millisecond):
	}
	if _, err := client.Write(append(first[1:], second...)); err != nil {
		t.Fatal(err)
	}
	if !await(t, result) || !conn.StartReader() {
		t.Fatal("valid packet rejected")
	}
	if await(t, observed) != "first" || await(t, observed) != "second" {
		t.Fatal("connection message order changed")
	}
}

type chunkedConn struct{ net.Conn }

func (c chunkedConn) Read(data []byte) (int, error) {
	if len(data) > 1 {
		data = data[:1]
	}
	return c.Conn.Read(data)
}
func (c chunkedConn) Write(data []byte) (int, error) {
	if len(data) > 2 {
		data = data[:2]
	}
	return c.Conn.Write(data)
}

func TestKCPStreamShortReadsAndWrites(t *testing.T) {
	r := testServerManager(t, nil)
	left, right := net.Pipe()
	defer left.Close()
	defer right.Close()
	left.SetDeadline(time.Now().Add(2 * time.Second))
	right.SetDeadline(time.Now().Add(2 * time.Second))
	conn := NewConnectionKCP(r.GetServerKCP().(*serverKCP), chunkedConn{left})
	frame, err := conn.(*connectionKCP).pack(1, wrapperspb.String("payload"))
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan bool, 1)
	go func() { result <- conn.StartReader() }()
	if _, err := right.Write(frame); err != nil {
		t.Fatal(err)
	}
	if !await(t, result) {
		t.Fatal("short reads rejected")
	}
	go func() { result <- conn.StartWriter(frame) }()
	output := make([]byte, len(frame))
	if _, err := io.ReadFull(right, output); err != nil {
		t.Fatal(err)
	}
	if !await(t, result) || !bytes.Equal(output, frame) {
		t.Fatal("short writes lost bytes")
	}
}

func TestSendErrorsAndCanceledSubmission(t *testing.T) {
	r := testServerManager(t, nil)
	server, _ := tcpPair(t)
	conn := NewConnectionTCP(r.GetServerTCP(), server)
	sender := conn.(ContextSender)
	if err := sender.SendMsgWithContext(context.Background(), -1, wrapperspb.String("x")); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("invalid ID: %v", err)
	}
	if err := sender.SendMsgWithContext(context.Background(), 1, wrapperspb.String(string(make([]byte, 5000)))); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("oversized payload: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sender.SendMsgWithContext(ctx, 1, wrapperspb.String("x")); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled send: %v", err)
	}
	conn.Close()
	if conn.DoTask(func() { t.Error("closed connection accepted task") }) {
		t.Fatal("canceled submission reported success")
	}
	if err := sender.SendMsgWithContext(context.Background(), 1, wrapperspb.String("x")); !errors.Is(err, ErrConnectionClosed) {
		t.Fatalf("closed send: %v", err)
	}
}

func TestWorkerPanicDrainAndRejection(t *testing.T) {
	pool := NewWorkerPool(1, 8)
	defer pool.Stop()
	started, release := make(chan struct{}), make(chan struct{})
	pool.Submit(func() { close(started); <-release; panic("task failure") })
	await(t, started)
	var completed atomic.Int32
	for i := 0; i < 6; i++ {
		if err := pool.Submit(func() { completed.Add(1) }); err != nil {
			t.Fatal(err)
		}
	}
	stopped := make(chan struct{})
	go func() { pool.Stop(); close(stopped) }()
	await(t, pool.done)
	if err := pool.Submit(func() {}); !errors.Is(err, ErrPoolClosed) {
		t.Fatalf("post-stop submission: %v", err)
	}
	close(release)
	await(t, stopped)
	if completed.Load() != 6 || pool.Pending() != 0 {
		t.Fatalf("accepted tasks lost: %d, pending %d", completed.Load(), pool.Pending())
	}
}

func TestWorkerBlockingSubmissionCancellation(t *testing.T) {
	pool := NewWorkerPool(1, 1)
	defer pool.Stop()
	started, release := make(chan struct{}), make(chan struct{})
	defer close(release)
	pool.Submit(func() { close(started); <-release })
	await(t, started)
	pool.Submit(func() {})
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- pool.SubmitCtx(ctx, func() { t.Error("canceled task ran") }) }()
	cancel()
	if err := await(t, result); !errors.Is(err, context.Canceled) {
		t.Fatalf("submission error: %v", err)
	}
}

func TestRateLimitHookCanReenter(t *testing.T) {
	r := testServerManager(t, func(conf *AppConf) { conf.MaxFlowSecond = 0 })
	server, _ := tcpPair(t)
	conn := NewConnectionTCP(r.GetServerTCP(), server)
	r.connections.SetConnOnRateLimiting(func(conn IConnection) { r.connections.SetConnOnRateLimiting(nil); conn.FlowControl() })
	result := make(chan bool, 1)
	go func() { result <- conn.FlowControl() }()
	if !await(t, result) {
		t.Fatal("zero-QPS limit allowed a message")
	}
}

func TestConcurrentRouterRegistrationAndLookup(t *testing.T) {
	r := testServerManager(t, nil)
	var workers sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		workers.Add(1)
		go func(worker int) {
			defer workers.Done()
			for i := 0; i < 50; i++ {
				id := int32(worker*50 + i + 1)
				r.messages.AddRouter(id, func() proto.Message { return &wrapperspb.StringValue{} }, func(IConnection, proto.Message) {})
				if _, ok := r.messages.GetRouter(id); !ok {
					t.Error("published router missing")
				}
			}
		}(worker)
	}
	workers.Wait()
}

func TestDataPackReportsMalformedAndOversizedPackets(t *testing.T) {
	pack := NewDataPack().(ErrorDataPack)
	if _, err := pack.UnPackWithError([]byte{1}); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("short header: %v", err)
	}
	if _, err := pack.UnPackWithError([]byte{1, 0, 0xff, 0xff}); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("oversized declaration: %v", err)
	}
	if _, err := pack.PackWithError(&Message{Id: 1, Data: make([]byte, 65536)}); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("oversized payload: %v", err)
	}
}

func TestClosedBeforeOpenDoesNotCallHooks(t *testing.T) {
	r := testServerManager(t, nil)
	server, _ := tcpPair(t)
	conn := NewConnectionTCP(r.GetServerTCP(), server)
	r.connections.SetConnOpened(func(IConnection) { t.Error("closed connection invoked opened hook") })
	r.connections.SetConnClosed(func(IConnection) { t.Error("unopened connection invoked closed hook") })
	conn.Close()
	conn.Open()
}

func TestSendBackpressureHonorsContext(t *testing.T) {
	r := testServerManager(t, func(conf *AppConf) { conf.MaxMsgChanLen = 1 })
	server, _ := tcpPair(t)
	conn := NewConnectionTCP(r.GetServerTCP(), server).(ContextSender)
	if err := conn.SendMsgWithContext(context.Background(), 1, wrapperspb.String("queued")); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := conn.SendMsgWithContext(ctx, 1, wrapperspb.String("blocked")); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("backpressure ignored deadline: %v", err)
	}
}

func TestRawProtobufFallbackClearsEnvelopeMetadata(t *testing.T) {
	r := testServerManager(t, func(conf *AppConf) { conf.ProtocolIsJson = false })
	wire := protowire.AppendTag(nil, 1, protowire.VarintType)
	wire = protowire.AppendVarint(wire, 65536)
	called := false
	r.messages.AddRouter(0, func() proto.Message { return &Message{} }, func(_ IConnection, msg proto.Message) {
		called = true
		if msg.ProtoReflect().Get(messageDescriptor.Fields().ByNumber(1)).Uint() != 0 {
			t.Error("raw fallback retained decoded message ID")
		}
		if !bytes.Equal(msg.(*Message).Data, wire) {
			t.Error("raw fallback lost request body")
		}
	})
	response := httptest.NewRecorder()
	r.GetServerHTTP().(*serverHTTP).handler().ServeHTTP(response, httptest.NewRequest("POST", "/", bytes.NewReader(wire)))
	if response.Code != http.StatusOK || !called {
		t.Fatal("explicit raw route was not invoked")
	}
}
