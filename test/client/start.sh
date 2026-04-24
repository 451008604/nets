#!/bin/bash

# 默认配置
SERVER="127.0.0.1"
TCPPORT=17001
WSPORT=17002
HTTPPORT=17003
KCPPORT=17004
PROTO="tcp"
CONN=10
INTERVAL=1000
DURATION=0
INSTANCES=1

# 解析参数
while getopts "s:t:w:h:k:p:c:i:d:n:" opt; do
  case $opt in
    s) SERVER=$OPTARG ;;
    t) TCPPORT=$OPTARG ;;
    w) WSPORT=$OPTARG ;;
    h) HTTPPORT=$OPTARG ;;
    k) KCPPORT=$OPTARG ;;
    p) PROTO=$OPTARG ;;
    c) CONN=$OPTARG ;;
    i) INTERVAL=$OPTARG ;;
    d) DURATION=$OPTARG ;;
    n) INSTANCES=$OPTARG ;;
    \?) echo "Usage: $0 -s server_ip [-t tcp_port] [-w ws_port] [-h http_port] [-k kcp_port] [-p proto] [-c conn_num] [-i interval_ms] [-d duration_s] [-n instances]"
       echo "  -n instances: number of containers to run (default: 1)"
       exit 1 ;;
  esac
done

echo "Starting ${INSTANCES} client container(s)..."
echo "  server: ${SERVER}"
echo "  proto: ${PROTO}"
echo "  conn: ${CONN}"
echo "  interval: ${INTERVAL}ms"

# 确定端口
case $PROTO in
  tcp)  PORT=$TCPPORT ;;
  ws)   PORT=$WSPORT ;;
  http) PORT=$HTTPPORT ;;
  kcp)  PORT=$KCPPORT ;;
  *)    PORT=$TCPPORT ;;
esac

# 构建镜像
docker build -t nets-test-client -f Dockerfile .

# 启动多个容器
for i in $(seq 1 $INSTANCES); do
  CONN_PER=$((CONN / INSTANCES))
  if [ $i -eq $INSTANCES ]; then
    CONN_PER=$((CONN - CONN_PER * (INSTANCES - 1)))
  fi
  
  echo "Starting instance $i with $CONN_PER connections..."
  docker run -d --rm \
    --name nets-test-client-$i \
    nets-test-client \
    -server ${SERVER} -${PROTO} ${PORT} -conn ${CONN_PER} -interval ${INTERVAL} -duration ${DURATION} &
done

echo "${INSTANCES} client container(s) started"