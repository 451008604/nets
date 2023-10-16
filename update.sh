#!/usr/bin/env sh

# 修改镜像标签
echo "默认：当前compose镜像标签 / 非空：当前年月日"
read -r input_tag
tag_name=$(date +%Y%m%d)
source_content=$(grep "" ./docker-compose.yml)
echo "====================修改镜像标签===================="
if [ -n "$input_tag" ]; then
	sed -i 's|game:.*|game:'"$tag_name"'|' ./docker-compose.yml
fi
grep "image:" ./docker-compose.yml

# 启动镜像
echo "======================启动镜像======================"
docker compose logs
docker compose pull
docker compose up -d
docker compose logs

# 恢复镜像标签
echo "==================镜像标签恢复默认=================="
if [ -n "$input_tag" ]; then
	echo "$source_content" >./docker-compose.yml
fi
grep "image:" ./docker-compose.yml
