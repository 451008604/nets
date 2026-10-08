package nets

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"
)

func serveConnections(ctx context.Context, listener net.Listener, factory func(net.Conn) IConnection, manager *ConnectionManager) error {
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() { defer close(done); <-ctx.Done(); _ = listener.Close() }()
	defer func() { cancel(); _ = listener.Close(); <-done }()
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		if ctx.Err() != nil {
			_ = conn.Close()
			continue
		}
		manager.TryAdd(factory(conn))
	}
}

func serveHTTP(ctx context.Context, r *ServerManager, conf ServerConf, handler http.Handler) error {
	if conf.Port == 0 || ctx.Err() != nil {
		return nil
	}
	listener, err := net.Listen("tcp", net.JoinHostPort(conf.Address, strconv.Itoa(conf.Port)))
	if err != nil {
		return err
	}
	defer listener.Close()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	timeout := time.Duration(r.config.AppConf.ConnRWTimeOut) * time.Second
	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: timeout, ReadTimeout: timeout, WriteTimeout: timeout, IdleTimeout: timeout,
		MaxHeaderBytes: maxHTTPHeaderBytes,
		ErrorLog:       slog.NewLogLogger(r.config.Logger.Handler(), slog.LevelError),
		BaseContext:    func(net.Listener) context.Context { return ctx },
	}
	shutdownDone := make(chan error, 1)
	go func() {
		<-ctx.Done()
		// Request cancellation is immediate; Shutdown still waits for handlers to return.
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), defaultHTTPTimeout)
		defer shutdownCancel()
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			err = errors.Join(err, server.Close())
		}
		shutdownDone <- err
	}()
	if conf.TLSCertPath != "" {
		err = server.ServeTLS(listener, conf.TLSCertPath, conf.TLSKeyPath)
	} else {
		err = server.Serve(listener)
	}
	cancel()
	shutdownErr := <-shutdownDone
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	return errors.Join(err, shutdownErr)
}
