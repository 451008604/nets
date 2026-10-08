package nets

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xtaci/kcp-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func testTLSCertificate(t *testing.T) (ServerConf, *tls.Config) {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "localhost"},
		NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour),
		DNSNames: []string{"localhost"}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:        true, BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, public, private)
	if err != nil {
		t.Fatal(err)
	}
	key, err := x509.MarshalPKCS8PrivateKey(private)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	conf := ServerConf{Address: "127.0.0.1", TLSCertPath: filepath.Join(dir, "cert.pem"), TLSKeyPath: filepath.Join(dir, "key.pem")}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	if err := os.WriteFile(conf.TLSCertPath, certPEM, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(conf.TLSKeyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key}), 0600); err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(certPEM) {
		t.Fatal("certificate was not added to trust pool")
	}
	return conf, &tls.Config{RootCAs: roots, ServerName: "localhost", MinVersion: tls.VersionTLS12}
}

func streamTLSServer(m *ServerManager, protocol string) IServer {
	if protocol == "tcp" {
		return m.GetServerTCP()
	}
	return m.GetServerKCP()
}

func configureStreamTLS(app *AppConf, protocol string, conf ServerConf) {
	if protocol == "tcp" {
		app.ServerTCP = conf
	} else {
		app.ServerKCP = conf
	}
}

func dialTLSStream(protocol, address string, config *tls.Config) (*tls.Conn, error) {
	var raw net.Conn
	var err error
	if protocol == "tcp" {
		raw, err = net.DialTimeout("tcp", address, 100*time.Millisecond)
	} else {
		var session *kcp.UDPSession
		session, err = kcp.DialWithOptions(address, nil, 0, 0)
		if err == nil {
			session.SetStreamMode(true)
			session.SetNoDelay(1, 10, 2, 1)
			session.SetWindowSize(256, 256)
			raw = session
		}
	}
	if err != nil {
		return nil, err
	}
	secure := tls.Client(raw, config)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := secure.HandshakeContext(ctx); err != nil {
		_ = raw.Close()
		return nil, err
	}
	return secure, nil
}

func TestStreamTLSRoundTrip(t *testing.T) {
	certificate, clientConfig := testTLSCertificate(t)
	for _, protocol := range []string{"tcp", "kcp"} {
		for _, version := range []uint16{tls.VersionTLS12, tls.VersionTLS13} {
			for _, jsonProtocol := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/version=%x/json=%v", protocol, version, jsonProtocol), func(t *testing.T) {
					conf := certificate
					conf.Port = availablePort(t, protocol == "kcp")
					m := testServerManager(t, func(app *AppConf) {
						app.ProtocolIsJson, app.MaxPackSize = jsonProtocol, 65535
						configureStreamTLS(app, protocol, conf)
					})
					opened := make(chan tls.ConnectionState, 1)
					m.connections.SetConnOpened(func(conn IConnection) {
						opened <- conn.GetNetConn().(*tls.Conn).ConnectionState()
					})
					m.messages.AddRouter(1001, func() proto.Message { return &wrapperspb.StringValue{} }, func(conn IConnection, msg proto.Message) { conn.SendMsg(1001, msg) })
					startTestServer(t, m, streamTLSServer(m, protocol))
					config := clientConfig.Clone()
					config.MinVersion, config.MaxVersion = version, version
					address := net.JoinHostPort(conf.Address, strconv.Itoa(conf.Port))
					conn := retry(t, func() (*tls.Conn, error) { return dialTLSStream(protocol, address, config) })
					defer conn.Close()
					if state := await(t, opened); !state.HandshakeComplete || state.Version != version {
						t.Fatalf("opened hook ran without negotiated TLS: %+v", state)
					}
					if len(conn.ConnectionState().VerifiedChains) == 0 {
						t.Fatal("peer certificate was not verified")
					}
					conn.SetDeadline(time.Now().Add(2 * time.Second))
					// Cross both TLS record and KCP segment boundaries.
					frame, err := (&ConnectionBase{manager: m}).pack(1001, wrapperspb.String(strings.Repeat("echo测试", 2500)))
					if err != nil {
						t.Fatal(err)
					}
					if !writeStream(conn, frame) {
						t.Fatal("TLS frame write failed")
					}
					response := make([]byte, len(frame))
					if _, err := io.ReadFull(conn, response); err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(response, frame) {
						t.Fatal("TLS round trip changed frame contents")
					}
				})
			}
		}
	}
}

