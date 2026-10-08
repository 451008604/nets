package nets

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestConnectionTaskReentryRejectsFullQueue(t *testing.T) {
	m := testServerManager(t, func(conf *AppConf) { conf.WorkerPoolSize, conf.WorkerTaskMaxLen = 1, 1 })
	left, right := net.Pipe()
	t.Cleanup(func() { left.Close(); right.Close() })
	c := NewConnectionKCP(m.GetServerKCP(), left)
	t.Cleanup(c.Close)
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	t.Cleanup(func() { once.Do(func() { close(release) }) })
	result, queued := make(chan bool, 1), make(chan struct{})
	if !c.DoTask(func() { close(entered); <-release; result <- c.DoTask(func() { t.Error("rejected task executed") }) }) {
		t.Fatal("initial task rejected")
	}
	await(t, entered)
	if !c.DoTask(func() { close(queued) }) {
		t.Fatal("queue slot rejected")
	}
	once.Do(func() { close(release) })
	if await(t, result) {
		t.Fatal("full queue accepted a reentrant task")
	}
	await(t, queued)
	if c.DoTask(nil) {
		t.Fatal("nil task accepted")
	}
	c.Close()
	if c.DoTask(func() {}) {
		t.Fatal("closed connection accepted a task")
	}
}

func TestProtocolReadersRetainBackpressureAndOrder(t *testing.T) {
	for _, protocol := range []string{"tcp", "kcp", "ws"} {
		t.Run(protocol, func(t *testing.T) {
			m := testServerManager(t, func(conf *AppConf) { conf.WorkerPoolSize, conf.WorkerTaskMaxLen = 1, 1 })
			var c IConnection
			var write func([]byte) error
			switch protocol {
			case "tcp", "kcp":
				var server, client net.Conn
				if protocol == "tcp" {
					tcp, peer := tcpPair(t)
					server, client = tcp, peer
					c = NewConnectionTCP(m.GetServerTCP(), tcp)
				} else {
					server, client = net.Pipe()
					c = NewConnectionKCP(m.GetServerKCP(), server)
				}
				t.Cleanup(func() { server.Close(); client.Close() })
				write = func(data []byte) error { _, err := client.Write(data); return err }
			case "ws":
				accepted := make(chan *websocket.Conn, 1)
				upgrader := websocket.Upgrader{}
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					conn, err := upgrader.Upgrade(w, req, nil)
					if err == nil {
						accepted <- conn
					}
				}))
				t.Cleanup(server.Close)
				client, response, err := websocket.DefaultDialer.Dial("ws"+server.URL[4:], nil)
				if response != nil {
					response.Body.Close()
				}
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { client.Close() })
				c = NewConnectionWS(m.GetServerWS(), await(t, accepted))
				write = func(data []byte) error { return client.WriteMessage(websocket.BinaryMessage, data) }
			}
			t.Cleanup(c.Close)
			entered, release := make(chan struct{}), make(chan struct{})
			var once sync.Once
			t.Cleanup(func() { once.Do(func() { close(release) }) })
			observed := make(chan string, 3)
			m.messages.AddRouter(1, func() proto.Message { return &wrapperspb.StringValue{} }, func(_ IConnection, msg proto.Message) {
				value := msg.(*wrapperspb.StringValue).Value
				if value == "first" {
					close(entered)
					<-release
				}
				observed <- value
			})
			read := func(value string) <-chan bool {
				result := make(chan bool, 1)
				go func() { result <- c.StartReader() }()
				frame, err := (&ConnectionBase{manager: m}).pack(1, wrapperspb.String(value))
				if err != nil {
					t.Fatal(err)
				}
				if err := write(frame); err != nil {
					t.Fatal(err)
				}
				return result
			}
			if !await(t, read("first")) {
				t.Fatal("first packet rejected")
			}
			await(t, entered)
			if !await(t, read("second")) {
				t.Fatal("second packet rejected")
			}
			third := read("third")
			select {
			case <-third:
				t.Fatal("reader failed to apply queue backpressure")
			case <-time.After(25 * time.Millisecond):
			}
			once.Do(func() { close(release) })
			if !await(t, third) {
				t.Fatal("third packet rejected after queue drained")
			}
			for _, expected := range []string{"first", "second", "third"} {
				if actual := await(t, observed); actual != expected {
					t.Fatalf("got %q, want %q", actual, expected)
				}
			}
		})
	}
}

