package lb_client

import (
	"fmt"
	"testing"
	"time"
)

func TestLbClient(t *testing.T) {
	var (
		lbStrategy     = 0                                                                          //负载策略(必传)
		zkList         = []string{"192.168.86.60:2191", "192.168.86.60:2192", "192.168.86.60:2193"} //zk列表(必传)
		root           = ""                                                                         //根目录
		routerType     = "iat"                                                                      //路由类型(如：iat)(必传)
		subRouterTypes = []string{"iat_gray", "iat_hefei"}                                          //子路由类型列表(如:["iat_gray","iat_hefei"])
		redieHost      = "192.168.86.60:6379"                                                       //redis主机(必传)
		redisPasswd    = ""                                                                         //redis密码
		maxActive      = 100                                                                        //redis最大连接数
		maxIdle        = 50                                                                         //redis最大空闲连接数
		db             = 0                                                                          //redis数据库
		idleTimeOut    = time.Second * 1000                                                         //redis空闲连接数超时时间
	)

	/*-----------------初始化----------------------*/
	var lc LbClienter = &LbClient{}
	err := lc.Init(WithLbStrategy(lbStrategy), WithZkList(zkList), WithRoot(root), WithRouterType(routerType), WithSubRouterTypes(subRouterTypes), WithRedisHost(redieHost),
		WithRedisPassword(redisPasswd), WithRedisDb(db), WithRedisMaxActive(maxActive), WithRedisMaxIdle(maxIdle), WithRedisIdleTimeOut(idleTimeOut))
	if err != nil {
		t.Errorf("init error:%s", err.Error())
		return
	}

	/*-----------------测试引擎上线----------------------*/
	var (
		svc       = "192.168.86.60:2181" //引擎节点
		totalInst = int32(200)           //总实例数
		idleInst  = int32(200)           //空闲实例数
		bestInst  = int32(150)           //最好的实例数
		param     = map[string]string{"param": "only_test"}
	)
	if err = lc.Login(svc, totalInst, idleInst, bestInst, param); err != nil {
		t.Errorf("svc:%s Login error:%s", svc, err.Error())
		return
	}
	fmt.Printf("svc:%s Login Success!\n", svc)
	time.Sleep(time.Second * 1)

	/*-----------------测试引擎更新数据----------------------*/
	totalInst = int32(300)
	idleInst = int32(100)
	bestInst = int32(200)
	err = lc.Upadate(totalInst, idleInst, bestInst)
	if err != nil {
		t.Errorf("svc:%s Upadate error:%s", svc, err.Error())
		return
	}
	fmt.Printf("svc:%s Upadate Success!\n", svc)
	time.Sleep(time.Second * 1)

	/*-----------------测试引擎下线----------------------*/
	err = lc.LoginOut()
	if err != nil {
		t.Errorf("svc:%s LoginOut error:%s", svc, err.Error())
		return
	}
	fmt.Printf("svc:%s LoginOut Success!\n", svc)
	time.Sleep(time.Second * 30)
}
