# socketServerFrame

## 开发计划

✅ 采用工作池模式实现协程复用，用于降低并发下协程创建销毁对性能的开销，支持动态扩容  
✅ 使用 protobuf 且兼容 json 进行通讯，支持 grpc  
✅ 日志的收集可通过飞书机器人实现 webhook 推送  
✅ 广播消息功能，可1对1、1对多、服务端主动进行群广播  
📌 引入 [zap](https://github.com/uber-go/zap) 重新设计日志收集系统  
📌 引入 [gorm](https://github.com/go-gorm/gorm) 用于数据库对增、删、改、查  
📌 优化配置文件读取

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
