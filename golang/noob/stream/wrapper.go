package main

import (
	"comwrapper"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mozillazg/go-pinyin"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"git.iflytek.com/AIaaS/xsf/utils"
)

var (
	wLogger       *utils.Logger
	logLevel      = "debug"
	logCount      = 10
	logSize       = 30
	logAsync      = true
	logPath       = "/log/app/wrapper.log"
	respKey       = "result"
	meterFunction = "chars.total"
)

var traceFunc func(usrTag string, key string, value string) (code int)
var meterFunc func(usrTag string, key string, count int) (code int)

// WrapperInit 插件初始化, 全局只调用一次. 本地调试时, cfg参数由aiges.toml提供
func WrapperInit(cfg map[string]string) (err error) {
	fmt.Println("---- wrapper init ----")
	for k, v := range cfg {
		fmt.Printf("config param %s=%s\n", k, v)
	}
	if v, ok := cfg["log_path"]; ok {
		logPath = v
	}
	if v, ok := cfg["log_level"]; ok {
		logLevel = v
	}
	if v, ok := cfg["log_count"]; ok {
		logCount, _ = strconv.Atoi(v)
	}
	if v, ok := cfg["log_size"]; ok {
		logSize, _ = strconv.Atoi(v)
	}
	if cfg["log_async"] == "false" {
		logAsync = false
	}
	if v, ok := cfg["resp_key"]; ok {
		respKey = v
	}

	wLogger, err = utils.NewLocalLog(
		utils.SetAsync(logAsync),
		utils.SetLevel(logLevel),
		utils.SetFileName(logPath),
		utils.SetMaxSize(logSize),
		utils.SetMaxBackups(logCount),
	)
	if err != nil {
		return fmt.Errorf("loggerErr:%v", err)
	}

	wLogger.Debugw("WrapperInit successful")
	return
}

// WrapperCreate 插件会话实例创建, 每次建立会话请求时调用. 本地调试时, params参数由xtest.toml提供
func WrapperCreate(usrTag string, params map[string]string, prsIds []int, cb comwrapper.CallBackPtr) (hdl unsafe.Pointer, err error) {
	sid := params["sid"]
	paramStr := ""
	for k, v := range params {
		paramStr += fmt.Sprintf("%s=%s;", k, v)
	}
	wLogger.Debugw("WrapperCreate params", "paramStr", paramStr, "sid", sid)

	inst := newEngine(usrTag, sid)

	if cb != nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					wLogger.Errorw("engine pusher crashed", "err", r, "sid", inst.sid)
				}
			}()

			if err = inst.push(usrTag, cb); err != nil {
				wLogger.Errorw("engine push error", "error", err, "sid", sid, "usrTag", usrTag)
			}
		}()
	}

	wLogger.Debugw("WrapperCreate successful", "sid", sid)
	return unsafe.Pointer(&inst), nil
}

// WrapperWrite 数据写入
func WrapperWrite(hdl unsafe.Pointer, req []comwrapper.WrapperData) (err error) {
	inst := (*engine)(hdl)

	if len(req) == 0 {
		wLogger.Debugw("WrapperWrite data is nil", "sid", inst.sid)
		return nil
	}

	var status comwrapper.DataStatus
	for _, v := range req {
		status = v.Status
		inst.meterCount += len(v.Data)

		if err = inst.write(string(v.Data), v.Status); err != nil {
			wLogger.Errorw("WrapperWrite inst.write", "error", err.Error(), "sid", inst.sid)
			break
		}
	}

	// 最后一帧时, 上报计量数据
	if status == comwrapper.DataEnd {
		code := meterFunc(inst.usrTag, meterFunction, inst.meterCount)
		if code == 0 {
			wLogger.Debugw("trace metering data", "code", code, "count", inst.meterCount, "sid", inst.sid)
		} else {
			wLogger.Errorw("failed to report metering data", "code", code, "count", inst.meterCount, "sid", inst.sid)
		}
	}

	return err
}

// WrapperRead 数据结果读取, 服务配置[aiges]项配置asyncMode = false
func WrapperRead(hdl unsafe.Pointer) (respData []comwrapper.WrapperData, err error) {
	inst := (*engine)(hdl)

	data, status, err := inst.read()
	if err != nil {
		wLogger.Errorw("WrapperRead inst.read", "error", err.Error(), "sid", inst.sid)
		return nil, err
	}

	resultData := comwrapper.WrapperData{
		Key:      respKey,
		Data:     data,
		Desc:     map[string]string{},
		Encoding: "utf-8",
		Type:     comwrapper.DataText,
		Status:   status,
	}

	respData = append(respData, resultData)

	return
}

// WrapperDestroy 会话资源销毁
func WrapperDestroy(hdl interface{}) (err error) {
	inst := (*engine)(hdl.(unsafe.Pointer))
	wLogger.Debugw("WrapperDestroy", "sid", inst.sid)
	return
}

