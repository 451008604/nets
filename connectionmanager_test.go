package nets

import (
	"net"
	"testing"
	"time"
)

func TestConnectionManager_Add(t *testing.T) {
	connManager := GetInstanceConnManager()
	go func() {
		for i := 0; i < 10000; i++ {
			connManager.Add(NewConnectionTCP(GetServerTCP(), &net.TCPConn{}))
		}
	}()

	go func() {
		for i := 0; i < 10000; i++ {
			connManager.Remove(NewConnectionTCP(GetServerTCP(), &net.TCPConn{}))
		}
	}()

	time.Sleep(5 * time.Second)
	if connManager.Len() != 0 {
		t.Error("TestConnectionManager_Add", connManager.Len())
	}
}
