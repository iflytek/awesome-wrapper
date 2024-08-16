package lb_client

import (
	"testing"
	"time"
)

func TestLbUtil(t *testing.T) {
	var (
		lbStrategy     = 0
		zkList         = []string{"192.168.86.60:2191", "192.168.86.60:2192", "192.168.86.60:2193"}
		root           = ""
		routerType     = "iat"
		subRouterTypes = []string{"iat_gray", "iat_hefei"}
		redieHost      = "192.168.86.60:6379"
		redisPasswd    = ""
		maxActive      = 100
		maxIdle        = 50
		db             = 0
		idleTimeOut    = time.Second * 1000
	)

	lu := &LbUtil{}
	//初始化连接
	err := lu.Init(WithLbStrategy(lbStrategy), WithZkList(zkList), WithRoot(root), WithRouterType(routerType), WithSubRouterTypes(subRouterTypes), WithRedisHost(redieHost),
		WithRedisPassword(redisPasswd), WithRedisDb(db), WithRedisMaxActive(maxActive), WithRedisMaxIdle(maxIdle), WithRedisIdleTimeOut(idleTimeOut))
	if err != nil {
		t.Errorf("lu_util Init error:%s", err.Error())
		return
	}

	//创建存活节点
	tempNode := "192.168.86.60:9090"
	for _, subRouterType := range lu.SubRouterTypes {
		if err = lu.createAliveNode(tempNode, subRouterType, nil); err != nil {
			t.Errorf("create temprory node:%s error:%s", tempNode, err.Error())
			return
		}
	}

	//删除存活节点
	for _, subRouterType := range lu.SubRouterTypes {
		if err = lu.deleteAliveNode(tempNode, subRouterType); err != nil {
			t.Errorf("delete temprory node:%s error:%s", tempNode, err.Error())
			return
		}
	}
}