// WrapperExec 非流式请求-同步响应
func WrapperExec(usrTag string, params map[string]string, reqData []comwrapper.WrapperData) (respData []comwrapper.WrapperData, err error) {
	wLogger.Debugw("WrapperExec", "params", params, "reqData", reqData)
	e := newEngine(usrTag, params["sid"])
	payload := reqData[0]
	if err = e.write(string(payload.Data), payload.Status); err != nil {
		wLogger.Errorw("WrapperExec engine write", "error", err, "sid", e.sid)
		return nil, err
	}

	data, _, err := e.read()
	if err != nil {
		wLogger.Errorw("WrapperExec engine read", "error", err, "sid", e.sid)
		return nil, err
	}

	respData = append(respData, comwrapper.WrapperData{
		Key:      respKey,
		Data:     data,
		Desc:     map[string]string{},
		Encoding: "utf-8",
		Type:     comwrapper.DataText,
		Status:   comwrapper.DataOnce,
	})

	return
}

func WrapperFini() (err error) {
	return
}

func WrapperVersion() (version string) {
	return
}
func WrapperLoadRes(res comwrapper.WrapperData, resId int) (err error) {
	return
}
func WrapperUnloadRes(resId int) (err error) {
	return
}
func WrapperDebugInfo(hdl interface{}) (debug string) {
	return
}

func WrapperSetCtrl(fType comwrapper.CustomFuncType, f interface{}) (err error) {
	switch fType {
	case comwrapper.FuncTraceLog:
		traceFunc = f.(func(usrTag string, key string, value string) (code int))
		fmt.Println("WrapperSetCtrl traceLogFunc set successful.")
	case comwrapper.FuncMeter:
		meterFunc = f.(func(usrTag string, key string, count int) (code int))
		fmt.Println("WrapperSetCtrl meterFunc set successful.")
	default:

	}
	return
}

func (e *engine) traceLogWithTime(key string, msg string) {
	tn := time.Now()
	formattedTime := tn.Format("2006-01-02 15:04:05.999999")
	traceFunc(e.usrTag, key, "time:"+formattedTime+"; "+msg)
}

// 模拟引擎 - 中文转拼音引擎
type engine struct {
	meterCount int
	usrTag     string
	sid        string
	base       pinyin.Args
	jobs       *jobs
	jobsLock   sync.Mutex
}

type jobs struct {
	head *jobNode
	tail *jobNode
}

type jobNode struct {
	hans   string
	py     [][]string
	status comwrapper.DataStatus
	next   *jobNode
}

func newEngine(usrTag, sid string) *engine {
	a := pinyin.NewArgs()
	a.Style = pinyin.Tone

	return &engine{
		usrTag: usrTag,
		sid:    sid,
		base:   a,
		jobs:   &jobs{},
	}
}

// 模拟向引擎模型写入数据
func (e *engine) write(hans string, status comwrapper.DataStatus) error {
	e.jobsLock.Lock()
	defer e.jobsLock.Unlock()

	if e.jobs.head == nil {
		e.jobs.head = &jobNode{hans: hans, py: pinyin.Pinyin(hans, e.base), status: status}
		e.jobs.tail = e.jobs.head
		return nil
	}

	e.jobs.tail.next = &jobNode{hans: hans, py: pinyin.Pinyin(hans, e.base), status: status}
	e.jobs.tail = e.jobs.tail.next

	return nil
}

// 模拟从引擎模型读取数据, 服务配置[aiges]项配置asyncMode = false
func (e *engine) read() ([]byte, comwrapper.DataStatus, error) {
	e.jobsLock.Lock()
	defer e.jobsLock.Unlock()

	if e.jobs.head == nil {
		return nil, 0, errors.New("no valid jobs")
	}

	data, err := json.Marshal(&Result{Hans: e.jobs.head.hans, Pinyin: e.jobs.head.py})
	if err != nil {
		return nil, 0, err
	}

	status := e.jobs.head.status

	e.jobs.head = e.jobs.head.next
	return data, status, nil
}

// 引擎主动向加载器推送数据, 服务配置[aiges]项配置asyncMode = true
func (e *engine) push(handle string, cb comwrapper.CallBackPtr) error {
	for {
		var head *jobNode
		e.jobsLock.Lock()
		head = e.jobs.head
		e.jobsLock.Unlock()

		if head == nil {
			time.Sleep(time.Millisecond * 200)
			continue
		}

		data, err := json.Marshal(&Result{Hans: head.hans, Pinyin: head.py})
		if err != nil {
			return cb(handle, nil, err)
		}

		resp := []comwrapper.WrapperData{
			{
				Key:      respKey,
				Data:     data,
				Desc:     nil,
				Encoding: "utf-8",
				Type:     comwrapper.DataText,
				Status:   head.status,
			},
		}

		if err = cb(handle, resp, nil); err != nil {
			return err
		}

		if head.status == comwrapper.DataEnd || head.status == comwrapper.DataOnce {
			break
		}

		e.jobsLock.Lock()
		e.jobs.head = e.jobs.head.next
		e.jobsLock.Unlock()
		time.Sleep(time.Millisecond * 200)
	}

	return nil
}

type Result struct {
	Hans   string     `json:"hans"`
	Pinyin [][]string `json:"pinyin"` // 可能包含多音字
}
