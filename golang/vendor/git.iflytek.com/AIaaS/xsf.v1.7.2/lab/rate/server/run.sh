#!/usr/bin/env bash

#./server -m 0

#docker run -it --rm --workdir=/xsf/rate/server 172.16.59.153/aiaas/xsf:test ./server -m 0

#外挂配置文件，本地模式启动

#启动
#docker run -it --rm --net=host --workdir=/xsf/rate/server -v /home/sqjian/server/server.toml:/xsf/rate/server/server.toml 172.16.59.153/aiaas/xsf:test ./server -m 0

#停止
#docker ps|grep xsf:test|awk '{print $1}'|xargs docker rm -f


#多实例启动脚本
#cnt=10
#
#for (( i=1; i <= $cnt; i++ ))
#do
#    echo -n "server instance NO.$i => "
#    docker run -it --rm --net=host --ulimit nofile=65535 --ulimit nproc=65535 --workdir=/xsf/rate/server -v /home/sqjian/server/server.toml:/xsf/rate/server/server.toml 172.16.59.153/aiaas/xsf:test ./server -m 0
#done