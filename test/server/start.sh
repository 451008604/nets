#!/bin/bash

# 默认配置
TCPPORT=17001
WSPORT=17002
HTTPPORT=17003
KCPPORT=17004

# 解析参数
while getopts "t:w:h:k:" opt; do
  case $opt in
    t) TCPPORT=$OPTARG ;;
    w) WSPORT=$OPTARG ;;
    h) HTTPPORT=$OPTARG ;;
    k) KCPPORT=$OPTARG ;;
    \?) echo "Usage: $0 [-t tcp_port] [-w ws_port] [-h http_port] [-k kcp_port]"
       exit 1 ;;
  esac
done

echo "Starting server on tcp=$TCPPORT ws=$WSPORT http=$HTTPPORT kcp=$KCPPORT..."

# 构建镜像
docker build -t nets-test-server -f Dockerfile .

# 运行容器
docker run --rm -it \
  -p ${TCPPORT}:${TCPPORT} \
  -p ${WSPORT}:${WSPORT} \
  -p ${HTTPPORT}:${HTTPPORT} \
  -p ${KCPPORT}:${KCPPORT} \
  --name nets-test-server \
  nets-test-server \
  -tcp ${TCPPORT} -ws ${WSPORT} -http ${HTTPPORT} -kcp ${KCPPORT}