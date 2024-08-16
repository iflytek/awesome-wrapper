# ※ release 1.4.66

- trace support opt -1,indicates no modification of trace 
- fix rrlb bug(Local polling non-single hosts always request to a single node)

# ※ release 1.4.65

- add extension system information collection api
- distinguish load balancing errors
- redirect the signal when the registration list is empty
- support direct eng ip
    - use directEngIp param


# ※ release 1.4.64

- fix address nil bugs in conhash

# ※ release 1.4.63

- add EBADADDR & EBADHANDLE
```
EBADADDR      = errors.New("ContinueCall: invalid address")
EBADHANDLE    = errors.New("ContinueCall: invalid handle")
```
- client export max receive param
- req export Size()
- res export Size()
  
- 消息大小相关错误：

**server：**

```
※ more than maxreceive:
  code = ResourceExhausted desc = grpc: received message larger than max (xxx vs. xxx)
※ more than maxsend:
  code = ResourceExhausted desc = grpc: trying to send message larger than max (xxx vs. xxx)
```

**client ：**

```
※ more than maxreceive:
  code = ResourceExhausted desc = grpc: received message larger than max (xxx vs. xxx)
※ send unlimited
```

	 

# ※ release 1.4.62

- remove finder singleton

# ※ release 1.4.61

- fix Hidden danger（"req.SetParam"、getRaw）

# ※ release 1.4.60

- priority use local ip

# ※ release 1.4.59

- fix a spelling mistake
- integrated new trace v1.1.0 version
- fix many connections may occur at startup

# ※ release 1.4.58

- fix connection pool bug
- support call GracefulStop from outside

# ※ release 1.4.57

- fix tmpNode bug

# ※ release 1.4.56

- catch SIGTERM, SIGINT and ignore SIGPIPE

# ※ release 1.4.55 ##

- support GetRawCfg

# ※ release 1.4.53 ##

### xsf ###
- 健康探针
  - 支持LIVE_PORT、LIVE_PROC两种方式指定服务进程
  - 引擎成功上报和health接口均正常时，探测通过
- 适配live，增加引擎上报相关标志量
- finder-v2.0.15
  - 异常场景处理逻辑更新，
  - 日志修改为自建，不影响业务使用go自带log
  
### livenss-probes（健康探针） ###

**功能**
- 探测服务是否正常启动
- 探测服务的工作状态是否正常

**集成说明**

- 基于定制镜像，目前可选镜像如下（ps：有其它需求请联系我定制）
	- 172.16.59.153/aiaas/ubuntu:14.04.probes
		- Ubuntu 14.04.5 LTS
	- 172.16.59.153/aiaas/ubuntu:14.04.gcc.gdb.golang.probes
		- Ubuntu 14.04.5 LTS
		- go version go1.11.2 linux/amd64
		- gcc version 4.8.4 (Ubuntu 4.8.4-2ubuntu1~14.04.4)
		- GNU gdb (Ubuntu 7.7.1-0ubuntu5~14.04.3) 7.7.1


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

# ※ release 1.4.52 ##

- lb上报策略调整
- 修复本地启动时服务注册异常
- 集成finder 2.0.14
	- 常规优化
- 集成trace 1.0.5
	- add full spanId into record header; 
	- add batch size into record header with format #size; 
	- add catch - error stack log, more catch in serialize;

## ※ release 1.4.51 ##

- 适配个性化通知机制调整
	- 个性化请求从框架层面调整客户端请求，增加tmp标志
- 日志支持 none level，停止输出日志
- 日志Printf，适配新版本finder
- 更新finder
	- companion和zk同时崩溃。。再重新换zk的时候，服务发现功能不正常    
	- 控制台刷log  
	- 只获取配置文件接口崩溃


## ※ release 1.4.50 ##

- fix trace port empty
- remove console log info
- remove service check when continue call 
- 错误码映射

```
INVAILDPARAM  int32 = 10106 映射为 10139
INVAILDHANDLE int32 = 10109 映射为 10140
INVAILDDATA   int32 = 10108 映射为 10141
NOUSEFULCONN  int32 = 10200 映射为 10221
NETWORKEXCEPT int32 = 10201 映射为 10222 //网络异常
INVAILDLB     int32 = 10202 映射为 10223 //lb找不到有效节点
INVAILDLBSRV  int32 = 10203 映射为 10224 //找不到lb节点
INVAILDSRV    int32 = 10204 映射为 10224 //找不到业务节点
```

## ※ release 1.4.49 ##

- 集成 flange 1.0.4（性能优化版本）

## ※ release 1.4.48 ##

- sid 2.0 规则调整！！！（必须更新）
- 本地负载增加 hash 策略
- 增加 getBool 方法（甲方需求）
- 集成 new finder 2.0.12（修复崩溃 - companion崩溃的时候，zk有通知到达的时候崩溃）
- cmdServer 支援原始获取 pprof 数据
- 集成 flange 0.2.16,（修复ip、port 出现零值的情况）


## ※ Developer's Guide ##

**Ⅰ、references**

- git.xfyun.cn/AIaaS/hermes/src/master/doc/hermes_guide.md


## ※ FAQ ##

**Ⅰ、示例编译**

- 下载
```
go get git.xfyun.cn/AIaaS/xsf
```
> 忽略package git.xfyun.cn/AIaaS/xsf: no Go files。。。

- 编译
```
cd git.xfyun.cn/AIaaS/xsf/src/master/example/server/server/
./build.sh
 结果存在于当前目录中
```
- 参数优先级顺序
```
1、代码内参数
2、配置文件参数
3、命令行参数
```

- grpc version
```
1.9.1
```

- node
```
1、减少序列化过程中的反射操作（开源gogo实现）
2、增加连接池
3、合并小函数，减少函数调用
4、合并小对象同时进行对象复用，降低gc压力
5、goroutine之间减少关联，彼此冲突执行
6、减少取消timer（减少gc压力），使用channel通知
7、减少锁粒度，使用unsafe转换类型

```

- 常用镜像
```
172.16.59.153/aiaas/ubuntu-prometheus:14.04
172.16.59.153/aiaas/ubuntu:14.04.gcc.golang
172.16.59.153/aiaas/ubuntu:14.04
```