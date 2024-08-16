package main

import (
	"comwrapper"
	"errors"
	"fmt"
	"strconv"
	"strings"
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

// WrapperExec 非流式调用. params参数由客户端请求时传入, 本地调试时, params参数由xtest.toml提供
func WrapperExec(usrTag string, params map[string]string, reqData []comwrapper.WrapperData) (respData []comwrapper.WrapperData, err error) {
	sid := params["sid"]
	wLogger.Debugw("WrapperExec Enter", "params", params, "sid", sid)

	respData = make([]comwrapper.WrapperData, 0, len(reqData))
	for _, req := range reqData {
		text := strings.TrimSpace(string(req.Data))
		numbers := strings.Split(text, ",")
		var total int64
		for _, number := range numbers {
			n, err := strconv.ParseInt(number, 10, 64)
			if err != nil {
				wLogger.Errorw("WrapperExec params validation", "error", err.Error(), "sid", sid)
				return nil, errors.New("support integer input only")
			}
			total += n
		}

		// 上报trace日志, 建议上传关键、小体积的日志信息 或 不上报
		if code := traceLogFunc(usrTag, "total", strconv.Itoa(int(total))); code != 0 {
			wLogger.Errorw("WrapperExec trace total", "code", code, "sid", sid)
		}

		respData = append(respData, comwrapper.WrapperData{
			Key:      "result",
			Data:     []byte(strconv.Itoa(int(total))),
			Desc:     nil,
			Encoding: "",
			Type:     comwrapper.DataText,
			Status:   comwrapper.DataOnce,
		})
	}

	return
}

// WrapperDestroy 插件资源销毁
func WrapperDestroy(hdl interface{}) (err error) {
	return
}

// WrapperCreate 插件会话实例创建, 每次建立会话请求时调用. 本地调试时, params参数由xtest.toml提供
func WrapperCreate(usrTag string, params map[string]string, prsIds []int, cb comwrapper.CallBackPtr) (hdl unsafe.Pointer, err error) {
	return
}

// WrapperWrite 数据写入
func WrapperWrite(hdl unsafe.Pointer, req []comwrapper.WrapperData) (err error) {
	return
}

// WrapperRead 数据结果读取
func WrapperRead(hdl unsafe.Pointer) (respData []comwrapper.WrapperData, err error) {
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
		cf := f.(func(usrTag string, key string, value string) (code int))
		traceLogFunc = cf
		fmt.Println("WrapperSetCtrl traceLogFunc set successful.")
	default:

	}
	return
}
