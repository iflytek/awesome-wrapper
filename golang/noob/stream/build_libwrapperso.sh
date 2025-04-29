#!/bin/bash

# 启用别名扩展
shopt -s expand_aliases

os_type=$(uname)
if [[ "$os_type" == "Darwin" ]]; then
  alias md5sum="md5"
fi

md5sum libwrapper.so
docker run -itd --name aiges-build-299 -v $(pwd):/home/AIGES/src/wrapper artifacts.iflytek.com/docker-private/aipaas/aiges-build:2.9.9 bash
docker exec -it aiges-build-299 bash -c "mkdir -p /home/AIGES/src/vendor/github.com/mozillazg"
docker exec -it aiges-build-299 bash -c "cp -r /home/AIGES/src/wrapper/go-pinyin /home/AIGES/src/vendor/github.com/mozillazg"
docker exec -it aiges-build-299 bash ./build.wrapper.sh
docker cp aiges-build-299:/home/AIGES/bin/libwrapper.so $(pwd)
md5sum libwrapper.so
docker rm -f aiges-build-299
