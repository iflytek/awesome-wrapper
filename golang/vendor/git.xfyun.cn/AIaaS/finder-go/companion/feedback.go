package finder

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	errors "git.xfyun.cn/AIaaS/finder-go/errors"
	"git.xfyun.cn/AIaaS/finder-go/utils/httputil"
)

func FeedbackForConfig(hc *http.Client, url string, f *common.ConfigFeedback) error {
	log.Println("called FeedbackForConfig")
	contentType := "application/x-www-form-urlencoded"
	params := []byte(fmt.Sprintf("push_id=%s&project=%s&group=%s&service=%s&version=%s&addr=%s&config=%s&update_time=%d&update_status=%d&load_time=%d&load_status=%d",
		f.PushID, f.ServiceMete.Project, f.ServiceMete.Group, f.ServiceMete.Service, f.ServiceMete.Version, f.ServiceMete.Address,
		f.Config, f.UpdateTime, f.UpdateStatus, f.LoadTime, f.LoadStatus))
	result, err := httputil.DoPost(hc, contentType, url, params)
	if err != nil {
		log.Println("FeedbackForConfig err:", err)
		return err
	} else {
		log.Println("FeedbackForConfig result:", string(result))
	}

	var r JSONResult
	err = json.Unmarshal([]byte(result), &r)
	if err != nil {
		log.Println("FeedbackForConfig err:", err)
		return err
	}
	if r.Ret != 0 {
		err = &errors.FinderError{
			Ret:  errors.FeedbackConfigError,
			Func: "FeedbackForConfig",
			Desc: r.Msg,
		}

		log.Println("FeedbackForConfig err:", err)
		return err
	}

	return nil
}

func FeedbackForService(hc *http.Client, url string, f *common.ServiceFeedback) error {
	contentType := "application/x-www-form-urlencoded"
	params := []byte(fmt.Sprintf("push_id=%s&project=%s&group=%s&consumer=%s&consumer_version=%s&addr=%s&provider=%s&provider_version=%s&update_time=%d&update_status=%d&load_time=%d&load_status=%d",
		f.PushID, f.ServiceMete.Project, f.ServiceMete.Group, f.ServiceMete.Service, f.ServiceMete.Version, f.ServiceMete.Address,
		f.Provider, f.ProviderVersion, f.UpdateTime, f.UpdateStatus, f.LoadTime, f.LoadStatus))
	result, err := httputil.DoPost(hc, contentType, url, params)
	if err != nil {
		log.Println(err)
		return err
	}

	var r JSONResult
	err = json.Unmarshal([]byte(result), &r)
	if err != nil {
		return err
	}
	if r.Ret != 0 {
		err = &errors.FinderError{
			Ret:  errors.FeedbackServiceError,
			Func: "FeedbackForService",
			Desc: r.Msg,
		}

		return err
	}

	return nil
}
