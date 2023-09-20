#!/usr/bin/env sh

# 读取远程文件内容并写入本地临时文件中
curl -s "http://101.43.0.205:6001/configFile?fileName=gensql.yml" -o "gensql.yml"

# 生成 sqlmodel 文件
gentool -c "gensql.yml" -outPath "./dao/sql"

# 删除本地临时文件
rm -f "gensql.yml"
