package nets

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func withHTTPTestConf(t *testing.T, update func(*AppConf)) {
	t.Helper()
	original := defaultServer.AppConf
	t.Cleanup(func() { defaultServer.AppConf = original })
	update(&defaultServer.AppConf)
}

func TestServerHTTPRejectsUnsupportedMethod(t *testing.T) {
	withHTTPTestConf(t, func(conf *AppConf) { conf.MaxConn = 1 })
	server := &serverHTTP{}
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
	withHTTPTestConf(t, func(conf *AppConf) {
		conf.MaxConn = 1
		conf.MaxPackSize = 1
	})
	server := &serverHTTP{}
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("x", int(getHTTPBodyLimit()+1))))
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
	withHTTPTestConf(t, func(conf *AppConf) {
		conf.MaxConn = 1
		conf.MaxPackSize = 1
	})
	server := &serverHTTP{}
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("x", int(getHTTPBodyLimit()+1))))
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
	withHTTPTestConf(t, func(conf *AppConf) { conf.MaxConn = 0 })
	server := &serverHTTP{}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))

	server.handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("状态码 = %d, 期望 %d", recorder.Code, http.StatusServiceUnavailable)
	}
}
