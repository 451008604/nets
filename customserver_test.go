package nets

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestSetCustomServerUsesInjectedLogger(t *testing.T) {
	original := *defaultServer
	t.Cleanup(func() { *defaultServer = original })

	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil)).With("service", "gateway")
	processDefault := slog.Default()
	SetCustomServer(&CustomServer{
		AppConf: GetServerConf(),
		Logger:  logger,
	})
	if slog.Default() != processDefault {
		t.Fatal("配置 SDK Logger 不应修改进程级 slog 默认实例")
	}

	getLogger().Info("sdk message", "connection_id", "conn-1")
	logLine := output.String()
	for _, expected := range []string{
		`"msg":"sdk message"`,
		`"service":"gateway"`,
		`"connection_id":"conn-1"`,
	} {
		if !strings.Contains(logLine, expected) {
			t.Fatalf("日志未保留注入属性 %s: %s", expected, logLine)
		}
	}
}

func TestSetCustomServerBuildsFallbackLoggerAtConfiguredLevel(t *testing.T) {
	original := *defaultServer
	t.Cleanup(func() { *defaultServer = original })

	conf := GetServerConf()
	conf.LogLevel = slog.LevelWarn
	SetCustomServer(&CustomServer{AppConf: conf})

	ctx := context.Background()
	if getLogger().Enabled(ctx, slog.LevelInfo) {
		t.Fatal("Info 日志不应在 Warn 级别启用")
	}
	if !getLogger().Enabled(ctx, slog.LevelWarn) {
		t.Fatal("Warn 日志应在 Warn 级别启用")
	}
}
