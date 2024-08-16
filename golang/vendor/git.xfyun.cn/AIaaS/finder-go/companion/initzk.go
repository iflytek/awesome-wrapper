package finder

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	errors "git.xfyun.cn/AIaaS/finder-go/errors"
	"git.xfyun.cn/AIaaS/finder-go/utils/httputil"
)

// GetZkInfo for getting zk metadata
func GetZkInfo(hc *http.Client, url string) (*common.ZkInfo, error) {
	var result []byte
	var err error
	retryNum := 0
	for {
		result, err = httputil.DoGet(hc, url)
		if err != nil {
			log.Println(err)
			if retryNum < 3 {
				retryNum++
				log.Println("The ", retryNum, "th GetZkInfo")
				time.Sleep(time.Millisecond * 100)
				continue
			} else {
				return nil, err
			}
		} else {
			break
		}
	}

	var r JSONResult
	err = json.Unmarshal([]byte(result), &r)
	if err != nil {
		return nil, err
	}
	if r.Ret != 0 {
		err = &errors.FinderError{
			Ret:  errors.ZkGetInfoError,
			Func: "GetZkInfo",
			Desc: r.Msg,
		}
		return nil, err
	}

	ok := true
	if _, ok = r.Data["config_path"]; !ok {
		err = &errors.FinderError{
			Ret:  errors.ZkMissRootPath,
			Func: "GetZkInfo",
			Desc: "miss config path",
		}

		return nil, err
	}

	if _, ok = r.Data["service_path"]; !ok {
		err = &errors.FinderError{
			Ret:  errors.ZkMissRootPath,
			Func: "GetZkInfo",
			Desc: "miss service path",
		}

		return nil, err
	}

	var zkAddr []string
	if _, ok = r.Data["zk_addr"]; !ok {
		err = &errors.FinderError{
			Ret:  errors.ZkMissAddr,
			Func: "GetZkInfo",
			Desc: "miss zk_info",
		}

		return nil, err
	}

	var value []interface{}
	if value, ok = r.Data["zk_addr"].([]interface{}); ok {
		zkAddr = convert(value)
		if len(zkAddr) == 0 {
			err = &errors.FinderError{
				Ret:  errors.ZkMissAddr,
				Func: "GetZkInfo",
				Desc: "convert failure",
			}

			return nil, err
		}
	}

	zkInfo := &common.ZkInfo{
		ConfigRootPath:  r.Data["config_path"].(string),
		ServiceRootPath: r.Data["service_path"].(string),
		ZkAddr:          zkAddr,
	}

	return zkInfo, nil
}

func convert(in []interface{}) []string {
	r := make([]string, 0)
	ok := true
	value := ""
	for _, v := range in {
		if value, ok = v.(string); ok {
			r = append(r, value)
		}
	}

	return r
}
