#!/bin/bash

# wrapper.go所在目录
workdir=$(pwd)

docker run -itd --name aiges-build-2914 -v "${workdir}":/home/AIGES/src/wrapper artifacts.iflytek.com/docker-private/aipaas/aiges-build:2.9.1.4 bash
docker exec -it aiges-build-2914 bash ./build.wrapper.sh
docker cp aiges-build-2914:/home/AIGES/bin/libwrapper.so "$workdir"
docker rm -f aiges-build-2914