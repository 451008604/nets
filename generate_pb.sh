#! /bin/bash
#--go_opt=paths=source_relative 使用相对路径生成文件
# shellcheck disable=SC2035
cd  ./proto/src &&
	protoc --go_out=../bin --go_opt=paths=source_relative \
		--go-grpc_out=../bin --go_opt=paths=source_relative \
		*.proto &&
	cd ../..