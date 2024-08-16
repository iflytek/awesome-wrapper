#!/usr/bin/env bash
./client -g 1 -c 1 -tm 1000 -mode 1 -retry 3

#外挂配置文件，本地模式启动

#启动
#docker run -it --rm --net=host --workdir=/xsf/rate/client -v /home/sqjian/client/log:/xsf/rate/client/log -v /home/sqjian/client/client.toml:/xsf/rate/client/client.toml 172.16.59.153/aiaas/xsf:test ./client -g 1 -c 1 -tm 1000 -mode 0 -retry 3