package main

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestHTTPSAndWSSUseIndependentTLSNegotiation(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/ws" {
			w.Write([]byte("ok"))
			return
		}
		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		kind, data, err := conn.ReadMessage()
		if err == nil {
			conn.WriteMessage(kind, data)
		}
	}))
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()
	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())
	original := clientTLSConfig
	clientTLSConfig = &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}
	defer func() { clientTLSConfig = original }()

	// net/http configures ALPN when its first request starts. A later WebSocket
	// client must still negotiate HTTP/1.1 using the same trust settings.
	https := NewHTTPClient(server.URL)
	defer https.Close()
	response, err := https.client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.ProtoMajor != 2 {
		t.Fatal("HTTP/2 setup was not exercised")
	}
	wss, err := NewWSClient("wss" + server.URL[len("https"):] + "/ws")
	if err != nil {
		t.Fatal(err)
	}
	defer wss.Close()
	wss.conn.SetReadDeadline(time.Now().Add(time.Second))
	if state := wss.conn.UnderlyingConn().(*tls.Conn).ConnectionState(); len(state.VerifiedChains) == 0 || (state.NegotiatedProtocol != "" && state.NegotiatedProtocol != "http/1.1") {
		t.Fatalf("unexpected WSS TLS negotiation: %+v", state)
	}
	if err := wss.Write(1001, []byte("echo")); err != nil {
		t.Fatal(err)
	}
	buffer := make([]byte, 128)
	n, err := wss.Read(buffer)
	if err != nil {
		t.Fatal(err)
	}
	message := unpack(buffer[:n])
	if message == nil || message.Id != 1001 || string(message.Data) != "echo" {
		t.Fatal("WSS echo did not survive mixed TLS negotiation")
	}
}
