package nets

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
)

var (
	ErrServerManagerClosed  = errors.New("server manager is closed")
	ErrServerManagerStarted = errors.New("server manager has already started")
	ErrCapacity             = errors.New("connection capacity exhausted")
	ErrInvalidMessage       = errors.New("invalid message")
	ErrConnectionClosed     = errors.New("connection is closed")
)

// ManagedServer extends IServer with cancellation and error reporting.
// Serve must wait for its resources and shutdown work before returning.
type ManagedServer interface {
	IServer
	Serve(context.Context) error
}

// ServerManager owns the shared configuration, routing, connections, workers and
// lifecycle for all registered protocols. Construct a new manager after stopping.
type ServerManager struct {
	config       CustomServer
	connections  *ConnectionManager
	messages     *MsgHandler
	poolOnce     sync.Once
	pool         *WorkerPool
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.Mutex
	started      bool
	isClosed     int32
	waitGroup    sync.WaitGroup
	shutdownOnce sync.Once
	done         chan struct{}
}

// NewServerManager validates and copies configuration for all of its protocols.
// A nil configuration uses GetServerConf defaults.
func NewServerManager(custom *CustomServer) (*ServerManager, error) {
	config := CustomServer{AppConf: GetServerConf()}
	if custom != nil {
		config = *custom
	}
	if config.AppConf.MaxPackSize == 0 {
		config.AppConf.MaxPackSize = 4096
	}
	if config.AppConf.MaxPackSize > 65535 {
		return nil, fmt.Errorf("MaxPackSize exceeds uint16 wire limit")
	}
	if config.AppConf.ConnRWTimeOut == 0 {
		config.AppConf.ConnRWTimeOut = 30
	}
	if config.AppConf.MaxFlowSecond < -1 {
		return nil, fmt.Errorf("MaxFlowSecond must be -1 or nonnegative")
	}
	for _, protocol := range []struct {
		name string
		conf ServerConf
	}{{"TCP", config.AppConf.ServerTCP}, {"KCP", config.AppConf.ServerKCP}, {"WebSocket", config.AppConf.ServerWS}, {"HTTP", config.AppConf.ServerHTTP}} {
		conf := protocol.conf
		if conf.Port < 0 || conf.Port > 65535 {
			return nil, fmt.Errorf("%s port outside [0, 65535]", protocol.name)
		}
		if (conf.TLSCertPath == "") != (conf.TLSKeyPath == "") {
			return nil, fmt.Errorf("%s TLS certificate and key must both be configured", protocol.name)
		}
	}
	if config.Logger == nil {
		config.Logger = newDefaultLogger(config.AppConf.LogLevel)
	}
	if config.DataPack == nil {
		config.DataPack = NewDataPack()
	}
	if _, ok := config.DataPack.(*dataPack); ok {
		config.DataPack = &dataPack{maxPackSize: config.AppConf.MaxPackSize}
	}
	if config.DataPack.GetHeadLen() <= 0 || config.DataPack.GetHeadLen() > 65535 {
		return nil, fmt.Errorf("invalid packet header size")
	}
	ctx, cancel := context.WithCancel(context.Background())
	m := &ServerManager{config: config, messages: NewMsgHandler(), ctx: ctx, cancel: cancel, done: make(chan struct{})}
	m.messages.logger = config.Logger
	m.connections = newConnectionManager(m)
	return m, nil
}

func (m *ServerManager) Config() AppConf                 { return m.config.AppConf }
func (m *ServerManager) MsgHandler() *MsgHandler         { return m.messages }
func (m *ServerManager) ConnManager() *ConnectionManager { return m.connections }
func (m *ServerManager) WorkerPool() *WorkerPool {
	m.poolOnce.Do(func() {
		m.pool = newWorkerPool(m.config.AppConf.WorkerPoolSize, m.config.AppConf.WorkerTaskMaxLen, m.messages)
		if m.IsClose() {
			m.pool.Stop()
		}
	})
	return m.pool
}

func GetInstanceServerManager() *ServerManager { return getDefaultServerManager() }

