#!/usr/bin/env sh

# --go_opt=paths=source_relative 使用相对路径生成文件
# shellcheck disable=SC2035
cd ./proto/src || exit

# 生产 pb.go 文件
protoc --go_out=../bin --go_opt=paths=source_relative --go-grpc_out=../bin --go_opt=paths=source_relative *.proto

cd ../bin || exit

# 使用 sed 指令删除所有 'omitempty'
OS=$(uname)
if [ "$OS" = "Darwin" ]; then
  sed -i '' "s/,omitempty//g" *.pb.go
else
  sed -i "s/,omitempty//g" *.pb.go
fi

cd ../..
