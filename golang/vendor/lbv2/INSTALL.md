## 简介 ##

- 使用zk和redis来实现服务发现和负载均衡

## 安装 ##

- 克隆代码    
```
    git clone http://172.16.59.151:80/AIaaS/lb.git
    拉取相应的tag即可编译
```

## 编译 ##

### docker ###

- 直接从镜像仓库拉取相应版本的镜像(docker pull 172.16.59.153/aiaas/lb:2.0.2)

### 本地编译 ###

- **golang版本要求1.9以上**

- 设置好相应的GOPATH(如:export GOPATH=/data/go)
- 在GOPATH下创建文件夹src
- 拷贝lb源码到src下,并进入lb目录
- 此时目录结构类似于:    
    |——/data/go/src/lb
- 在lb目录下,执行cmd：go build -v -o lb
- 拷贝可执行文件./lb 到bin目录下
- 参阅:gobuild.sh

## 运行 ##

### cmd ###

- 进入bin目录     
    |——lb(可执行文件)   
    |——lb.toml-本地的配置文件(相关配置的意义都有注释)   
- 启动

```
1、配置中心方式启动时
cmd： ./lb -m 1 -c lb.toml -p AIaaS -s iat_lb -u http://10.1.86.223:9080 -g aitest   
2、本地配置方式启动时
cmd: ./lb -m 0 -c lb.toml  -s iat_lb 

注释:
    -m 制定运行方式(0-本地配置,1-配置中心)    
    -c 配置文件名(Native模式时，用本地的配置文件，Center模式时使用配置中心的配置文件)    
    -p 项目名    
    -s 启动的服务名(注意:配置文件，配置中心的服务名需一致)  
    -u 配置中心地址    
    -g 配置项目组   
```

###  docker启动  ###

```
docker run -itd --net=host -v /data/lb/findercache/:/lb/findercache -v /data/lb/log:/lb/log/  -v /data/lb/metric:/lb/metric
-v /data/trace:/lb/trace 172.16.59.153/aiaas/lb:2.0.2 ./lb -m 1 -c  lb.toml -p AIaaS -s iat_lb -u http://10.1.86.223:9080 -g aitest 
注释：
    1、容器的目录结构：
        |——/lb
        |——/lb/log(存储日志文件)
        |——/lb/lb(可执行文件)
        |——/lb/lb.toml(配置文件，本地启动时需要挂载，配置中心启动时不需要)
        |——/lb/findercache(配置中心缓存目录) 
        |——/lb/metric(sonar本地缓存，配置中开启sonar本地磁盘时)
        |——/lb/trace(trace本地缓存，配置中开启trace本地磁盘时)
    2、挂载目录根据具体来指定
    3、运行的镜像版本号根据发布的版本号做修改 

```

## 配置项 ###

```
    #服务自身的配置
    #注意此section名需对应bootConfig中的service
    title = "lbs Service Configuration"
    [iat_lb]
    #host="127.0.0.1"                 #若host为空，则取netcard对应的ip，若二者均为空，则取hostname对应的ip
    #netcard = "eth0"
    port = 9095                       #指定端口
    reuseport = 0                     #缺省0
    cmdserver = 1                     #缺省0
    finder = 1                        #使用服务发现,缺省0
    debug = 0                         #缺省0

    #trace日志所用
    [trace]
    host = "172.16.51.3"              #trace收集服务的地址,缺省127.0.0.1
    port = 4546                       #trance的端口号,缺省4545
    backend = 1                       #trace服务的协程数,缺省4
    deliver = 1                       #是否将日志写入到远端,缺省1
    dump = 1                          #是否将日志 落入磁盘,缺省0
    able = 0                          #是否禁用trace,缺省1

    [log]
    level = "debug"                  #日志文件类型,缺省warn
    file = "log/lb.log"              #日志文件名,缺省xsfs.log
    size = 3                         #日志文件的大小,单位MB,缺省10
    count = 3                        #日志文件的备份数量,缺省10
    die = 3                          #日志文件的有效期,单位Day,缺省10
    cache = -1                       #缓存大小,单位条数,超过会丢弃,(缺省-1，代表不丢数据，堆积到内存)
    batch = 1600                     #批处理大小,单位条数,一次写入条数（触发写事件的条数）
    async = 1                        #异步日志,缺省异步
    caller = 1                       #是否添加调用行信息,缺省0

    [lb]#loadReporter负载上报相关配置，其中strategy为上报的策略，其余字段为用户需要自定义上报的数据
    able =0

    [fc]#flowControl 包括sessionManager和qpsLimiter
    #限流器的类型，若所填值非sessionManager和qpsLimiter或者没填，那么限流器不会初始化
    able = 1 #缺省为0
    router = "qpsLimiter"            #路由字段，可选项为sessionManager和qpsLimiter
    max = 10000                      #会话模式时代表最大的授权量，非会话模式代表间隔时间里的最大请求数
    ttl = 3                          #会话模式代表会话的超时时间，非会话模式代表有效期（间隔时间）
    best = 10                        #最佳授权数
    roll = 3                         #sessionManager内部遍历超时session的时间间隔  缺省1s
    report =1                        #上报时间间隔 缺省1s


    #zk配置
    [lb_zk]
    #zk列表
    zkList= ["192.168.86.60:2190", 
            "192.168.86.60:2191", 
            "192.168.86.60:2192",
                ]
    root = "/"                       #根目录
    routerType = "iat"               #路由类型(指定大的业务类型)
    updateCacheTime = 100            #更新本地缓存的时间(单位毫秒)

    #redis相关配置
    [lb_redis]
    redisHost = "192.168.86.60:6379" #redis主机
    redisPassword = ""               #redis密码
    db = 0                           #redis存储的数据库表
    maxActive = 300                  #redis最大连接数
    maxIdle = 100                    #redis最大空闲连接数
    idleTimeOut = 3600               #redis的空闲连接超时

    #sonar日志所用
    [sonar]
    host = "10.1.86.60"              #trace收集服务的地,缺省127.0.0.1
    port = 4546                      #trace收集服务的端口,缺省4545
    backend = 1                      #trace服务的协程数,缺省4
    deliver = 1                      #是否将日志写入到远端,缺省1
    dump = 1                         #是否将日志写入磁盘,缺省0
    able = 0                         #是否禁用trace,缺省1
    rate = 5000                      #上报频率,单位毫秒
    ds = "vagus"                     #缺省vagus

```

## 工具 ##

- client

```
    1、cd  lb/client/lb_client
    2、编译： go build 
    3、运行 ./lb_client -config lb.toml(lb.toml中参数都有具体的解释)
```

- 模拟工具(注册和更新实例数)

```
    1、cd  lb/tool/lb_client_simulate/lb_client_update
    2、编译： go build 
    3、运行 ./lb_client_update -cfg lb_client_update.toml(lb_client_update.toml中参数都有具体的解释)
```

- 模拟工具(查询实例数)

```
    1、cd  lb/tool/lb_client_simulate/lb_client_query
    2、编译： go build 
    3、运行 ./lb_client_query -cfg lb_client_query.toml(lb_client_query.toml中参数都有具体的解释)
```