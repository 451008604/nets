# socketServerFrame

- 受 [zinx](https://github.com/aceld/zinx) 框架启发采用工作池模式实现消息队列协程复用，用于降低高并发下的协程开销
- 同时引入 [GoFrame](https://github.com/gogf/gf) 的 mysql 数据库模块用于支持数据的增、删、改、查
- 使用 protobuf 进行通讯且支持 grpc 做微服务开发
- 优化日志的收集可通过简单扩展实现 webhook 推送
- 实现推送广播消息功能，可1v1、1vN、StoC群广播

```
├── README.md
├── main.go                             // 入口文件
├── go.mod
├── go.sum
├── generate_pb.sh                      // proto 文件转 pb.go 脚本
├── generate_sqlModel.sh                // 使用 gf 工具生成 sql 数据模型脚本
├── api                                 // apis 逻辑实现
│    ├── base.go
│    ├── login.go
│    ├── ping.go
│    └── router.go
├── client                              // 客户端服务（示例代码，可直接删除）
│    ├── base
│    └── main.go
├── config                              // 项目配置文件
│    ├── conf.go
│    └── config.json
├── database                            // 数据库相关内容
│    ├── init.go                        // 初始化数据库相关服务
│    └── sql                            // 使用 gf 工具自动生成的 sql 数据表模型
├── iface                               // 定义抽象接口
│    ├── iconnection.go                 // 连接对象（一个客户端对应一个连接）
│    ├── iconnmanager.go                // 连接管理器
│    ├── idatapack.go                   // 处理消息封包/拆包
│    ├── imessage.go                    // 消息对象（包含消息ID、消息长度、消息内容）
│    ├── imsghandler.go                 // 消息处理对象（接收消息处理中间层）
│    ├── inotify.go                     // 广播对象
│    ├── irequest.go                    // 请求对象（包含请求对应的连接和数据）
│    ├── irouter.go                     // 路由对象
│    └── iserver.go                     // 服务器对象（负责管理整个服务的生命周期以及连接建立和断开时的处理）
├── logic                               // 逻辑层，作为 api 的上层（负责处理公共逻辑）
│    └── base.go                        // 初始化所有的功能模块，方便自身和其他模块的引用
├── logs                                // 打印日志管理器（异步打印）
│    ├── printlog2console.go            // 输出到控制台
│    ├── printlog2file.go               // 输出到日志文件（文件会按照50M的大小进行切割）
│    └── printlogManager.go             // 对外提供打印接口（可以控制输出到控制台或日志文件）
├── network                             // 网络层的抽象接口实现
│    ├── connection.go
│    ├── connmanager.go
│    ├── datapack.go
│    ├── message.go
│    ├── msghandler.go
│    ├── notify.go
│    ├── notifyManager.go
│    ├── request.go
│    ├── router.go
│    └── server.go
├── proto                               // 管理 proto 文件
│    ├── bin
│    └── src
└── shell                               // 项目中需要用到的工具
    ├── gf_2.3.0_m1                     // M1 系统的 gf 工具
    └── gf_2.3.0_windows.exe            // windows 系统的 gf 工具
```

## 使用 gf 生成 mysql_model

在根目录创建一个 sh 脚本（generate_sqlModel.sh）内容如下，不同系统需自行修改 shell 下的二进制引用

```shell
# windows
./shell/gf_2.3.0_windows.exe gen dao -l "mysql:userName:userPass@tcp(127.0.0.1:3306)/DBName?charset=utf8mb4&parseTime=true&loc=Local" -p ./database/sql
```

## grpc 配置

- 安装protoc编译器

> https://github.com/protocolbuffers/protobuf/releases/  
> 下载后解压到任意目录把`bin`里面的`protoc.exe`复制到`%GOPATH%/bin`里面，并配置`PATH`环境变量，确保 protoc 可以正常执行

- 安装相关模块

> go install google.golang.org/protobuf/proto  
> go install google.golang.org/protobuf/cmd/protoc-gen-go@latest  
> go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest  
