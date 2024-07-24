package network

import (
	"github.com/451008604/nets/iface"
	"github.com/golang/mock/gomock"
	"sync"
	"testing"
)

func TestAddAndRemoveConnection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := &connectionManager{
		connections: NewConcurrentStringer[Integer, iface.IConnection](),
		closeConnId: make(chan int, defaultServer.AppConf.MaxConn),
		removeList:  make(chan iface.IConnection, defaultServer.AppConf.MaxConn),
	}
	go onConnRemoveList(manager)

	wg := sync.WaitGroup{}
	wg.Add(100000)
	for i := 0; i < 100000; i++ {
		go func() {
			defer wg.Done()
			conn := iface.NewMockIConnection(ctrl)
			conn.EXPECT().GetConnId().Return(manager.NewConnId()).AnyTimes()
			conn.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes()
			conn.EXPECT().Stop().AnyTimes()
			conn.EXPECT().IsClose().Return(false).AnyTimes()

			manager.Add(conn)
			manager.Remove(conn)
		}()
	}
	wg.Wait()
}

func TestGetConnection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := &connectionManager{
		connections: NewConcurrentStringer[Integer, iface.IConnection](),
		closeConnId: make(chan int, defaultServer.AppConf.MaxConn),
		removeList:  make(chan iface.IConnection, defaultServer.AppConf.MaxConn),
	}
	go onConnRemoveList(manager)

	wg := sync.WaitGroup{}
	wg.Add(100000)
	for i := 0; i < 100000; i++ {
		go func() {
			defer wg.Done()
			id := manager.NewConnId()
			conn := iface.NewMockIConnection(ctrl)
			conn.EXPECT().GetConnId().Return(id).AnyTimes()
			conn.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes()

			manager.Add(conn)
			retrievedConn, ok := manager.Get(id)
			if !ok || retrievedConn != conn {
				t.Errorf("expected to retrieve the added connection")
			}
		}()
	}
	wg.Wait()
}

func TestClearConnections(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := &connectionManager{
		connections: NewConcurrentStringer[Integer, iface.IConnection](),
		closeConnId: make(chan int, defaultServer.AppConf.MaxConn),
		removeList:  make(chan iface.IConnection, defaultServer.AppConf.MaxConn),
	}
	go onConnRemoveList(manager)

	for i := 0; i < 100000; i++ {
		conn := iface.NewMockIConnection(ctrl)
		conn.EXPECT().GetConnId().Return(manager.NewConnId()).AnyTimes()
		conn.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes()
		conn.EXPECT().Stop().AnyTimes()
		conn.EXPECT().IsClose().Return(false).AnyTimes()
		manager.Add(conn)
	}
	for i := 0; i < 10; i++ {
		go manager.ClearConn()
	}
	for {
		if l := manager.Len(); l == 0 {
			break
		}
	}
}

func TestConnHooks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	conn := iface.NewMockIConnection(ctrl)
	conn.EXPECT().GetConnId().Return(1).AnyTimes()
	conn.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes()

	manager := GetInstanceConnManager()
	manager.Add(conn)

	openedCalled := false
	closedCalled := false

	manager.SetConnOnOpened(func(c iface.IConnection) {
		openedCalled = true
	})
	manager.SetConnOnClosed(func(c iface.IConnection) {
		closedCalled = true
	})

	manager.ConnOnOpened(conn)
	if !openedCalled {
		t.Errorf("expected opened hook to be called")
	}

	manager.ConnOnClosed(conn)
	if !closedCalled {
		t.Errorf("expected closed hook to be called")
	}
}
