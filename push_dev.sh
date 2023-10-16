#!/usr/bin/env sh

tag_name="dev"

echo "=====================是否删除本地镜像====================="
echo "默认：删除 / 非空：不删除"
read delect

echo "======================上传镜像=========================="
sed -i 's/:latest/:'"$tag_name"'/g' ./docker-compose.yml
grep "image:" ./docker-compose.yml

# 开始编译镜像
docker compose build
docker compose push

if [ -z "$delect" ]; then
  docker compose down
  removeImages=$(docker images -a --format "table {{.Repository}}:{{.Tag}}" | grep "451008604/game" | grep $tag_name)
  docker rmi "$removeImages"
fi

echo "=====================镜像标签恢复默认====================="
sed -i 's/:'"$tag_name"'/:latest/g' ./docker-compose.yml
grep "image:" ./docker-compose.yml
