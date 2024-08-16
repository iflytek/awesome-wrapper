## ※ release xsf v1.7.2 ##

```
1、支持多lb多ent上报，上报格式要求为： `^(([\w.\-|]+,){5}([\w.\-|]+))((;([\w.\-|]+,){5}([\w.\-|]+)))*$`
```

## ※ release xsf v1.6.6 ##

```
1、支持以冷备模式访问 lb 节点，eg: caller.WithLBParams("lb-lock", "iat", map[string]string{"sort": "true"})
```

## ※ release xsf v1.6.3 ##

```
NETWORKEXCEPT = 10222 //网络异常
INVAILDLB     = 10223 //lb找不到有效节点
INVAILDLBSRV  = 10224 //找不到lb节点
INVAILDSRV    = 10225 //找不到业务节点
INVAILDRMLB   = 10226 //请求lb失败
INVAILDRENG   = 10227 //请求ent失败
```

```
https://git.iflytek.com/rdg_ai_services/redisgo.git
https://git.iflytek.com/rdg_ai_services/lbclientpb.git
https://git.iflytek.com/rdg_ai_services/uuid.git
https://git.iflytek.com/rdg_ai_services/thrift.git
https://git.iflytek.com/rdg_ai_services/sonar.git
https://git.iflytek.com/rdg_ai_services/lumberjack-ccr.git
https://git.iflytek.com/rdg_ai_services/lb_client.git
https://git.iflytek.com/rdg_ai_services/flume
https://git.iflytek.com/rdg_ai_services/flange
```

## ※ release 1.6.2 ##

- update finder-go v1.0.4 to fix lb 10224
- fix signal handler bug

## ※ release 1.6.0 ##

- add config bvt.namespace

## ※ release 1.5.0 ##

- release v1.5.0

## ※ release 3.0.2 ##

- support to extract Windows local IP

## ※ release 3.0.1 ##

- move /xsf_status to xsf_status

## ※ release 3.0.0 ##

- Migration vendor to module

**Notes**

- 由于 go vendor 与 go module 使用逻辑上存在较大差异，为保证现有业务不受影响，故而升级 xsf 主线版本至 v3，注意调整导入路径为 git.iflytek.com/AIaaS/xsf/v3
- 由于 xsf 所依赖的部分包已无人维护，故将这部分依赖在 git.iflytek.com 上重建仓库并调整相关导入路径， 且因为 git.iflytek.com 是公司私有库，为避免 go.sum 校验，请设置 GOPRIVATE 变量
- 由于组织关闭了 git https 登陆，导致 go module 默认的拉取逻辑无法正常获取依赖，需对部分仓库的的拉取方式进行调整方可拉取，请参考如下命令

```bash
git config --global url."git@git.iflytek.com:AIaaS/finder-go.git".insteadOf "https://git.iflytek.com/AIaaS/finder-go.git"
git config --global url."git@git.iflytek.com:AIaaS/lumberjack-ccr.git".insteadOf "https://git.iflytek.com/rdg_ai_services/lumberjack-ccr.git"
git config --global url."git@git.iflytek.com:AIaaS/xsf.git".insteadOf "https://git.iflytek.com/AIaaS/xsf.git"
git config --global url."git@git.iflytek.com:sqjian/flange.git".insteadOf "https://git.iflytek.com/rdg_ai_services/flange.git"
git config --global url."git@git.iflytek.com:sqjian/sonar.git".insteadOf "https://git.iflytek.com/rdg_ai_services/sonar.git"
git config --global url."git@git.iflytek.com:sqjian/flume.git".insteadOf "https://git.iflytek.com/rdg_ai_services/flume.git"
git config --global url."git@git.iflytek.com:sqjian/thrift.git".insteadOf "https://git.iflytek.com/rdg_ai_services/thrift.git"
git config --global url."git@git.iflytek.com:sqjian/uuid.git".insteadOf "https://git.iflytek.com/rdg_ai_services/uuid.git"
git config --global url."git@git.iflytek.com:sqjian/lb_client.git".insteadOf "https://git.iflytek.com/rdg_ai_services/lb_client.git"
git config --global url."git@git.iflytek.com:sqjian/lbclientpb.git".insteadOf "https://git.iflytek.com/rdg_ai_services/lbclientpb.git"
git config --global url."git@git.iflytek.com:sqjian/redisgo.git".insteadOf "https://git.iflytek.com/rdg_ai_services/redisgo.git"
```

## ※ release 2.5.6 ##

- reset cloud from int to string

## ※ release 2.5.5 ##

- support cloud_id

eg:

```toml
[lbv2] #已做缺省处理,此section如不传缺省不启用
tm = 1000 #缺省1000，批量上报的超时时间，单位毫秒
backend = 100#上报的的协程数，缺省4
finderttl = 100 #更新本地地址的时间，通过访问服务发现实现，缺省一分钟
lbname = "lbv2"
apiversion = "1.0"
able = 1
cloud = 1
sub = "iat"
subsvc = "sms,sms-16k"
task = 10 #任务队列长度
```

