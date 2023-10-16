FROM golang:1.21-alpine3.17 AS builder-base

ENV build_work_dir /go/src
ENV GOPROXY https://goproxy.cn
WORKDIR $build_work_dir

#构建运行环境
COPY ./go.mod $build_work_dir/
COPY ./go.sum $build_work_dir/
RUN go mod download

#构建可执行文件
COPY ./ $build_work_dir/
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux \
    && go build -ldflags '-w -s' -o $build_work_dir/main $build_work_dir/

#第二阶段压缩镜像体积
FROM 451008604/alpine:3.17

ENV running_work_dir /app
WORKDIR $running_work_dir

#拷贝运行所需文件
COPY --from=builder-base /go/src/main $running_work_dir/
COPY --from=builder-base /go/src/config/jsons $running_work_dir/config/jsons

ENTRYPOINT ["./main"]
