# socketServerFrame

## 开发计划

✅ 采用工作池模式实现协程复用，用于降低并发下协程创建销毁对性能的开销，支持动态扩容  
✅ 使用 protobuf 且兼容 json 进行通讯，支持 grpc  
✅ 日志的收集可通过飞书机器人实现 webhook 推送  
✅ 广播消息功能，可1对1、1对多、服务端主动进行群广播  
✅ 引入 [gorm](https://github.com/go-gorm/gorm) 用于数据库对增、删、改、查  
📌 引入 [zap](https://github.com/uber-go/zap) 重新设计日志收集系统  
📌 优化配置文件读取

## 使用 gentool 生成 dao 文件
```shell
go install gorm.io/gen/tools/gentool@latest
```

在根目录增加yml文件，内容如下

```yaml
version : "0.1"
database:
    dsn              : "root:userName:userPass@tcp(127.0.0.1:3306)/DBName?charset=utf8mb4&parseTime=true&loc=Local"
    db               : "mysql"
    withUnitTest     : false
    fieldNullable    : false
    fieldWithIndexTag: false
    fieldWithTypeTag : true
    modelPkgName     : "sqlmodel"
#    outPath          : ""
#    tables           : ""
#    outFile          : ""
```

同级目录创建一个 sh 脚本（generate_sqlModel.sh）内容如下

```shell
gentool -c "./gensql.yml" -outPath "./dao/sql"
```

## grpc 配置

- 安装protoc编译器

> https://github.com/protocolbuffers/protobuf/releases/  
> 下载后解压到任意目录把`bin`里面的`protoc.exe`复制到`%GOPATH%/bin`里面，并配置`PATH`环境变量，确保 protoc 可以正常执行

- 安装相关模块

> go install google.golang.org/protobuf/proto  
> go install google.golang.org/protobuf/cmd/protoc-gen-go@latest  
> go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest  
