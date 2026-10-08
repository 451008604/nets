package nets

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type cleanupServer struct {
	started, stopping, release, stopped chan struct{}
}

func (s *cleanupServer) GetServerName() string { return "cleanup" }
func (s *cleanupServer) Start()                {}
func (s *cleanupServer) Serve(ctx context.Context) error {
	close(s.started)
	<-ctx.Done()
	close(s.stopping)
	<-s.release
	close(s.stopped)
	return nil
}

func TestServerManagerWaitsForServerCleanup(t *testing.T) {
	r := testServerManager(t, nil)
	server := &cleanupServer{make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})}
	result := make(chan error, 1)
	go func() { result <- r.Run(context.Background(), server) }()
	await(t, server.started)
	r.Stop()
	await(t, server.stopping)
	select {
	case <-result:
		t.Fatal("Run returned before service cleanup")
	default:
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := r.Shutdown(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("shutdown did not honor context: %v", err)
	}
	close(server.release)
	if err := await(t, result); err != nil {
		t.Fatal(err)
	}
	await(t, server.stopped)
	if err := r.Run(context.Background(), server); !errors.Is(err, ErrServerManagerClosed) {
		t.Fatalf("stopped manager restarted: %v", err)
	}
}

func TestBindFailureCancelsSiblingServices(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	_, portText, _ := net.SplitHostPort(listener.Addr().String())
	port, _ := strconv.Atoi(portText)
	r := testServerManager(t, func(conf *AppConf) { conf.ServerTCP = ServerConf{Address: "127.0.0.1", Port: port} })
	peer := &cleanupServer{make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})}
	close(peer.release)
	result := make(chan error, 1)
	go func() { result <- r.Run(context.Background(), peer, r.GetServerTCP()) }()
	if err := await(t, result); err == nil {
		t.Fatal("bind failure was swallowed")
	}
	await(t, peer.stopped)
}

type heldConnection struct {
	IConnection
	id   string
	done chan struct{}
	once sync.Once
}

func (c *heldConnection) GetConnId() string { return c.id }
func (c *heldConnection) Open()             { <-c.done }
func (c *heldConnection) Close()            { c.once.Do(func() { close(c.done) }) }

func TestSharedAdmissionLimitAndShutdown(t *testing.T) {
	r := testServerManager(t, func(conf *AppConf) { conf.MaxConn = 16 })
	var admitted atomic.Int32
	var attempts, requests sync.WaitGroup
	releaseRequests := make(chan struct{})
	for i := 0; i < 100; i++ {
		attempts.Add(1)
		if i%2 == 0 {
			requests.Add(1)
			go func() {
				defer requests.Done()
				accepted := r.connections.acquireRequest()
				if accepted {
					admitted.Add(1)
				}
				attempts.Done()
				if accepted {
					<-releaseRequests
					r.connections.releaseRequest()
				}
			}()
		} else {
			go func(id int) {
				defer attempts.Done()
				if r.connections.TryAdd(&heldConnection{id: strconv.Itoa(id), done: make(chan struct{})}) {
					admitted.Add(1)
				}
			}(i)
		}
	}
	attempts.Wait()
	if admitted.Load() != 16 {
		t.Fatalf("shared budget admitted %d, expected 16", admitted.Load())
	}
	r.Stop()
	if r.connections.acquireRequest() {
		t.Fatal("shutdown admitted a new request")
	}
	if r.connections.TryAdd(&heldConnection{id: "late", done: make(chan struct{})}) {
		t.Fatal("shutdown admitted a new connection")
	}
	close(releaseRequests)
	requests.Wait()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := r.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if r.connections.Len() != 0 {
		t.Fatal("shutdown retained connections")
	}
	r.connections.mu.Lock()
	active := r.connections.active
	r.connections.mu.Unlock()
	if active != 0 {
		t.Fatalf("shutdown retained %d capacity reservations", active)
	}
}

func TestHTTPAndLongConnectionUseSameCapacity(t *testing.T) {
	r := testServerManager(t, func(conf *AppConf) { conf.MaxConn = 1 })
	conn := &heldConnection{id: "long", done: make(chan struct{})}
	if !r.connections.TryAdd(conn) {
		t.Fatal("connection rejected")
	}
	response := httptest.NewRecorder()
	r.GetServerHTTP().(*serverHTTP).handler().ServeHTTP(response, httptest.NewRequest("POST", "/", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("HTTP exceeded shared limit: %d", response.Code)
	}
	r.connections.Remove(conn)
	if !r.connections.acquireRequest() {
		t.Fatal("connection did not release capacity")
	}
	if r.connections.TryAdd(&heldConnection{id: "another", done: make(chan struct{})}) {
		t.Fatal("long connection exceeded HTTP reservation")
	}
	r.connections.releaseRequest()
}

func TestServerManagerRejectsForeignServerAndInvalidConfiguration(t *testing.T) {
	a, b := testServerManager(t, nil), testServerManager(t, nil)
	if err := a.Run(context.Background(), b.GetServerHTTP()); err == nil {
		t.Fatal("foreign service accepted")
	}
	for _, change := range []func(*AppConf){
		func(conf *AppConf) { conf.MaxPackSize = 65536 },
		func(conf *AppConf) { conf.ServerTCP.Port = -1 },
		func(conf *AppConf) { conf.ServerHTTP.TLSCertPath = "cert-only" },
	} {
		conf := GetServerConf()
		change(&conf)
		if _, err := NewServerManager(&CustomServer{AppConf: conf}); err == nil {
			t.Fatal("invalid configuration accepted")
		}
	}
}

func TestStandaloneServerManagerStopsWorkerPool(t *testing.T) {
	r := testServerManager(t, nil)
	pool := r.WorkerPool()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := r.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pool.Submit(func() {}); !errors.Is(err, ErrPoolClosed) {
		t.Fatalf("pool still accepting tasks: %v", err)
	}
}

func TestLegacyGlobalAPILifecycleAndConfigurationFreeze(t *testing.T) {
	defaultServerManagerMu.Lock()
	originalConfig, originalManager := *defaultServer, defaultServerManager
	defaultServerManager = nil
	defaultServerManagerMu.Unlock()
	t.Cleanup(func() {
		defaultServerManagerMu.Lock()
		*defaultServer, defaultServerManager = originalConfig, originalManager
		defaultServerManagerMu.Unlock()
	})
	conf := GetServerConf()
	conf.WorkerPoolSize = 1
	if err := ConfigureDefault(&CustomServer{AppConf: conf}); err != nil {
		t.Fatal(err)
	}
	manager := GetInstanceServerManager()
	if GetInstanceMsgHandler() != manager.MsgHandler() || GetInstanceConnManager() != manager.ConnManager() || GetInstanceWorkerPool() != manager.WorkerPool() {
		t.Fatal("global facade uses different instances")
	}
	if err := ConfigureDefault(&CustomServer{AppConf: conf}); !errors.Is(err, ErrServerManagerStarted) {
		t.Fatalf("late configuration accepted: %v", err)
	}
	peer := &cleanupServer{make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})}
	close(peer.release)
	done := make(chan struct{})
	go func() { manager.RegisterServer(peer); close(done) }()
	await(t, peer.started)
	manager.StopAll()
	await(t, done)
	await(t, manager.done)
}
