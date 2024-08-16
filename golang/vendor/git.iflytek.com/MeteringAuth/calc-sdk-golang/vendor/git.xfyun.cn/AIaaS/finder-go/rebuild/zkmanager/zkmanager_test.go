package zkmanager

import (
	"fmt"
	"github.com/cooleric/go-zookeeper/zk"
	"testing"
	"time"
)

func Test_zkManager_GetPath(t *testing.T) {
	conn,_,err:=zk.Connect([]string{"10.1.87.66:5181","10.1.87.67:5181","10.1.87.68:5181","10.1.87.69:5181","10.1.86.70:5181"},2*time.Second)
	ss,err:=conn.Create("/test",[]byte("hello"),2,zk.WorldACL(zk.PermAll))
	ss,err =conn.Create("/test",[]byte("hello"),2,zk.WorldACL(zk.PermAll))
	fmt.Println("create",ss,err)
	_,err=conn.Set("/test0000000028",[]byte("hello"),0)
	fmt.Println("set rr",err)
	data,_,err:=conn.Get("/test0000000028")
	fmt.Println(err,string(data))
}