func TestStreamTLSRejectsInvalidPeers(t *testing.T) {
	certificate, clientConfig := testTLSCertificate(t)
	_, otherCA := testTLSCertificate(t)
	for _, protocol := range []string{"tcp", "kcp"} {
		t.Run(protocol, func(t *testing.T) {
			conf := certificate
			conf.Port = availablePort(t, protocol == "kcp")
			m := testServerManager(t, func(app *AppConf) { configureStreamTLS(app, protocol, conf) })
			opened := make(chan struct{}, 4)
			m.connections.SetConnOpened(func(IConnection) { opened <- struct{}{} })
			startTestServer(t, m, streamTLSServer(m, protocol))
			address := net.JoinHostPort(conf.Address, strconv.Itoa(conf.Port))
			// Establish readiness without weakening certificate validation.
			ready := retry(t, func() (*tls.Conn, error) { return dialTLSStream(protocol, address, clientConfig) })
			await(t, opened)
			ready.Close()
			for _, failure := range []string{"untrusted-ca", "wrong-name", "obsolete-version"} {
				t.Run(failure, func(t *testing.T) {
					config := clientConfig.Clone()
					switch failure {
					case "untrusted-ca":
						config.RootCAs = otherCA.RootCAs
					case "wrong-name":
						config.ServerName = "other.invalid"
					case "obsolete-version":
						config.MinVersion, config.MaxVersion = tls.VersionTLS10, tls.VersionTLS11
					}
					conn, err := dialTLSStream(protocol, address, config)
					if err == nil {
						conn.Close()
						t.Fatal("invalid TLS peer was accepted")
					}
				})
			}
			// A failed handshake affects only that connection, not the accept loop.
			valid := retry(t, func() (*tls.Conn, error) { return dialTLSStream(protocol, address, clientConfig) })
			defer valid.Close()
			await(t, opened)
			select {
			case <-opened:
				t.Fatal("failed TLS handshake ran opened hook")
			default:
			}
		})
	}
}

type handshakeProbe struct {
	net.Conn
	readStarted chan struct{}
	once        sync.Once
}

func (c *handshakeProbe) Read(data []byte) (int, error) {
	c.once.Do(func() { close(c.readStarted) })
	return c.Conn.Read(data)
}

