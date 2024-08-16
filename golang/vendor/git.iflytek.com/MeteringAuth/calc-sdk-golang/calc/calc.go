/*
 * @CopyRight: Copyright 2019 IFLYTEK Inc
 * @Author: jianjiang@iflytek.com
 * @LastEditors: Please set LastEditors
 * @Desc:
 * @createTime: 2020-07-23 15:25:36
 * @modifyTime: 2020-07-24 19:18:07
 */

package calc

import (
	"fmt"
	"git.iflytek.com/MeteringAuth/calc-sdk-golang/calc/internal/logger"
	"strings"
	"time"
)

var VERSION = "2.3.1"

// 2.0版本协议默认不记录
var ProtocolVersion = "2.0"

const PUBLICCLOUD = "0"

var (
	calcCore *CalcCore   = &CalcCore{}
	config   *ConfigCalc = &ConfigCalc{}
	inited   bool        = false
	isStop   bool        = false
)

var (
	qSize   int = 10000
	pNumber int = 10
	topic   string
	timeout time.Duration
	enable  bool

	configName string = "calc.toml"

	//companionUrl string = "http://10.1.87.69:6868"
	//project      string = "guiderAllService"
	//group        string = "gas"
	service string = "calc-client"

	level        = "error"
	logFormat    = "normal"
	logPath      = "./calc.log"
	consolePrint = false
)

// 配置文件存储结构
var (
	tomlconfig CalcTomlConfig
)

// 输入包含具体的用户参数信息，appid , channel , funcs不可为空，c 不可小于等于0
func Calc(appid, channel, funcs string, c int64) (errorcode int, err error) {

	if !enable {
		return 0, nil
	}
	if !inited {
		logger.Warnw("sdk have not been inited")
		errorcode, err = CalcHasNotInited.GetCode(), CalcHasNotInited
	}
	if isStop {
		logger.Warnw("sdk have not been started")
		errorcode, err = CalcStoped.GetCode(), CalcStoped
	}
	if calcErr := paramsCheck(appid, channel, funcs, c); calcErr != nil {
		logger.Warnw("invalid parameters", "desc", calcErr.Error())
		errorcode, err = calcErr.GetCode(), calcErr
	}
	if calcErr := calcCore.Set(appid, "", channel, funcs, int(c)); calcErr != nil {
		logger.Warnw("calcCore set error", "desc", calcErr.Error())
		errorcode, err = calcErr.GetCode(), calcErr
	}
	return

}

// SUB生成规则<SUBID_APPID_subId>
func CalcWithSubId(appid, subId, channel, funcs string, c int64) (errorcode int, err error) {

	uniqueDid := strings.Join([]string{"SUBID", appid, subId}, "_")
	return Calc(uniqueDid, channel, funcs, c)

}

func CalcMultiTenancy(appid, cloudid, composeid, serviceid, funcs string, c int64) (errorcode int, err error) {
	var calcErr *CalcError
	channel, calcErr := generateChannel(cloudid, composeid, serviceid)
	if err != nil {
		return calcErr.GetCode(), calcErr
	}

	return Calc(appid, channel, funcs, c)
}

func CalWithSubIdMultiTenancy(appid, subId, cloudid, composeid, serviceid, funcs string, c int64) (errorcode int, err error) {
	var calcErr *CalcError
	channel, calcErr := generateChannel(cloudid, composeid, serviceid)
	if err != nil {
		return calcErr.GetCode(), calcErr
	}

	return CalcWithSubId(appid, subId, channel, funcs, c)
}

//func CalcWithUid(uid, channel, funcs string, c int64) (errorcode int, err error) {
//	return Calc(uid, channel, funcs, c)
//}

func Init(url, pro, gro, service, version string, isNative bool, nativeLogPath string) error {

	inited = true
	config.isNative = isNative
	config.nativeLogPath = nativeLogPath
	if err := config.init(url, pro, gro, service, version); err != nil {
		return err
	}

	err := logger.L(logger.ConfigOption{
		Level:   level,
		LogPath: logPath,
		LogFormat: func() logger.LogFormat {
			if logFormat == "json" {
				return logger.JSON
			} else {
				return logger.CONSOLE
			}
		}(),
		ConsolePrint: consolePrint,
	})
	if err != nil {
		fmt.Printf("logger init failed , err = %s\n", err)
		return err
	}
	logger.Infow("toml content after loading", "content", tomlconfig)
	if !enable {
		logger.Warnw("calc sdk has been disabled")
		return nil
	}

	if err := calcCore.Init(qSize, pNumber, topic, tomlconfig.Calc.Hosts, timeout); err != nil {
		logger.Errorw("calc sdk init failed", "error", err)
		return err
	}
	calcCore.Run()
	isStop = false
	logger.Infow("calc sdk init success")
	return nil
}

func Fini() {
	isStop = true
	calcCore.Fini()
	if err := logger.Flush(); err != nil {
		fmt.Printf("log flush failed when sdk fini , error = %s\n", err)
	}
}

func paramsCheck(appid, channel, funcs string, c int64) *CalcError {

	if appid == "" {
		return CalcInvalidParams.SetDetail("appid can not be null")
	}

	if channel == "" {
		return CalcInvalidParams.SetDetail("channel can not be null")
	}

	if funcs == "" {
		return CalcInvalidParams.SetDetail("funcs can not be null")
	}

	if c < 0 {
		return CalcInvalidParams.SetDetail("count can not less than 0")
	}
	return nil
}

func generateChannel(cloudId, composeId, serviceId string) (c string, err *CalcError) {
	if cloudId != "" && composeId == "" && serviceId == "" {
		err = CalcInvalidParams
		return
	}
	// 当cloudId为0云或cloudId为空，则均认为是公有云
	if (cloudId == PUBLICCLOUD || cloudId == "") && composeId == "" && serviceId != "" {
		c = serviceId
		return
	}

	c = strings.Join([]string{cloudId, composeId, serviceId}, "@")
	return

}
