#!/bin/bash

os_type=$(uname)

docker run -itd --privileged --name aiges-build artifacts.iflytek.com/docker-private/aipaas/aiges-build:v3.0.0 bash

if [[ "$os_type" == "Linux" ]]; then
  md5sum libwrapper.so
elif [[ "$os_type" == "Darwin" ]]; then
  md5 libwrapper.so
fi

docker exec -it aiges-build bash -c "rm -rf /home/AIGES/vendor"
docker cp go.mod aiges-build:/home/AIGES
docker cp go.sum aiges-build:/home/AIGES
docker cp vendor aiges-build:/home/AIGES/vendor
docker exec -it aiges-build bash -c 'mkdir -p /home/AIGES/wrapper'
docker cp wrapper.go aiges-build:/home/AIGES/wrapper
docker exec -it aiges-build bash ./build.wrapper.sh
docker cp aiges-build:/home/AIGES/bin/libwrapper.so $(pwd)

if [[ "$os_type" == "Linux" ]]; then
  md5sum libwrapper.so
elif [[ "$os_type" == "Darwin" ]]; then
  md5 libwrapper.so
fi

#docker rm -f aiges-build