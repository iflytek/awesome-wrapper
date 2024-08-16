### 运营管控之鉴权/并发上报SDK

---

![](./doc/logo.jpg)

鉴权SDK提供接入鉴权系统的系列接口，配合上报SDK，可实现

- 业务/引擎用量控制
- 业务/引擎可使用时间段控制
- 业务日流控
- 业务/引擎日试用量控制
- 业务QPS控制
- 业务并发控制
- 设备/用户量控



#### feature

- 支持配置中心/服务发现
- 支持受限资源模糊匹配
- 支持鉴权选项可配置，热加载



#### 接口说明

**鉴权SDK**

```go
/**
	鉴权接口，返回入参权限情况
	in:
		@appid 应用ID，如测试APPID 4CC5779A
		@uid   用户ID，可复用为设备ID
		@channel 渠道名称,如合成渠道tts，听写渠道iat
		@funcs 功能名称，如合成业务发音人vcn.xiaoyan
	out:
		@authInfo 无权限的参数列表及对应错误码,map中key为无权限的参数值，value为错误码如：
			k:v => vcn.xiaoyan:11200 表示vcn.xiaoyan发音人未授权
		@logInfo 鉴权流程示意字符串，建议集成方打印至日志中，便于排查权限相关问题
		@err 鉴权系统内部问题，当出现err不为空时，建议业务将此次鉴权请求看作有权限，避免因鉴权系统问题导致的业务不可用
*/
func Check(appid string, uid string, channel string, funcs []string) (authInfo map[string]string, logInfo string, err error)

/**
	多租户鉴权接口，返回入参权限情况
	in:
		@appid 应用ID，如测试APPID 4CC5779A
		@uid   用户ID，可复用为设备ID
		@cloudid 云编号，如公有云为0
		@composeid 组合模型编号
		@serviceid 服务ID
		@funcs 功能名称，如合成业务发音人vcn.xiaoyan
	out:
		@authInfo 无权限参数列表及对应错误码,map中key为无权限的参数值，value为错误码如：
			k:v => vcn.xiaoyan:11200 表示vcn.xiaoyan发音人未授权
		@logInfo 鉴权流程示意字符串，建议集成方打印至日志中，便于排查权限相关问题
		@err 鉴权系统内部问题，当出现err不为空时，建议业务将此次鉴权请求看作有权限，避免因鉴权系统问题导致的业务不可用
*/
func CheckMultiTenancy(appid , uid , cloudid , composeid , serviceid string, funcs []string) (authInfo map[string]string, logInfo string, err error)

/**
	获取应用配置的 limit 信息
	in:
		@appid 应用名
		@channel 服务名
		@funcs 功能名列表
	out:
		@map key: 功能， val: 对应 limit
		@err 内部错误信息
*/
func GetAcfLimits(appid, channel string, funcs []string) (map[string]string, error)

/**
	获取应用配置的 limit 信息
	in:
		@appid 应用名
		@cloudid 云编号，如公有云为0
		@composeid 组合模型编号
		@serviceid 服务ID
		@funcs 功能列表
	out:
		@map key: 功能， val: 对应 limit
		@err 内部错误信息
*/
func GetAcfLimitsMT(appid, cloudid, composeid, serviceid string, funcs []string) (map[string]string, error)

/**
	鉴权sdk初始化接口
	in:
		@channel 渠道名称,支持传入多个渠道，渠道参数用于获取该渠道下的受限制资源；该参数为正则匹配表达式
	out:
		@err 初始化错误信息
*/
func Init(channel []string) (err error)

/**
	设置配置中心地址
*/
func SetCompanionUrl(url string) 

/**
	设置服务所在项目名称
*/
func SetProjectName(projectName string) *InitOption

/**
	设置服务所在分组名称
*/
func SetGroup(group string) *InitOption

/**
	设置服务名称
*/
func SetServiceName(service string) *InitOption

/**
	设置是否使用远端配置
	in:
		@m 是否使用远端配置开关，m=1为使用远端配置，m=0为不适用远端配置
*/
func SetCfgMode(m int) *InitOption

```



**并发值上报SDK**

```go
/**
	上报并发数值
	in:
		@concInfo 并发数据，key表示appid，value表示该appid对应的并发值
	out:
		@err 上报接口返回的错误信息
*/
func Report(concInfo map[string]uint) (err error)

/**
	指定远端地址上报并发值
	in:
		@concInfo 并发数据，key表示appid，value表示该appid对应的并发值
		@addr 收集侧地址
*/
func ReportWithAddr(concInfo map[string]uint, addr string) (err error)

/**
	上报SDK初始化函数
	in:
		@channel 渠道名称
		@addr 当前上报服务所在地址
*/
func Init(channel string, addr string) (err error) 

/**
	sdk销毁函数
*/
func Fini()

/**
	设置配置中心地址
*/
func SetCompanionUrl(url string) 

/**
	设置服务所在项目名称
*/
func SetProjectName(projectName string) *InitOption

/**
	设置服务所在分组名称
*/
func SetGroup(group string) *InitOption

/**
	设置服务名称
*/
func SetServiceName(service string) *InitOption

/**
	设置是否使用远端配置
	in:
		@m 是否使用远端配置开关，m=1为使用远端配置，m=0为不适用远端配置
*/
func SetCfgMode(m int) *InitOption

```



#### 错误码

```shell
11200 未授权或权限到期
11201 用量不足
11202 QPS超过限定阈值
11203 并发值超过限定阈值
```



#### 配置文件

**鉴权SDK配置文件 janus-client.toml**

```toml
[janus-check-func]
conn-timeout = 20
conn-pool-size = 12         #rpc连接池数量。缺省4
lb-mode= 3  #0禁用lb,2使用lb。缺省0
lb-retry = 1
timeout = 50 #ms
#CtrlDayFlow CtrlTimeFlow//时授 CtrlCountFlow//量授  CtrlUserFlow//用户级 CtrlFreeFlow//免费次数 CtrlSecFlow//秒级流控 CtrlConcFlow //并发 
#z左边是最低位，右边是最高位  1 11111
check_option = 125
server_name = "janus"

[janus-acf-limit]
conn-timeout = 1000
lb-mode= 0
lb-retry = 2
timeout = 1000 #ms
updatetime = 5000  #ms
server_name = "janus"

[janus-limit-func]
conn-timeout = 1000
lb-mode= 0
lb-retry = 2
timeout = 3000 #ms
update_time = 60000  #ms
server_name = "janus"

[trace]
host="127.0.0.1"
port="4545"
able = -1
dump = 0
bcluster = "5s"
idc = "dz"

[log]
level = "error" #缺省warn
file = "/log/server/janus-client.log" #缺省xsfs.log
#日志文件的大小，单位MB
size = 100 #缺省10
#日志文件的备份数量
count = 10 #缺省10
#日志文件的有效期，单位Day
die = 3 #缺省10
#缓存大小，单位条数,超过会丢弃
cache = 100000 #缺省-1，代表不丢数据，堆积到内存中
#批处理大小，单位条数，一次写入条数（触发写事件的条数）
batch = 160#缺省16*1024

```

**计量SDK配置文件**

```toml
[janus-report]
conn-timeout = 1000
lb-mode= 3
timeout = 500
conn-pool-size = 12
lb-retry = 3

[trace]
host="127.0.0.1"
port="4545"
able = -1
dump = 0
bcluster = "5s"
idc = "dz"

[log]
level = "error" #缺省warn
file = "./log/xsfs-report.log" #缺省xsfs.log
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
async = 0 #缺省异步
#是否添加调用行信息
caller = 1 #缺省0
wash = 60 #写入磁盘的缺省时间

```

