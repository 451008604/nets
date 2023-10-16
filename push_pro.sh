#!/usr/bin/env sh

tag_name=$(date +%Y%m%d)

image_name=ccr.ccs.tencentyun.com/451008604/game:$tag_name
# 对 dev 版本进行改名
docker tag ccr.ccs.tencentyun.com/451008604/game:dev "$image_name"
docker push "$image_name"

# 删除线上标签镜像
removeImages=$(docker images -a --format "table {{.Repository}}:{{.Tag}}" | grep "451008604/game" | grep "$tag_name")
docker rmi "$removeImages"