func TestConnectionCleanupWaitsForAcceptedTasks(t *testing.T) {
	for _, panicTask := range []bool{false, true} {
		t.Run(map[bool]string{false: "normal", true: "panic"}[panicTask], func(t *testing.T) {
			m := testServerManager(t, func(conf *AppConf) { conf.WorkerPoolSize, conf.WorkerTaskMaxLen = 1, 1 })
			left, right := net.Pipe()
			t.Cleanup(func() { left.Close(); right.Close() })
			c := NewConnectionKCP(m.GetServerKCP(), left)
			opened, closed := make(chan struct{}), make(chan any, 1)
			m.connections.SetConnOpened(func(conn IConnection) { conn.SetProperty("session", "valid"); close(opened) })
			m.connections.SetConnClosed(func(conn IConnection) { closed <- conn.GetProperty("session") })
			if !m.connections.TryAdd(c) {
				t.Fatal("connection rejected")
			}
			await(t, opened)
			entered, closeNow, closeReturned, release := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
			var once sync.Once
			t.Cleanup(func() { once.Do(func() { close(release) }) })
			property, reclaimed := make(chan any, 1), make(chan struct{})
			if !c.DoTask(func() {
				close(entered)
				<-closeNow
				c.Close() // Must not wait for this task itself.
				close(closeReturned)
				<-release
				property <- c.GetProperty("session")
				if panicTask {
					panic("task failed during shutdown")
				}
			}) {
				t.Fatal("task rejected")
			}
			await(t, entered)
			m.messages.AddRouter(1, func() proto.Message { return &wrapperspb.StringValue{} }, func(IConnection, proto.Message) { t.Error("canceled packet executed business handler") })
			message := GetMessage()
			message.Id = 1
			if !submitReadTask(c, func() { defer close(reclaimed); defer PutMessage(message); readerTaskHandler(c, message) }) {
				PutMessage(message)
				t.Fatal("packet rejected")
			}
			close(closeNow)
			await(t, closeReturned)
			waited := make(chan struct{})
			go func() { m.connections.Wait(); close(waited) }()
			select {
			case <-closed:
				t.Fatal("closed hook ran before running task completed")
			case <-waited:
				t.Fatal("connection wait completed before running task")
			case <-time.After(25 * time.Millisecond):
			}
			once.Do(func() { close(release) })
			if await(t, property) != "valid" {
				t.Fatal("session property cleared while task was running")
			}
			await(t, reclaimed)
			if await(t, closed) != "valid" {
				t.Fatal("closed hook lost the session resource")
			}
			await(t, waited)
			if c.GetProperty("session") != nil {
				t.Fatal("session property not cleared after cleanup")
			}
		})
	}
}

type delayedRegistration struct {
	IConnection
	id      string
	release chan struct{}
}

func (c *delayedRegistration) GetConnId() string { return c.id }
func (c *delayedRegistration) Close()            {}
func (c *delayedRegistration) Open()             { <-c.release }

func TestRegistrationTokenProtectsReplacement(t *testing.T) {
	m := testServerManager(t, func(conf *AppConf) { conf.MaxConn = 1 })
	old := &delayedRegistration{id: "fixed", release: make(chan struct{})}
	replacement := &delayedRegistration{id: "fixed", release: make(chan struct{})}
	t.Cleanup(func() { close(old.release); close(replacement.release) })
	if !m.connections.TryAdd(old) {
		t.Fatal("old connection rejected")
	}
	oldEntry, _ := m.connections.connections.Get("fixed")
	m.connections.Remove(old)
	if !m.connections.TryAdd(replacement) {
		t.Fatal("replacement rejected")
	}
	// Exercise the exact token captured by old.Open's deferred cleanup, without
	// depending on goroutine scheduling to observe whether it removed the new one.
	m.connections.removeEntry(oldEntry)
	m.connections.Remove(old)
	if actual, ok := m.connections.Get("fixed"); !ok || actual != replacement {
		t.Fatal("stale cleanup removed replacement")
	}
	var ranged []IConnection
	m.connections.RangeConnections(func(c IConnection) { ranged = append(ranged, c) })
	if len(ranged) != 1 || ranged[0] != replacement || m.connections.Len() != 1 {
		t.Fatal("replacement missing from registry")
	}
	if m.connections.acquireRequest() {
		m.connections.releaseRequest()
		t.Fatal("stale cleanup released replacement capacity")
	}
	m.connections.Remove(replacement)
	m.connections.Remove(replacement)
	if !m.connections.acquireRequest() {
		t.Fatal("replacement did not release capacity exactly once")
	}
	m.connections.releaseRequest()
}

