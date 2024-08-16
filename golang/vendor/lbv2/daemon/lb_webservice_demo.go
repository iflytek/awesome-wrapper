package daemon

import (
	"fmt"
	"time"
)

func main() {
	WebServiceInst := WebService{}
	WebServiceInst.Init("http://172.16.154.235:8081/ws", "xfyun", "12345678", time.Second*3, `100IME`, "db-service-v3-3.0.0.1001", "bj", "ifly_cp_msp_balance", "seg_list")
	columnJson := map[string]string{
		"seg_id":    "seg_id",
		"type":      "type",
		"server_ip": "server_ip",
	}
	whereJson := map[string]string{
		"type": "vmail",
	}
	fmt.Println(WebServiceInst.GetList(columnJson, whereJson))
	fmt.Println(WebServiceInst.GetListNoWhereJson(columnJson))
	fmt.Println(WebServiceInst.Insert(map[string]string{"seg_id": "xxx", "type": "xxx", "server_ip": "x.x.x.x"}))
	fmt.Println(WebServiceInst.Insert(map[string]string{"seg_id": "xxx", "type": "xxx", "server_ip": "x.x.x.x"}))
	fmt.Println(WebServiceInst.Delete(map[string]string{"seg_id": "xxx", "type": "xxx", "server_ip": "x.x.x.x"}))
}

//output:
/*
{"ret":"0","result":[{"seg_id":"21","server_ip":"127.0.0.9","type":"vmail"}]} <nil>
{"ret":"0","result":[{"seg_id":"21","server_ip":"127.0.0.9","type":"vmail"},{"seg_id":"11","server_ip":"127.0.0.1","type":"sms"}]} <nil>
*/
