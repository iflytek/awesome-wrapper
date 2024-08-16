package main

import "C"
import (
	"comwrapper"
	"fmt"
	"strconv"
	"time"
	"unsafe"

	"git.iflytek.com/AIaaS/xsf/utils"
)

var (
	wLogger  *utils.Logger
	logLevel = "debug"
	logCount = 10
	logSize  = 30
	logAsync = true
	logPath  = "/log/app/wrapper/wrapper.log"
)

var traceLogFunc func(usrTag string, key string, value string) (code int)

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

// WrapperCreate 插件会话示例创建, 每次建立会话请求时创建. 本地调试时, params参数由xtest.toml提供
func WrapperCreate(usrTag string, params map[string]string, prsIds []int, cb comwrapper.CallBackPtr) (hdl unsafe.Pointer, err error) {
	sid := params["sid"]
	paramStr := ""
	for k, v := range params {
		paramStr += fmt.Sprintf("%s=%s;", k, v)
	}
	wLogger.Debugw("WrapperCreate params", "paramStr", paramStr, "sid", sid)

	inst := wrapperInst{
		sid:    sid,
		usrTag: usrTag,
	}

	wLogger.Debugw("WrapperCreate successful", "sid", sid)
	return unsafe.Pointer(&inst), nil
}

// WrapperWrite 数据写入
func WrapperWrite(hdl unsafe.Pointer, req []comwrapper.WrapperData) (err error) {
	inst := (*wrapperInst)(hdl)

	if len(req) == 0 {
		wLogger.Debugw("WrapperWrite data is nil", "sid", inst.sid)
		return nil
	}

	for _, v := range req {
		if err = inst.write(v.Status, v.Data); err != nil {
			wLogger.Errorw("WrapperWrite inst.write", "error", err.Error(), "sid", inst.sid)
			return err
		}
	}

	return nil
}

// WrapperRead 数据结果读取
func WrapperRead(hdl unsafe.Pointer) (respData []comwrapper.WrapperData, err error) {
	inst := (*wrapperInst)(hdl)

	status, data, err := inst.read()
	if err != nil {
		wLogger.Errorw("WrapperRead inst.read", "error", err.Error(), "sid", inst.sid)
		return nil, err
	}

	resultData := comwrapper.WrapperData{
		Type:   comwrapper.DataText,
		Key:    "result",
		Data:   data,
		Desc:   map[string]string{},
		Status: status,
	}

	respData = append(respData, resultData)

	return
}

func WrapperDestroy(hdl interface{}) (err error) {
	return
}

func WrapperFini() (err error) {
	return
}

func WrapperExec(usrTag string, params map[string]string, reqData []comwrapper.WrapperData) (respData []comwrapper.WrapperData, err error) {
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
		cf := f.(func(usrTag string, key string, value string) (code int))
		traceLogFunc = cf
		fmt.Println("WrapperSetCtrl traceLogFunc set successful.")
	default:

	}
	return
}

type wrapperInst struct {
	usrTag string
	sid    string
	db     *storage
}

type storage struct {
	status comwrapper.DataStatus
	data   []byte
}

func (inst *wrapperInst) write(status comwrapper.DataStatus, data []byte) error {
	if inst.db == nil {
		inst.db = &storage{}
	}
	inst.db.status = status
	inst.db.data = data
	return nil
}

func (inst *wrapperInst) read() (status comwrapper.DataStatus, data []byte, err error) {
	return inst.db.status, inst.db.data, nil
}

func (inst *wrapperInst) traceLogWithTime(key string, msg string) {
	tn := time.Now()
	formattedTime := tn.Format("2006-01-02 15:04:05.999999")
	traceLogFunc(inst.usrTag, key, "time:"+formattedTime+"; "+msg)
}
