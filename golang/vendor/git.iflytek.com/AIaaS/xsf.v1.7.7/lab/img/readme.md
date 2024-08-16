## ※ img_v1 ##
- 172.16.59.153/aiaas/ubuntu:14.04.gcc.golang
  - Ubuntu 14.04.5 LTS
  - go version go1.11.2 linux/amd64
  - gcc version 4.8.4 (Ubuntu 4.8.4-2ubuntu1~14.04.4)
- 172.16.59.153/aiaas/ubuntu:14.04.gcc_v4.8.4.golang_v1.11.2

## ※ img_v2 ##
- 172.16.59.153/aiaas/ubuntu:14.04

## ※ img_v3 ##
- 172.16.59.153/aiaas/ubuntu:14.04.gcc_v4.8.4
- 172.16.59.153/aiaas/ubuntu:14.04.gcc_v4.8.4.gdb_v7.7.1

## ※ img_v4 ##
- 172.16.59.153/aiaas/ubuntu:14.04.gcc_v4.8.4.golang_v1.12.7
- 172.16.59.153/aiaas/ubuntu:14.04.gcc_v4.8.4.gdb_v7.7.1.golang_v1.12.7

## ※ img_v5 ##
- 172.16.59.153/aiaas/ubuntu:14.04.gcc_v4.8.4.golang_v1.12.7.stack.4k
- 172.16.59.153/aiaas/ubuntu:14.04.gcc_v4.8.4.golang_v1.12.7.stack.8k

## ※ img_v6

- 172.16.59.153/aiaas/ubuntu:14.04.probes

  - Ubuntu 14.04.5 LTS

- 172.16.59.153/aiaas/ubuntu:14.04.gcc_v4.8.4.gdb_v7.7.1.golang_v1.12.7.probes

  - Ubuntu 14.04.5 LTS

  - go version go1.12.7 linux/amd64

  - gcc version 4.8.4 (Ubuntu 4.8.4-2ubuntu1~14.04.4)

  - GNU gdb (Ubuntu 7.7.1-0ubuntu5~14.04.3) 7.7.1

## ※ img_grafana

- 172.16.59.153/aiaas/grafana/grafana
    

## ※ livenss-probes（健康探针） ##

**功能**
- 探测服务是否正常启动
- 探测服务的工作状态是否正常

**集成说明**

- 基于定制镜像，目前可选镜像：img_v6（ps：有其它需求请联系我定制）


- 手动集成
	- 拷贝live到镜像并添加至环境变量

- 启动容器过程添加 -e "LIVE_PROC=$(服务名)" 即可
	- 如：docker run -e "LIVE_PROC=$(服务名)" 172.16.59.153/aiaas/ubuntu:14.04.probes

**使用说明**

- 进入容器执行live
	- 成功
		- 终端显示：success
		- 退出码：0
	
	- 失败
		- 终端显示：failure
		- 退出码：1
- live -v 1可查看更为详细的信息
## ※ go _rebuild ##
- down and install official golang complier
	- go env
		- GOPATH="/root/go"
		- GOROOT="/usr/local/go"
- down and unpack gosrc
- export GOROOT_BOOTSTRAP=$GOROOT
- cd go/src & ./all.bash