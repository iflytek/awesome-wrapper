package main

import (
	"bytes"
	"fmt"
	"git.iflytek.com/AIaaS/xsf/client"
	"git.iflytek.com/AIaaS/xsf/lab/basic/client/internal/protocol"
	"git.iflytek.com/AIaaS/xsf/utils"
	"github.com/gogo/protobuf/proto"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func ssb(c *xsf.Caller, tm time.Duration) (*xsf.Res, string, int32, error) {
	//fmt.Println("enter ssb")
	//defer fmt.Println("leave ssb")

	req := xsf.NewReq()
	req.SetParam("k1", "v1")
	req.SetParam("k2", "v2")
	req.SetParam("k3", "v3")

	res, code, e := c.SessionCall(xsf.CREATE, "xsf-server", "ssb", req, tm)
	if code != 0 || e != nil {
		//log.Fatal("ssb err", e)
		return res, "", code, e
	}
	var sess = res.Session()

	return res, sess, code, e
}

func auw(c *xsf.Caller, sess string, tm time.Duration) (*xsf.Res, int32, error) {
	//fmt.Println("enter auw")
	//defer fmt.Println("leave auw")
	req := xsf.NewReq()
	req.SetParam("k1", "v1")
	req.SetParam("k2", "v2")
	req.SetParam("k3", "v3")

	_ = req.Session(sess)

	res, code, e := c.SessionCall(xsf.CONTINUE, "xsf-server", "auw", req, tm)
	if code != 0 || e != nil {
		//log.Fatal("auw err")
		return res, code, e
	}

	return res, code, e
}

func sse(c *xsf.Caller, sess string, tm time.Duration) (*xsf.Res, int32, error) {
	//fmt.Println("enter sse")
	//defer fmt.Println("leave sse")
	req := xsf.NewReq()
	req.SetParam("k1", "v1")
	req.SetParam("k2", "v2")
	req.SetParam("k3", "v3")
	req.Session(sess)
	res, code, e := c.SessionCall(xsf.CONTINUE, "xsf-server", "sse", req, tm)
	if code != 0 || e != nil {
		//log.Fatal("sse err", e)
		return res, code, e
	}
	return res, code, e
}

func sessionCallExample(c *xsf.Caller, tm time.Duration) {
	ssbBase := time.Now()
	_, sess, code, err := ssb(c, tm)
	if code != 0 || err != nil {
		log.Printf("ssbDur:%vms,code:%v, err:%v\n", time.Since(ssbBase).Microseconds(), code, err.Error())
		return
	}
	auwBase := time.Now()
	_, code, err = auw(c, sess, tm)
	if code != 0 || err != nil {
		log.Printf("auwBase:%vms,code:%v, err:%v\n", time.Since(auwBase).Microseconds(), code, err.Error())
		return
	}
	sseBase := time.Now()
	_, code, err = sse(c, sess, tm)
	if code != 0 || err != nil {
		log.Printf("sseBase:%vms,code:%v, err:%v\n", time.Since(sseBase).Microseconds(), code, err.Error())
		return
	}

}
func pingTest(c *xsf.Caller, tm time.Duration) {
	span := utils.NewSpan(utils.SrvSpan).Start()
	defer span.End().Flush()

	c.WithApiVersion(apiVersion)
	c.WithRetry(1)

	req := xsf.NewReq()

	req.SetTraceID(span.Meta()) //将span信息带到后端

	_, code, e := c.Call("xsf-server", "req", req, tm)
	if code != 0 || e != nil {
		log.Fatal("sse err", code, e)
	}
	//_, code, e = c.Call("xsf-server2", "req", req, tm)
	//if code != 0 || e != nil {
	//	log.Fatal("sse err", code, e)
	//}
	//_, code, e = c.Call("xsf-server3", "req", req, tm)
	//if code != 0 || e != nil {
	//	log.Fatal("sse err", code, e)
	//}
}
func sessionCallWithOneShortFlag(c *xsf.Caller, tm time.Duration) {
	c.WithApiVersion(apiVersion)
	c.WithLBParams("xsf-lbv2", "iat", nil)
	fmt.Println("enter oneShort")
	defer fmt.Println("leave oneShort")

	req := xsf.NewReq()
	req.SetParam("k1", "v1")
	req.SetParam("k2", "v2")
	req.SetParam("k3", "v3")

	res, code, e := c.SessionCall(xsf.ONESHORT, "sms", "req", req, tm)
	fmt.Println(res, code, e)

}

func callTest(c *xsf.Caller, tm time.Duration) {
	span := utils.NewSpan(utils.SrvSpan).Start()
	defer span.End().Flush()
	req := func() *xsf.Req {
		dataIn := protocol.EngInputData{}
		dataIn.EngParam = make(map[string]string)
		dataIn.EngParam["sid"] = "xxx"
		//dataIn.EngParam["PatchId"] = "1345922469150941184"
		//dataIn.EngParam["PatchIdPath"] = "/s3/buckets-0/IFLYREC_OST_none_iat_spp_1346073617245757440_model_nil_nil_nil_nil_nil"
		dataIn.EngParam["patch_id"] = "1245922469154199284"
		dataIn.EngParam["patch_id_path"] = "/s3/buckets-0/IFLYREC_OST__iat_1346023995517964288_model_nil_nil_nil_nil_nil"

		dataMeta := protocol.MetaData{
			DataId:   "",
			FrameId:  0,
			DataType: protocol.MetaData_TEXT,
			Status:   protocol.MetaData_ONCE,
			Format:   "",
			Encoding: "",
			Data:     nil,
			Desc:     nil,
		}
		dataIn.DataList = append(dataIn.DataList, &dataMeta)

		in, _ := proto.Marshal(&dataIn)
		rd := xsf.NewData()
		rd.Append(in)
		req := xsf.NewReq()
		req.AppendData(rd)
		req.SetParam("SeqNo", "1")
		req.SetParam("waitTime", "30000") //ms
		return req
	}()

	{ /*
			res, code, e := c.CallWithAddr("", "AILoad", "172.31.98.182:5090", req, time.Minute)

			if code != 0 || e != nil {
				log.Fatalf("h:%v,err=>code:%v,err:%v", req.Handle(), code, e)
			} else {
				fmt.Printf("handle:%v,res:%v\n", req.Handle(), fmt.Sprintf("params->%v,data->%v", res.GetAllParam(), res.GetData()))
			}
		*/}
	{
		res, code, e := c.CallWithAddr("", "AIUnLoad", "172.31.98.182:5090", req, time.Minute)

		if code != 0 || e != nil {
			log.Fatalf("h:%v,err=>code:%v,err:%v", req.Handle(), code, e)
		} else {
			fmt.Printf("handle:%v,res:%v\n", req.Handle(), fmt.Sprintf("params->%v,data->%v", res.GetAllParam(), res.GetData()))
		}
	}
}

func callExample(c *xsf.Caller, tm time.Duration) {
	//span := utils.NewSpan(utils.SrvSpan).Start()
	//defer span.End().Flush()

	//span = span.WithName("callExample")
	//span = span.WithTag("customKey1", "customVal1")
	//span = span.WithTag("customKey2", "customVal2")
	//span = span.WithTag("customKey3", "customVal3")
	//c.WithRetry(3)
	//c.WithApiVersion(apiVersion)
	//c.WithLBParams("xsf-lbv2", "iat", nil)
	req := xsf.NewReq()
	req.SetParam("k1", "v1")
	req.SetParam("k2", "v2")
	req.SetParam("k3", "v3")
	//req.SetParam("failed", "2")
	//req.SetTraceID(span.Meta()) //将span信息带到后端

	_, code, e := c.Call("xsf-server", "req", req, tm)
	//r, code, e := c.Call("sms", "req", req, tm)
	if code != 0 || e != nil {
		log.Println("call err", code, e)
	}
	////ip, _ := r.GetParam("ip")
	////port, _ := r.GetParam("port")
	////fmt.Println(ip, port)
	////all := r.GetAllParam()
	////fmt.Printf("allParams:%+v\n", all)
	//vcpu, _ := r.GetParam("vcpu")
	//fmt.Printf("NO.1 vcpu:%v,sessMap:%+v\n", vcpu, r.GetSess())
	//
	//r.Session()
	//_ = req.Session(r.Session())
	//r, code, e = c.Call("xsf-server", "req", req, tm)
	//if code != 0 || e != nil {
	//	log.Fatal("sse err", code, e)
	//}
	//vcpu, _ = r.GetParam("vcpu")
	//fmt.Printf("NO.2 vcpu:%v,sessMap:%+v\n", vcpu, r.GetSess())
	//
	//r.Session()
	//_ = req.Session(r.Session())
	//r, code, e = c.Call("xsf-server", "req", req, tm)
	//if code != 0 || e != nil {
	//	log.Fatal("sse err", code, e)
	//}
	//vcpu, _ = r.GetParam("vcpu")
	//fmt.Printf("NO.3 vcpu:%v,sessMap:%+v\n", vcpu, r.GetSess())
}
func callTest2(c *xsf.Caller, tm time.Duration) {
	span := utils.NewSpan(utils.CliSpan).Start()
	defer span.End().Flush()

	span = span.WithName("callExample")
	span = span.WithTag("customKey1", "customVal1")
	span = span.WithTag("customKey2", "customVal2")
	span = span.WithTag("customKey3", "customVal3")

	c.WithApiVersion(apiVersion)
	c.WithRetry(3)
	c.WithLBParams("xsf-lbv2", "iat", nil)

	{ //第一组测试
		req := xsf.NewReq()
		req.SetParam("k1", "v1")
		req.SetParam("k2", "v2")
		req.SetParam("k3", "v3")

		req.SetTraceID(span.Meta()) //将span信息带到后端
		baseTime := time.Now()
		res, code, e := c.Call("xsf-server", "req", req, tm)
		dur := time.Now().Sub(baseTime)
		if code != 0 || e != nil {
			log.Fatalf("sse err,code:%v,e:%v,dur:%v\n", code, e, dur.Seconds())
		} else {
			fmt.Printf("F.NO.1 => handle:%s,dur:%vs\n", res.Handle(), dur.Seconds())
		}

		res, code, e = c.Call("xsf-server", "req", req, tm)
		if code != 0 || e != nil {
			log.Fatal("sse err", code, e)
		} else {
			fmt.Printf("F.NO.2 => handle:%s\n", res.Handle())
		}

		res, code, e = c.Call("xsf-server", "req", req, tm)
		if code != 0 || e != nil {
			log.Fatal("sse err", code, e)
		} else {
			fmt.Printf("F.NO.3 => handle:%s\n", res.Handle())
		}
	}

	//{ //第二组测试
	//	c.WithHashKey("555")
	//	req := xsf.NewReq()
	//	req.SetParam("k1", "v1")
	//	req.SetParam("k2", "v2")
	//	req.SetParam("k3", "v3")
	//
	//	req.SetTraceID(span.Meta()) //将span信息带到后端
	//
	//	res, code, e := c.Call("xsf-server", "req", req, tm)
	//	if code != 0 || e != nil {
	//		log.Fatal("sse err", code, e)
	//	} else {
	//		fmt.Printf("S.NO.1 => handle:%s\n", res.Handle())
	//	}
	//
	//	res, code, e = c.Call("xsf-server", "req", req, tm)
	//	if code != 0 || e != nil {
	//		log.Fatal("sse err", code, e)
	//	} else {
	//		fmt.Printf("S.NO.2 => handle:%s\n", res.Handle())
	//	}
	//
	//	res, code, e = c.Call("xsf-server", "req", req, tm)
	//	if code != 0 || e != nil {
	//		log.Fatal("sse err", code, e)
	//	} else {
	//		fmt.Printf("S.NO.3 => handle:%s\n", res.Handle())
	//	}
	//}
}
func callWithAddr(c *xsf.Caller, tm time.Duration) {
	baseTime := time.Now()

	req := xsf.NewReq()
	data := xsf.NewData()
	data.Append(bytes.Repeat([]byte("b"), 100*1024*1024))
	req.AppendData(data)

	r, code, e := c.CallWithAddr("", "req", "127.0.0.1:1234", req, tm)
	fmt.Printf("dur:%v,r:%v,code:%v,e:%v\n", time.Now().Sub(baseTime).String(), r, code, e)
}
func callConHashTest(c *xsf.Caller, tm time.Duration) {
	span := utils.NewSpan(utils.SrvSpan).Start()
	defer span.End().Flush()

	span = span.WithName("callExample")
	span = span.WithTag("customKey1", "customVal1")
	span = span.WithTag("customKey2", "customVal2")
	span = span.WithTag("customKey3", "customVal3")

	c.WithApiVersion(apiVersion)
	c.WithRetry(3)
	var count int64

	test := func(hashKey, svc string) {
		baseTime := time.Now()
		addr, addrErr := c.GetHashAddr(hashKey, svc)
		fmt.Printf("NO.%v dur:%v,addr:%v,addrErr:%v,hashKey:%v,svc:%v\n",
			time.Now().Sub(baseTime).String(), atomic.AddInt64(&count, 1), addr, addrErr, hashKey, svc)
	}
	{
		fmt.Println("------------------------------")
		hashKey, svc := "111", "xsf-server"
		test(hashKey, svc)
		test(hashKey, svc)
		test(hashKey, svc)
		test(hashKey, svc)
	}
	{
		fmt.Println("------------------------------")
		hashKey, svc := "432", "xsf-server"
		test(hashKey, svc)
		test(hashKey, svc)
		test(hashKey, svc)
		test(hashKey, svc)
	}
	{
		req := xsf.NewReq()
		req.SetParam("k1", "v1")
		req.SetParam("k2", "v2")
		req.SetParam("k3", "v3")

		req.SetTraceID(span.Meta()) //将span信息带到后端
		baseTime := time.Now()
		c.WithHashKey("classifyID")
		res, code, e := c.Call("xsf-server", "req", req, tm)
		dur := time.Now().Sub(baseTime)
		if code != 0 || e != nil {
			log.Fatalf("sse err,code:%v,e:%v,dur:%v\n", code, e, dur.Seconds())
		} else {
			fmt.Printf("F.NO.1 => handle:%s,dur:%vs\n", res.Handle(), dur.Seconds())
		}
	}
}
func crashTest(c *xsf.Caller) {
	go func() {
		for {
			time.Sleep(time.Second)
			fmt.Println(runtime.NumGoroutine())
		}
	}()
	var wg sync.WaitGroup
	for ix := 0; ix < 500000; ix++ {
		wg.Add(1)
		go func() {
			for {
				_, _, e := c.CallWithAddr("iat", "req", "10.1.87.69:8080", xsf.NewReq(), time.Millisecond*200)
				if e != nil {
					fmt.Println(e)
				}
			}
		}()
	}
	wg.Wait()
}
