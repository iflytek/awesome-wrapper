package calc

import (
	"errors"
	"strings"

	"git.iflytek.com/MeteringAuth/ma-sdk-golang/calc/core"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/calc/monitor"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/calc/ver"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/config"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/tool"
)

const PUBLICCLOUD = "0"

var (
	calcCore *core.CalcCore = &core.CalcCore{}
	inited   bool
	started  bool
)

// 输入包含具体的用户参数信息，appid , channel , funcs不可为空，c 不可小于等于0
func Calc(appid, channel, funcs string, c int64) (errorcode int, err error) {
	if !started {
		tool.L.Warnw("calc-sdk | sdk have not been started")
		errorcode, err = config.CalcStoped.GetCode(), config.CalcStoped
		return
	}

	if c == 0 {
		tool.L.Warnw("calc-sdk | invalid parameters", "desc", "c can not be 0")
		errorcode, err = config.CalcInvalidParams.GetCode(), config.CalcInvalidParams
		return
	}

	if calcErr := paramsCheck(appid, channel, funcs, c); calcErr != nil {
		tool.L.Warnw("calc-sdk | invalid parameters", "desc", calcErr.Error())
		errorcode, err = calcErr.GetCode(), calcErr
	}
	if calcErr := calcCore.Set(appid, "", channel, funcs, int(c)); calcErr != nil {
		tool.L.Warnw("calc-sdk | calcCore set error", "desc", calcErr.Error())
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
	var calcErr *config.CalcError
	channel, calcErr := generateChannel(cloudid, composeid, serviceid)
	if calcErr != nil {
		return calcErr.GetCode(), calcErr
	}

	return Calc(appid, channel, funcs, c)
}

func CalWithSubIdMultiTenancy(appid, subId, cloudid, composeid, serviceid, funcs string, c int64) (errorcode int, err error) {
	var calcErr *config.CalcError
	channel, calcErr := generateChannel(cloudid, composeid, serviceid)
	if calcErr != nil {
		return calcErr.GetCode(), calcErr
	}

	return CalcWithSubId(appid, subId, channel, funcs, c)
}

func Init(url, pro, gro, service, version string, mode int) (err error) {
	if inited {
		return errors.New("already inited")
	}
	inited = true

	tool.CalcPrinter.Println("version:", ver.Version)
	tool.CalcPrinter.Println("config:", url, pro, gro, service, version, mode)

	if err = tool.Init(url, pro, gro, service, version, mode, 0); err != nil {
		return
	}

	if config.C.Metrics.Able == 1 {
		err = monitor.Init()
		if err != nil {
			return
		}
	}

	asyncInit := config.C.Calc.AsyncInit
	retryTimes := config.C.Calc.InitRetry
	if asyncInit {
		tool.CalcPrinter.Println("async init")
		go func() {
			lazyInit(url, pro, gro, service, version, mode, retryTimes)
		}()
	} else {
		tool.CalcPrinter.Println("sync init")
		err = lazyInit(url, pro, gro, service, version, mode, retryTimes)
	}

	return
}

func lazyInit(url, pro, gro, service, version string, mode int, retry int) (err error) {
	err = initOnce(url, pro, gro, service, version, mode)
	for i := 1; i < retry && err != nil; i++ {
		err = initOnce(url, pro, gro, service, version, mode)
	}
	if err != nil {
		tool.CalcPrinter.Println("failed to init, retry times:", retry, "error:", err)
	} else {
		tool.CalcPrinter.Println("init done")
		tool.L.Infow("calc-sdk | calc sdk init success")
		started = true
	}

	return
}

func initOnce(url, pro, gro, service, version string, mode int) (err error) {
	if err = calcCore.Init(); err != nil {
		return err
	}
	calcCore.Run()

	return
}

func Fini() {
	started = false
	calcCore.Fini()
}

func paramsCheck(appid, channel, funcs string, c int64) *config.CalcError {
	if appid == "" {
		return config.CalcInvalidParams.SetDetail("appid can not be null")
	}

	if channel == "" {
		return config.CalcInvalidParams.SetDetail("channel can not be null")
	}

	if funcs == "" {
		return config.CalcInvalidParams.SetDetail("funcs can not be null")
	}

	if c < 0 {
		return config.CalcInvalidParams.SetDetail("count can not less than 0")
	}
	return nil
}

func generateChannel(cloudId, composeId, serviceId string) (c string, err *config.CalcError) {
	if cloudId != "" && composeId == "" && serviceId == "" {
		err = config.CalcInvalidParams
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