func TestStreamTLSHandshakeLifecycle(t *testing.T) {
	certificate, _ := testTLSCertificate(t)
	config, err := loadStreamTLSConfig(certificate)
	if err != nil {
		t.Fatal(err)
	}
	for _, protocol := range []string{"tcp", "kcp"} {
		for _, action := range []string{"timeout", "shutdown", "plaintext"} {
			t.Run(protocol+"/"+action, func(t *testing.T) {
				m := testServerManager(t, func(app *AppConf) { app.MaxConn, app.ConnRWTimeOut = 1, 1 })
				left, right := net.Pipe()
				t.Cleanup(func() { left.Close(); right.Close() })
				probe := &handshakeProbe{Conn: left, readStarted: make(chan struct{})}
				secure := tls.Server(probe, config)
				var c IConnection
				if protocol == "tcp" {
					c = newConnectionTCP(m.GetServerTCP(), secure)
				} else {
					c = NewConnectionKCP(m.GetServerKCP(), secure)
				}
				opened := make(chan struct{}, 1)
				m.connections.SetConnOpened(func(IConnection) { opened <- struct{}{} })
				m.messages.AddRouter(1001, func() proto.Message { return &wrapperspb.StringValue{} }, func(IConnection, proto.Message) { t.Error("plaintext reached business handler") })
				if !m.connections.TryAdd(c) {
					t.Fatal("TLS connection rejected before handshake")
				}
				await(t, probe.readStarted)
				if m.connections.acquireRequest() {
					m.connections.releaseRequest()
					t.Fatal("pending handshake did not consume shared connection capacity")
				}
				switch action {
				case "shutdown":
					ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
					defer cancel()
					if err := m.Shutdown(ctx); err != nil {
						t.Fatalf("pending handshake blocked shutdown: %v", err)
					}
				case "plaintext":
					right.SetWriteDeadline(time.Now().Add(time.Second))
					frame, err := (&ConnectionBase{manager: m}).pack(1001, wrapperspb.String("plaintext"))
					if err != nil {
						t.Fatal(err)
					}
					_, _ = right.Write(frame) // The peer may close after the first five bytes.
				}
				done := make(chan struct{})
				go func() { m.connections.Wait(); close(done) }()
				await(t, done)
				if m.connections.Len() != 0 || !c.IsClose() {
					t.Fatal("failed handshake retained connection registration")
				}
				select {
				case <-opened:
					t.Fatal("incomplete handshake ran opened hook")
				default:
				}
				if action != "shutdown" {
					if !m.connections.acquireRequest() {
						t.Fatal("failed handshake leaked capacity")
					}
					m.connections.releaseRequest()
				}
			})
		}
	}
}

func TestTLSCloseDoesNotWaitForPeerToRead(t *testing.T) {
	certificate, clientConfig := testTLSCertificate(t)
	config, err := loadStreamTLSConfig(certificate)
	if err != nil {
		t.Fatal(err)
	}
	// TLS 1.2 completes without post-handshake tickets requiring extra peer reads.
	config.MaxVersion, clientConfig.MaxVersion = tls.VersionTLS12, tls.VersionTLS12
	m := testServerManager(t, nil)
	left, right := net.Pipe()
	t.Cleanup(func() { left.Close(); right.Close() })
	c := NewConnectionKCP(m.GetServerKCP(), tls.Server(left, config))
	opened := make(chan struct{})
	m.connections.SetConnOpened(func(IConnection) { close(opened) })
	if !m.connections.TryAdd(c) {
		t.Fatal("TLS connection rejected")
	}
	client := tls.Client(right, clientConfig)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.HandshakeContext(ctx); err != nil {
		t.Fatal(err)
	}
	await(t, opened)
	closed := make(chan struct{})
	go func() { c.Close(); close(closed) }()
	// The peer deliberately does not read close_notify; Close must still return.
	await(t, closed)
}

func TestStreamTLSInvalidCertificateStopsSibling(t *testing.T) {
	certificate, _ := testTLSCertificate(t)
	other, _ := testTLSCertificate(t)
	for _, protocol := range []string{"tcp", "kcp"} {
		for _, invalid := range []string{"missing", "mismatched"} {
			t.Run(protocol+"/"+invalid, func(t *testing.T) {
				conf := certificate
				conf.Port = availablePort(t, protocol == "kcp")
				if invalid == "missing" {
					conf.TLSCertPath = filepath.Join(t.TempDir(), "missing.pem")
				} else {
					conf.TLSKeyPath = other.TLSKeyPath
				}
				m := testServerManager(t, func(app *AppConf) { configureStreamTLS(app, protocol, conf) })
				peer := &cleanupServer{make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})}
				close(peer.release)
				result := make(chan error, 1)
				go func() { result <- m.Run(context.Background(), peer, streamTLSServer(m, protocol)) }()
				if err := await(t, result); err == nil || !strings.Contains(err.Error(), strings.ToUpper(protocol)+" TLS") {
					t.Fatalf("certificate error not reported by protocol: %v", err)
				}
				await(t, peer.stopped)
			})
		}
	}
}
