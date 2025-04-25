#!/bin/bash

# wrapper.go所在目录
workdir=$(pwd)

md5sum libwrapper.so
docker run -itd --name aiges-build-299 -v "${workdir}":/home/AIGES/src/wrapper artifacts.iflytek.com/docker-private/aipaas/aiges-build:2.9.9 bash
docker exec -it aiges-build-299 bash -c "mkdir -p /home/AIGES/src/vendor/github.com/mozillazg"
docker exec -it aiges-build-299 bash -c "cp -r /home/AIGES/src/wrapper/go-pinyin /home/AIGES/src/vendor/github.com/mozillazg"
docker exec -it aiges-build-299 bash ./build.wrapper.sh
docker cp aiges-build-299:/home/AIGES/bin/libwrapper.so "$workdir"
md5sum libwrapper.so
docker rm -f aiges-build-299