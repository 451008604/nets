# Go pprof 性能分析使用指南

pprof 是 Go 内置的性能分析工具，可以分析 CPU、内存、goroutine 等性能指标。

---

## 1. 在代码中启用 pprof

```go
package main

import (
    "log"
    "net/http"
    _ "net/http/pprof"  // 只需导入，自动注册路由
)

func main() {
    // pprof 会注册到默认的 http.DefaultServeMux
    // 访问 http://localhost:6060/debug/pprof/
    log.Println(http.ListenAndServe("localhost:6060", nil))
}
```

**注意**：nets 框架的 `test/server/server.go` 已导入 (`_ "net/http/pprof"`)，只需要在某个端口启动 HTTP 服务即可访问。

---

## 2. 常用分析端点

| 端点 | 说明 |
|------|------|
| `/debug/pprof/` | 所有 profile 列表 |
| `/debug/pprof/profile` | CPU profile（30秒采样）|
| `/debug/pprof/heap` | 内存分配情况 |
| `/debug/pprof/goroutine` | goroutine 堆栈 |
| `/debug/pprof/allocs` | 历史内存分配 |
| `/debug/pprof/block` | 阻塞分析 |
| `/debug/pprof/mutex` | 锁竞争分析 |
| `/debug/pprof/threadcreate` | 线程创建分析 |

---

## 3. 命令行使用

### 3.1 CPU 分析（采集 30 秒）
```bash
go tool pprof http://localhost:6060/debug/pprof/profile
```

### 3.2 内存分析
```bash
# 查看当前内存分配
go tool pprof http://localhost:6060/debug/pprof/heap

# 查看历史内存分配（包含已释放的内存）
go tool pprof http://localhost:6060/debug/pprof/allocs
```

### 3.3 Goroutine 分析
```bash
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### 3.4 下载文件后离线分析
```bash
curl http://localhost:6060/debug/pprof/heap > heap.out
go tool pprof heap.out
```

---

## 4. 交互式分析命令

进入 pprof 交互界面后，常用命令：

```bash
# 查看 Top 10 热点
(pprof) top10

# 查看 Top 20
(pprof) top20

# 生成火焰图（需要安装 graphviz）
(pprof) web

# 生成文本调用图
(pprof) text

# 查看具体函数源码及性能数据
(pprof) list YourFunctionName

# 输出 SVG 图片
(pprof) svg > profile.svg

# 输出 PNG 图片
(pprof) png > profile.png

# 显示调用图
(pprof) png

# 退出
(pprof) quit
```

---

## 5. 火焰图（最直观）

### 5.1 安装 pprof 工具
```bash
go install github.com/google/pprof@latest
```

### 5.2 启动 Web 界面
```bash
# 直接打开浏览器查看火焰图
pprof -http=:8080 http://localhost:6060/debug/pprof/profile

# 查看内存火焰图
pprof -http=:8080 http://localhost:6060/debug/pprof/heap

# 查看 goroutine 火焰图
pprof -http=:8080 http://localhost:6060/debug/pprof/goroutine
```

访问 http://localhost:8080 查看交互式火焰图。

---

## 6. 在 nets 项目中使用

nets 框架的 `test/server/server.go` 已经导入了 pprof。启用方式：

```go
package main

import (
    "log"
    "net/http"
    _ "net/http/pprof"
    
    "github.com/451008604/nets"
)

func main() {
    // 在另一个端口启动 pprof（避免和业务端口冲突）
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // 你的业务代码...
    nets.GetInstanceServerManager().RegisterServer(...)
}
```

---

## 7. 快速检查内存泄漏

```bash
# 1. 采集第一次 heap
 curl -s http://localhost:6060/debug/pprof/heap > heap1.out

# 2. 运行一段时间，让程序处理更多请求...
sleep 60

# 3. 采集第二次 heap
curl -s http://localhost:6060/debug/pprof/heap > heap2.out

# 4. 对比两次内存差异
go tool pprof --diff_base heap1.out heap2.out
```

---

## 8. 常用分析场景

### 8.1 CPU 性能瓶颈
```bash
# 采集 30 秒 CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile

# 在交互式界面中
(pprof) top10          # 查看最耗时的函数
(pprof) list main      # 查看 main 包的函数详情
(pprof) web            # 生成火焰图
```

### 8.2 内存泄漏排查
```bash
# 查看当前内存分配
go tool pprof http://localhost:6060/debug/pprof/heap

# 在交互式界面中
(pprof) top10          # 查看内存分配最多的函数
(pprof) list main      # 查看具体代码行的内存分配
(pprof) alloc_space    # 查看累计分配内存
(pprof) inuse_space    # 查看当前持有的内存
```

### 8.3 Goroutine 泄漏
```bash
# 查看所有 goroutine 堆栈
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 直接查看文本格式的 goroutine 堆栈
curl http://localhost:6060/debug/pprof/goroutine?debug=1

# 查看详细的 goroutine 堆栈（包含更多信息）
curl http://localhost:6060/debug/pprof/goroutine?debug=2
```

### 8.4 阻塞分析
```bash
# 需要先在代码中开启阻塞分析
import "runtime"

func init() {
    runtime.SetBlockProfileRate(1) // 采集所有阻塞事件
}

# 然后查看阻塞分析
go tool pprof http://localhost:6060/debug/pprof/block
```

---

## 9. 图形化界面选项

### 9.1 使用 `-http` 标志
```bash
# 在本机启动 Web 服务器查看 profile
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap
```

### 9.2 使用 `--nodefraction` 过滤
```bash
# 只显示占用超过 5% 的节点
go tool pprof -http=:8080 --nodefraction=0.05 heap.out
```

### 9.3 使用 `--edgefraction` 过滤边
```bash
# 只显示占用超过 10% 的调用边
go tool pprof -http=:8080 --edgefraction=0.10 heap.out
```

---

## 10. 常见问题

### Q: 为什么 `web` 命令报错？
A: 需要安装 graphviz：
```bash
# macOS
brew install graphviz

# Ubuntu/Debian
sudo apt-get install graphviz

# Windows
choco install graphviz
```

### Q: 如何只采集特定时间的 CPU profile？
A: 使用 `seconds` 参数：
```bash
# 采集 10 秒
curl http://localhost:6060/debug/pprof/profile?seconds=10 > cpu.prof
go tool pprof cpu.prof
```

### Q: 如何分析测试的 profile？
A: 运行测试时生成 profile：
```bash
go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...
go tool pprof cpu.prof
```

### Q: pprof 对性能影响大吗？
A: 
- CPU profiling：约 5% 的性能损耗
- Memory profiling：约 10-20% 的性能损耗
- Goroutine/block/mutex profiling：影响很小
- 生产环境建议只在需要时开启，或采样较低频率

---

## 11. 参考链接

- [Go pprof 官方文档](https://golang.org/pkg/net/http/pprof/)
- [Profiling Go Programs](https://go.dev/blog/pprof)
- [Go 性能分析工具](https://github.com/google/pprof)
- [火焰图介绍](http://www.brendangregg.com/flamegraphs.html)

---

*Generated for nets project*