## ※ release 2.5.4 ##

- update new finder-go by sjliu7
- support new lb option

```
HERMESLBPROJECT        = "lbproject"
HERMESLBGROUP          = "lbgroup"
```

## ※ release 2.5.3 ##

- fix from meta exception when in flexible scheduling

## ※ release 2.5.2 ##

- support built-in status

## ※ release 2.5.1 ##

- fix ase exceptional request body

## ※ release 2.5.0 ##

- add findermode to lbv2 section

## ※ release 2.4.0 ##

- multi-svc-featured conhash

## ※ release 2.3.0 ##

- 修正QPS线性增加问题
- client增加custom mode（本地配置+服务发现）

## ※ release 2.2.0 ##

- 服务端上报并发路数、QPS（监控用）

```
# HELP concurrent_statistics concurrent_statistics
# TYPE concurrent_statistics gauge
concurrent_statistics{cs="3s",idc="dx",name="xsf-server",sub="xsf"} 0
# HELP module_delay module_delay
# TYPE module_delay gauge
module_delay{cs="3s",idc="dx",name="xsf-server",sub="xsf",type="avg"} 0
module_delay{cs="3s",idc="dx",name="xsf-server",sub="xsf",type="max"} 0
module_delay{cs="3s",idc="dx",name="xsf-server",sub="xsf",type="min"} 0
module_delay{cs="3s",idc="dx",name="xsf-server",sub="xsf",type="qps"} 0
```

- notLicense可调整

```
func (s *SessionManager) IncrLicBy(n int32) {
	atomic.AddInt32(&s.NowLic, n)
}

func (s *SessionManager) DecrLicBy(n int32) {
	atomic.AddInt32(&s.NowLic, -n)
}
```

## ※ release 2.1.6 ##

- support flexible
- support multi-lbInstance

## ※ release 2.1.4 ##

- Optimize integration with BVT

## ※ release 2.1.2 ##

- support report failed nodes details to lb

## ※ release 2.1.0 ##

- support bvt result callback

## ※ release 2.0.4 ##

- improve metrics

## ※ release 2.0.3 ##

- support bvt platform

## ※ release 2.0.2 ##

- fix block when endpoint exception occurs

## ※ release 2.0.1 ##

- fix crash when broken network

## ※ release 2.0.0 ##

- move git.xfyun.cn to git.iflytek.com
- support client to report the exception node
- support slowly report the authorization number
- support hystrix
    - https://git.iflytek.com/AIaaS/metrics/blob/master/doc/hystrix开发提纲.md
- support metrics
    - https://git.iflytek.com/AIaaS/xrpc/blob/master/doc/server/metrics.md
- support new strategy
    - https://git.iflytek.com/AIaaS/xrpc/blob/master/doc/busin-lb/负载均衡（业务侧）.md
- rollback to old ver:1.1.0 from ver:1.1.1

## ※ release 1.4.72 ##

- add func NewSpanWithOpts(spanType SpanType, opts ...TraceOpt) *Span

## ※ release 1.4.71 ##

- fix crash in connPoolMeta.closeAll

## ※ release 1.4.70 ##

- new strategy busin-lb
- support adjusting lb reported parameters
- support to turn off information collection
- remove arguments n in unregisterAll
- fix error correction in modle_delay

# ※ release 1.4.69

- fix pickfirstBalancer: failed to NewSubConn: grpc: the client connection is closing
- update flange v1.1.1

# ※ release 1.4.68

- fix time error in cleaner
- both client and server support keepalive and keepalive-timeout

# ※ release 1.4.67

- metrics sdk add UnregisterAll
- optimize rrlb

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

​

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

- 常用镜像_v1

```
172.16.59.153/aiaas/ubuntu-prometheus:14.04
172.16.59.153/aiaas/ubuntu:14.04.gcc.golang
172.16.59.153/aiaas/ubuntu:14.04
```

- 常用镜像_v2
    - img_v1
        - 172.16.59.153/aiaas/ubuntu:14.04.gcc.golang
        - 172.16.59.153/aiaas/ubuntu:14.04.gcc_v4.8.4.golang_v1.11.2
    - img_v2
        - 172.16.59.153/aiaas/ubuntu:14.04
    - img_v3
        - 172.16.59.153/aiaas/ubuntu:14.04.gcc_v4.8.4
    - img_v4
        - 172.16.59.153/aiaas/ubuntu:14.04.gcc_v4.8.4.golang_v1.12.7
    - img_v5
        - 172.16.59.153/aiaas/ubuntu:14.04.gcc_v4.8.4.golang_v1.12.7.stack.20k
