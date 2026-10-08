package nets

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func withHTTPTestConf(t *testing.T, update func(*AppConf)) *serverHTTP {
	t.Helper()
	conf := GetServerConf()
	update(&conf)
	manager, err := NewServerManager(&CustomServer{AppConf: conf})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := manager.Shutdown(ctx); err != nil {
			t.Error(err)
		}
	})
	return manager.GetServerHTTP().(*serverHTTP)
}

func TestServerHTTPRejectsUnsupportedMethod(t *testing.T) {
	server := withHTTPTestConf(t, func(conf *AppConf) { conf.MaxConn = 1 })
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	server.handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("状态码 = %d, 期望 %d", recorder.Code, http.StatusMethodNotAllowed)
	}
	if allow := recorder.Header().Get("Allow"); allow != http.MethodPost {
		t.Fatalf("Allow = %q, 期望 %q", allow, http.MethodPost)
	}
}

func TestServerHTTPRejectsOversizedBody(t *testing.T) {
	server := withHTTPTestConf(t, func(conf *AppConf) {
		conf.MaxConn = 1
		conf.MaxPackSize = 1
	})
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("x", int((int64(server.manager.Config().MaxPackSize)*2+httpBodyExtraBytes)+1))))
	recorder := httptest.NewRecorder()

	server.handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("状态码 = %d, 期望 %d", recorder.Code, http.StatusRequestEntityTooLarge)
	}
	if active := server.activeRequest.Load(); active != 0 {
		t.Fatalf("请求结束后 activeRequest = %d, 期望 0", active)
	}
}

func TestServerHTTPRejectsChunkedOversizedBody(t *testing.T) {
	server := withHTTPTestConf(t, func(conf *AppConf) {
		conf.MaxConn = 1
		conf.MaxPackSize = 1
	})
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("x", int((int64(server.manager.Config().MaxPackSize)*2+httpBodyExtraBytes)+1))))
	request.ContentLength = -1
	recorder := httptest.NewRecorder()

	server.handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("状态码 = %d, 期望 %d", recorder.Code, http.StatusRequestEntityTooLarge)
	}
	if active := server.activeRequest.Load(); active != 0 {
		t.Fatalf("请求结束后 activeRequest = %d, 期望 0", active)
	}
}

func TestServerHTTPRejectsWhenCapacityIsExhausted(t *testing.T) {
	server := withHTTPTestConf(t, func(conf *AppConf) { conf.MaxConn = 0 })
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))

	server.handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("状态码 = %d, 期望 %d", recorder.Code, http.StatusServiceUnavailable)
	}
}
