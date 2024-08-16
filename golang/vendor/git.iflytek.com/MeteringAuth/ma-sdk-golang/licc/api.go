package licc

import (
	"errors"
	"strings"
	"time"

	"git.iflytek.com/MeteringAuth/ma-sdk-golang/licc/acflimit"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/licc/licc"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/licc/lmt"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/licc/monitor"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/licc/ver"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/config"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/tool"
)

var (
	ErrParam = errors.New("check params invalid")
)

const (
	PUBLICCLOUD = "0"
)

var (
	// 受限资源管理类
	l *lmt.LimitFuncsManager
	// 鉴权类
	a *licc.AuthenticateCheck
	// acf limit checker
	alc *acflimit.AcfLimitCheck
)

var (
	inited  bool
	started bool
)

// 服务发现相关的配置项 , 初始化SDK时需要传入
func Init(url, pro, gro, service, version string, mode int, channel []string) (err error) {
	if inited {
		return errors.New("already inited")
	}
	inited = true

	tool.LiccPrinter.Println("licc-sdk version:", ver.Version)
	tool.LiccPrinter.Println("licc-sdk config:", url, pro, gro, service, version, mode)

	if err := tool.Init(url, pro, gro, service, version, mode, 0); err != nil {
		return err
	}

	if config.C.Metrics.Able == 1 {
		err = monitor.Init()
		if err != nil {
			return
		}
	}

	asyncInit := config.C.Licc.AsyncInit
	retryTimes := config.C.Licc.InitRetry
	if asyncInit {
		tool.LiccPrinter.Println("async init")
		go func() {
			lazyInit(url, pro, gro, service, version, mode, channel, retryTimes)
		}()
	} else {
		tool.LiccPrinter.Println("sync init")
		err = lazyInit(url, pro, gro, service, version, mode, channel, retryTimes)
	}
	return
}

func lazyInit(url, pro, gro, service, version string, mode int, channel []string, retry int) (err error) {
	err = initOnce(url, pro, gro, service, version, mode, channel)
	for i := 1; i < retry && err != nil; i++ {
		err = initOnce(url, pro, gro, service, version, mode, channel)
	}
	if err != nil {
		tool.LiccPrinter.Println("failed to init, retry times:", retry, "error:", err)
	} else {
		tool.LiccPrinter.Println("init done")
		tool.L.Infow("rep-sdk | init success")
		started = true
	}
	return
}

func initOnce(url, pro, gro, service, version string, mode int, channel []string) (err error) {
	cfg := config.C.Licc
	tool.LiccPrinter.Printf("init licc-sdk cfg: %+v\n", cfg)

	l, err = lmt.NewLimitFuncsManager(channel)
	if err != nil {
		time.Sleep(time.Second)
		l, err = lmt.NewLimitFuncsManager(channel)
		if err != nil {
			return
		}
	}
	tool.LiccPrinter.Println(l)

	a, err = licc.NewCheckLicManager()
	if err != nil {
		time.Sleep(time.Second)
		a, err = licc.NewCheckLicManager()
		if err == nil {
			return
		}
	}
	tool.LiccPrinter.Println(a)

	alc, err = acflimit.NewAcfLimitCheckManager()
	if err != nil {
		time.Sleep(time.Second)
		alc, err = acflimit.NewAcfLimitCheckManager()
		if err != nil {
			return err
		}
	}
	tool.LiccPrinter.Println(alc)

	return
}

/*
ret = 0 , limitInfo=map[string]string{"add limitFuncs [test2]":"11200"} , err = <nil>

ret = 0 , limitInfo=map[string]string(nil) , err = <nil>
*/
func Check(appid string, uid string, channel string, funcs []string, tag string) (authInfo map[string]string, logInfo string, err error) {
	if !started {
		tool.L.Errorw("licc-sdk | not started")
		return
	}

	// 合成业务已经支持了用户级鉴权，且其uid是通过auth_id登录而来，能保证在语音云系统中的唯一性。
	// 因数据库中已经有了合成业务的用户授权数据，故无法统一至子ID生成规则中，
	// 为了避免在业务层面上对此做兼容，特在sdk中做此兼容
	if channel != "tts" {
		uid = strings.Join([]string{"SUBID", appid, uid}, "_")
	}
	// filter nolimit resource
	if l != nil && a != nil {
		start := time.Now()
		checkFuncs := strings.Join(l.Filter(channel, funcs), ";")
		authInfo, logInfo, err = a.HasLicense(appid, uid, channel, checkFuncs, licc.CtrlNone, tag)
		cost := time.Since(start)
		monitor.WithCost("check", cost)
	} else {
		return nil, logInfo, errors.New("janus client have not inited")
	}
	return
}

func CheckEx(appid, uid, cloudid, composeid, serviceid string, funcs []string, mode licc.CtrlMode, tag string) (authInfo map[string]string, logInfo string, err error) {
	if !started {
		tool.L.Errorw("licc-sdk | not started")
		return
	}

	channel, err := generateChannel(cloudid, composeid, serviceid)
	if err != nil {
		return nil, "", err
	}

	// 合成业务已经支持了用户级鉴权，且其uid是通过auth_id登录而来，能保证在语音云系统中的唯一性。
	// 因数据库中已经有了合成业务的用户授权数据，故无法统一至子ID生成规则中，
	// 为了避免在业务层面上对此做兼容，特在sdk中做此兼容
	if channel != "tts" {
		uid = strings.Join([]string{"SUBID", appid, uid}, "_")
	}

	// filter nolimit resource
	if l != nil && a != nil {
		start := time.Now()
		checkFuncs := strings.Join(l.Filter(channel, funcs), ";")
		authInfo, logInfo, err = a.HasLicense(appid, uid, channel, checkFuncs, mode, tag)
		cost := time.Since(start)
		monitor.WithCost("checkEx", cost)
	} else {
		return nil, logInfo, errors.New("janus client have not inited")
	}
	return
}

func CheckMultiTenancy(appid, uid, cloudid, composeid, serviceid string, funcs []string, tag string) (authInfo map[string]string, logInfo string, err error) {
	if !started {
		tool.L.Errorw("licc-sdk | not started")
		return
	}

	channel, err := generateChannel(cloudid, composeid, serviceid)
	if err != nil {
		return nil, "", err
	}
	return Check(appid, uid, channel, funcs, tag)
}

// 逆初始化
func Fini() (err error) {
	return
}

func generateChannel(cloudId, composeId, serviceId string) (c string, err error) {
	if cloudId != "" && composeId == "" && serviceId == "" {
		err = ErrParam
		return
	}
	if (cloudId == PUBLICCLOUD || cloudId == "") && composeId == "" && serviceId != "" {
		c = serviceId
		return
	}

	c = strings.Join([]string{cloudId, composeId, serviceId}, "@")
	return

}

func GetAcfLimits(appid, channel string, funcs []string, tag string) (map[string]string, error) {
	if !started {
		tool.L.Errorw("licc-sdk | not started")
		return nil, nil
	}

	return alc.GetAcfLimits(appid, channel, strings.Join(l.Filter(channel, funcs), ";"), tag)
}

func GetAcfLimitsMT(appid, cloudid, composeid, serviceid string, funcs []string, tag string) (map[string]string, error) {
	if !started {
		tool.L.Errorw("licc-sdk | not started")
		return nil, nil
	}

	channel, err := generateChannel(cloudid, composeid, serviceid)
	if err != nil {
		return nil, err
	}
	return GetAcfLimits(appid, channel, funcs, tag)
}
