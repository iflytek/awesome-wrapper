### 运营管控系统之计量SDK

---

该SDK提供用量统计功能，用量记入运营管控系统，配合鉴权服务可实现用量鉴权控制。

<img src="./doc/logo.png"/>

#### feature

- 异步计量接口
- 支持配置中心
- 多rmq地址配置
- 计量热开关
- 异常连接自修复（退避算法）

#### 接口说明

```go
/*
    初始化接口
    in:
 @url 配置中心地址
 @pro 配置中心项目名称
 @gro 配置中心分组名称
 @service 配置中心服务名称，集成该sdk的服务
 @version 配置中心服务版本号，集成该sdk的版本号
 @isNative 是否使用本地配置，true为使用本地配置
 @nativeLogPath 本地配置文件路径
 out:
  error 当初始化异常时返回error信息，正常时error为nil
 
*/
func Init(url, pro, gro, service, version string, isNative bool, nativeLogPath string) error

/*
 销毁接口
*/
func Fini()

/*
 appid计量接口
 in:
 @appid 应用ID,如测试APPID 4CC5779A
 @channel 渠道名称,如合成渠道tts，听写渠道iat
 @funcs 功能名称，如合成业务发音人vcn.xiaoyan
 @c 用量值
 out:
 @errorcode 返回错误码，正常时返回0
 @err 错误信息，正常时为nil
*/
func Calc(appid, channel, funcs string, c int64) (errorcode int, err error)

/*
 subID计量接口
 in:
 @appid 应用ID,如测试APPID 4CC5779A
 @subId 应用子ID,如uid（用户ID），did（设备ID）等
 @channel 渠道名称,如合成渠道tts，听写渠道iat
 @funcs 功能名称，如合成业务发音人vcn.xiaoyan
 @c 用量值
 out:
 @errorcode 返回错误码，正常时返回0
 @err 错误信息，正常时为nil
*/
func CalcWithSubId(appid, subId, channel, funcs string, c int64) (errorcode int, err error)

/*
 appid多租户计量接口
 in:
 @appid 应用ID,如测试APPID 4CC5779A
 @cloudid 云标识，如公有云id为0
 @composeid 组合能力id
 @service 服务ID
 @funcs 功能名称，如合成业务发音人vcn.xiaoyan
 @c 用量值
 out:
 @errorcode 返回错误码，正常时返回0
 @err 错误信息，正常时为nil
*/
func CalcMultiTenancy(appid, cloudid , composeid , serviceid , funcs string, c int64) (errorcode int, err error)

/*
 subID计量接口
 in:
 @appid 应用ID,如测试APPID 4CC5779A
 @subId 应用子ID,如uid（用户ID），did（设备ID）等
 @cloudid 云标识，如公有云id为0
 @composeid 组合能力id
 @service 服务ID
 @funcs 功能名称，如合成业务发音人vcn.xiaoyan
 @c 用量值
 out:
 @errorcode 返回错误码，正常时返回0
 @err 错误信息，正常时为nil
*/
func CalcWithSubIdMultiTenancy(appid, subId, cloudid , composeid , serviceid , funcs string, c int64) (errorcode int, err error)
```

#### 错误码

#### 配置文件

```toml
[common]
#use = "mq" 
use = "rpc"
queue_size = 10000

[log]#已做缺省处理
level = "debug" #缺省warn
file = "./log/calc-sdk.log" #缺省xsfs.log
#日志文件的大小，单位MB
size = 300 #缺省10
#日志文件的备份数量
count = 3 #缺省10
#日志文件的有效期，单位Day
die = 3 #缺省10
#缓存大小，单位条数,超过会丢弃
cache = 100000 #缺省-1，代表不丢数据，堆积到内存中
#批处理大小，单位条数，一次写入条数（触发写事件的条数）
batch = 160#缺省16*1024
#异步日志
async = false #缺省异步
#是否添加调用行信息
caller = true #缺省0
wash = 60 #写入磁盘的缺省时间

[rmq]
able = 1
producer_number = 2
hosts = ["10.1.87.18:10800"]
topic = "lic_lkc_bj"
# 消息队列服务连接超时时间
# millisecond
timeout = 500

[pulsar]
able = 1
idc="hf"
appids = ["testCalcSDK","*"]
topic = "persistent://aiaas/metering/isol"
endpoint = "pulsar://10.1.87.69:6650"

[sea-client]
conn-timeout = 100
conn-pool-size = 12         #rpc连接池数量。缺省4
lb-mode= 3  #0禁用lb,2使用lb。缺省0
lb-retry = 1
rpc_timeout = 300 #ms
#conn-rbuf =  1048576
#conn-wbuf = 1048576
finder = 0
#taddrs = "router@172.31.98.182:8098"

[trace]
able = -1

[sonar]
able = 0

[sdk]
conn-timeout = 100
conn-pool-size = 12         #rpc连接池数量。缺省4
lb-mode= 3  #0禁用lb,2使用lb。缺省0
lb-retry = 1
#conn-rbuf =  1048576
#conn-wbuf = 1048576
finder = 1

[metrics]
#参数齐则开启metrics
able = 1
idc = "hf"
sub = "calc"
cs = "1s"
timePerSlice = 1000 #滑动窗口bucket大小，单位毫秒
winSize = 10 #窗口大小
```

#### 基本原理与架构

**交互图**

![](./doc/交互图.png)

**热开关逻辑流程**

![](./doc/热开关逻辑路程图.png)

**sdk销毁流程**
Fini后的用户交互

![](./doc/user_interactive_after_fini.png)

队列内留存数据清理

![](./doc/clean_queue_data_after_fini.png)