type uncomparableConnection struct {
	IConnection
	id   string
	data []byte
	done chan struct{}
	once *sync.Once
}

func (c uncomparableConnection) GetConnId() string { return c.id }
func (c uncomparableConnection) Open()             { <-c.done }
func (c uncomparableConnection) Close()            { c.once.Do(func() { close(c.done) }) }

func TestRemoveSupportsUncomparableConnection(t *testing.T) {
	for _, nested := range []bool{false, true} {
		t.Run(map[bool]string{false: "slice", true: "nested-interface"}[nested], func(t *testing.T) {
			m := testServerManager(t, nil)
			var c IConnection = uncomparableConnection{id: "custom", data: []byte{1}, done: make(chan struct{}), once: &sync.Once{}}
			if nested {
				// The outer type is comparable, but its interface value contains a slice.
				c = struct{ IConnection }{c}
			}
			if !m.connections.TryAdd(c) {
				t.Fatal("custom connection rejected")
			}
			m.connections.Remove(c)
			m.connections.Remove(c)
			done := make(chan struct{})
			go func() { m.connections.Wait(); close(done) }()
			await(t, done)
			if m.connections.Len() != 0 {
				t.Fatal("custom registration retained after Open returned")
			}
		})
	}
}

func TestTaskAdmissionConcurrentWithClose(t *testing.T) {
	m := testServerManager(t, nil)
	left, right := net.Pipe()
	t.Cleanup(func() { left.Close(); right.Close() })
	c := NewConnectionKCP(m.GetServerKCP(), left)
	opened := make(chan struct{})
	m.connections.SetConnOpened(func(IConnection) { close(opened) })
	if !m.connections.TryAdd(c) {
		t.Fatal("connection rejected")
	}
	await(t, opened)
	start := make(chan struct{})
	var submitters sync.WaitGroup
	for i := 0; i < 64; i++ {
		submitters.Add(1)
		go func() {
			defer submitters.Done()
			<-start
			c.DoTask(func() {})
		}()
	}
	close(start)
	c.Close()
	submitters.Wait()
	done := make(chan struct{})
	go func() { m.connections.Wait(); close(done) }()
	await(t, done)
	if c.DoTask(func() {}) {
		t.Fatal("task admitted after cleanup completed")
	}
}

func TestTLSConfigurationMatchesProtocolSupport(t *testing.T) {
	for _, name := range []string{"TCP", "KCP", "HTTP", "WebSocket"} {
		for _, paths := range [][2]string{{"cert", ""}, {"", "key"}, {"cert", "key"}} {
			t.Run(name+"/"+paths[0]+"/"+paths[1], func(t *testing.T) {
				conf := GetServerConf()
				protocol := map[string]*ServerConf{"TCP": &conf.ServerTCP, "KCP": &conf.ServerKCP, "HTTP": &conf.ServerHTTP, "WebSocket": &conf.ServerWS}[name]
				protocol.Port, protocol.TLSCertPath, protocol.TLSKeyPath = 0, paths[0], paths[1]
				manager, err := NewServerManager(&CustomServer{AppConf: conf})
				valid := paths[0] != "" && paths[1] != ""
				if valid {
					if err != nil {
						t.Fatal(err)
					}
					if err := manager.Shutdown(context.Background()); err != nil {
						t.Fatal(err)
					}
				} else if err == nil || !strings.Contains(err.Error(), name) {
					t.Fatalf("expected %s validation error, got %v", name, err)
				}
			})
		}
	}
}