// RegisterServer preserves the blocking, signal-aware legacy entry point.
// New applications should use ServerManager.Run with their own context.
func (m *ServerManager) RegisterServer(servers ...IServer) {
	if len(servers) == 0 {
		return
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer cancel()
	if err := m.Run(ctx, servers...); err != nil {
		m.config.Logger.Error("server stopped", "err", err)
	}
}

// Run waits for services, connections and accepted tasks to stop.
func (m *ServerManager) Run(ctx context.Context, servers ...IServer) error {
	if ctx == nil {
		return errors.New("nil run context")
	}
	for _, server := range servers {
		if server == nil {
			return errors.New("nil server")
		}
		if owner, ok := server.(interface{ managerOwner() *ServerManager }); ok && owner.managerOwner() != m {
			return errors.New("server belongs to a different manager")
		}
	}
	m.mu.Lock()
	if m.IsClose() {
		m.mu.Unlock()
		return ErrServerManagerClosed
	}
	if m.started {
		m.mu.Unlock()
		return ErrServerManagerStarted
	}
	m.started = true
	m.mu.Unlock()
	watchDone := make(chan struct{})
	defer close(watchDone)
	go func() {
		select {
		case <-ctx.Done():
			m.StopAll()
		case <-watchDone:
		}
	}()
	var services sync.WaitGroup
	results := make(chan error, len(servers))
	for _, server := range servers {
		services.Add(1)
		go func(server IServer) {
			defer services.Done()
			defer func() {
				if recovered := recover(); recovered != nil {
					results <- fmt.Errorf("%s panic: %v", server.GetServerName(), recovered)
					m.StopAll()
				}
			}()
			if managed, ok := server.(ManagedServer); ok {
				if err := managed.Serve(m.ctx); err != nil {
					results <- fmt.Errorf("%s: %w", server.GetServerName(), err)
					m.StopAll()
				}
			} else {
				server.Start()
			}
		}(server)
	}
	<-m.ctx.Done()
	m.connections.ClearConn()
	services.Wait()
	m.finish()
	close(results)
	var failures []error
	for err := range results {
		failures = append(failures, err)
	}
	return errors.Join(failures...)
}
func (m *ServerManager) IsClose() bool          { return atomic.LoadInt32(&m.isClosed) != 0 }
func (m *ServerManager) WaitGroupAdd(delta int) { m.waitGroup.Add(delta) }
func (m *ServerManager) WaitGroupDone()         { m.waitGroup.Done() }

// Stop requests shutdown without waiting.
func (m *ServerManager) Stop() { m.StopAll() }

func (m *ServerManager) StopAll() {
	m.mu.Lock()
	first := atomic.CompareAndSwapInt32(&m.isClosed, 0, 1)
	if first {
		m.cancel()
	}
	started := m.started
	m.mu.Unlock()
	if first && !started {
		go m.finish()
	}
}

func (m *ServerManager) finish() {
	m.shutdownOnce.Do(func() {
		m.connections.ClearConn()
		m.connections.Wait()
		m.WorkerPool().Stop()
		m.waitGroup.Wait()
		close(m.done)
	})
}

// Shutdown requests termination and waits, bounded by the caller's context.
// A handler that ignores cancellation can delay termination; the caller may time out.
func (m *ServerManager) Shutdown(ctx context.Context) error {
	m.Stop()
	select {
	case <-m.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func managerForServer(server IServer) *ServerManager {
	if owner, ok := server.(interface{ managerOwner() *ServerManager }); ok {
		return owner.managerOwner()
	}
	return getDefaultServerManager()
}
func managerForConnection(conn IConnection) *ServerManager {
	if owner, ok := conn.(interface{ managerOwner() *ServerManager }); ok {
		return owner.managerOwner()
	}
	return getDefaultServerManager()
}

var defaultServerManagerMu sync.Mutex
var defaultServerManager *ServerManager

func getDefaultServerManager() *ServerManager {
	defaultServerManagerMu.Lock()
	defer defaultServerManagerMu.Unlock()
	if defaultServerManager == nil {
		var err error
		defaultServerManager, err = NewServerManager(defaultServer)
		if err != nil {
			panic(fmt.Sprintf("nets: invalid default configuration: %v", err))
		}
	}
	return defaultServerManager
}
