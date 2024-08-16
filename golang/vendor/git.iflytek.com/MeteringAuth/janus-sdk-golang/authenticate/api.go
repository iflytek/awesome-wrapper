package authenticate

import (
	"strings"

	"errors"
	"fmt"
	"time"

	"git.iflytek.com/MeteringAuth/janus-sdk-golang/ver"
	xsf "git.xfyun.cn/AIaaS/xsf/client"
	"git.xfyun.cn/AIaaS/xsf/utils"
)

type CtrlMode uint16

var (
	JanusParamsError = errors.New("check params invalid")
)

const (
	//
	CtrlNone = iota

	//
	CtrlDayFlow CtrlMode = 1 << (iota - 1) // 日流控
	//CtrlHourFlow                                  // 小时流控
	CtrlTimeFlow  // 时授
	CtrlCountFlow // 量授
	CtrlUserFlow  // 用户级
	CtrlFreeFlow  // 免费次数
	CtrlSecFlow   // 秒级流控
	CtrlConcFlow  // 并发
	CtrlTail

	// 覆盖所有类型位（用于通过&获取类型位）
	CtrlALL = CtrlDayFlow | CtrlSecFlow | CtrlCountFlow | CtrlConcFlow | CtrlUserFlow | CtrlFreeFlow
)

const PUBLICCLOUD = "0"

var (
	APP_CONFIG     = "janus-client.toml"
	CnameJanus     = "janus-limit-func"
	CnameAcfJanus  = "janus-acf-limit"
	CnameHashJanus = "janus-check-func"
	SVC            = "janus"
	CheckOption    = CtrlALL

	DefaultInitOption = &InitOption{
		companionUrl:   "http://companion.xfyun.iflytek:6868",
		project:        "AIaaS",
		group:          "dx",
		service:        "janus",
		version:        ver.Version,
		isCacheService: false,
		isCacheConfig:  false,
		cachePath:      "./janus-sdk-cache",
		cfgMode:        1,
	}
)

var (
	// 受限资源管理类
	l *LimitFuncsManager
	// 鉴权类
	a *AuthenticateCheck
	// acf limit checker
	alc *AcfLimitCheck
)

// 服务发现相关的配置项 , 初始化SDK时需要传入
func Init(channel []string) (err error) {
	// 初始化两个xsfClient，分别用于获取受限制资源的定时请求（RobinRound策略）与 鉴权请求（一致性哈希策略）
	/*init polling  xsf*/
	/*get limit funcs */

	l, err = NewLimitFuncsManager(channel, APP_CONFIG, CnameJanus)
	if err != nil {
		go func() {
			t := time.NewTicker(1 * time.Second)
			//	defer t.Stop()
			for {
				select {
				case <-t.C:
					l, err = NewLimitFuncsManager(channel, APP_CONFIG, CnameJanus)
					if err == nil {
						return
					}
				}
			}
		}()
		return
	}
	a, err = NewCheckLicManager(APP_CONFIG, CnameHashJanus)
	if err != nil {
		go func() {
			t := time.NewTicker(1 * time.Second)
			//	defer t.Stop()
			for {
				select {
				case <-t.C:
					a, err = NewCheckLicManager(APP_CONFIG, CnameHashJanus)
					if err == nil {
						return
					}
				}
			}
		}()
		return
	}
	alc, err = NewAcfLimitCheckManager(APP_CONFIG, CnameAcfJanus)
	if err != nil {
		return fmt.Errorf("acflimit init error:%v", err)
	}
	l.xsfc.Log.Infof("init janus sdk success\n")
	return
}

/*
ret = 0 , limitInfo=map[string]string{"add limitFuncs [test2]":"11200"} , err = <nil>

ret = 0 , limitInfo=map[string]string(nil) , err = <nil>
*/
func Check(appid string, uid string, channel string, funcs []string) (authInfo map[string]string, logInfo string, err error) {
	// 合成业务已经支持了用户级鉴权，且其uid是通过auth_id登录而来，能保证在语音云系统中的唯一性。
	// 因数据库中已经有了合成业务的用户授权数据，故无法统一至子ID生成规则中，
	// 为了避免在业务层面上对此做兼容，特在sdk中做此兼容
	if channel != "tts" {
		uid = strings.Join([]string{"SUBID", appid, uid}, "_")
	}
	// filter nolimit resource
	if l != nil && a != nil {
		checkFuncs := strings.Join(l.filter(channel, funcs), ";")
		authInfo, logInfo, err = a.HasLicense(appid, uid, channel, checkFuncs, CheckOption)
	} else {
		return nil, logInfo, errors.New("janus client have not inited")
	}
	return
}

func CheckEx(appid, uid, cloudid, composeid, serviceid string, funcs []string, mode CtrlMode) (authInfo map[string]string, logInfo string, err error) {
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
		checkFuncs := strings.Join(l.filter(channel, funcs), ";")
		authInfo, logInfo, err = a.HasLicense(appid, uid, channel, checkFuncs, mode)
	} else {
		return nil, logInfo, errors.New("janus client have not inited")
	}
	return
}

func CheckMultiTenancy(appid, uid, cloudid, composeid, serviceid string, funcs []string) (authInfo map[string]string, logInfo string, err error) {
	channel, err := generateChannel(cloudid, composeid, serviceid)
	if err != nil {
		return nil, "", err
	}
	return Check(appid, uid, channel, funcs)
}

//逆初始化
func Fini() (err error) {
	return
}

func generateChannel(cloudId, composeId, serviceId string) (c string, err error) {
	if cloudId != "" && composeId == "" && serviceId == "" {
		err = JanusParamsError
		return
	}
	if (cloudId == PUBLICCLOUD || cloudId == "") && composeId == "" && serviceId != "" {
		c = serviceId
		return
	}

	c = strings.Join([]string{cloudId, composeId, serviceId}, "@")
	return

}

func newXsfClient(cfgName, cname string) (xsfClient *xsf.Client, err error) {
	xsfClient, err = xsf.InitClient(
		cname,
		utils.CfgMode(DefaultInitOption.cfgMode),
		utils.WithCfgName(cfgName),
		utils.WithCfgURL(DefaultInitOption.companionUrl),
		utils.WithCfgPrj(DefaultInitOption.project),
		utils.WithCfgGroup(DefaultInitOption.group),
		utils.WithCfgService(DefaultInitOption.service),
		utils.WithCfgVersion(DefaultInitOption.version),
		utils.WithCfgCacheService(DefaultInitOption.isCacheService),
		utils.WithCfgCacheConfig(DefaultInitOption.isCacheConfig),
		utils.WithCfgCachePath(DefaultInitOption.cachePath),
		utils.WithCfgCB(func(c *utils.Configure) bool {
			coption, err := c.GetInt(CnameHashJanus, "check_option")
			if err == nil {
				CheckOption = CtrlMode(uint32(coption))
				fmt.Println("cfg changed <check_option>:", CheckOption)
				return true
			}
			return false
		}),
	)
	if err != nil {
		return
	}
	return
}

type InitOption struct {
	// 服务发现地址
	companionUrl   string
	project        string
	group          string
	service        string
	version        string
	isCacheService bool
	isCacheConfig  bool
	cachePath      string
	cfgMode        int
}

func SetCompanionUrl(url string) *InitOption {
	DefaultInitOption.companionUrl = url
	return DefaultInitOption
}

func SetProjectName(projectName string) *InitOption {
	DefaultInitOption.project = projectName
	return DefaultInitOption
}

func SetGroup(group string) *InitOption {
	DefaultInitOption.group = group
	return DefaultInitOption
}

func SetServiceName(service string) *InitOption {
	DefaultInitOption.service = service
	return DefaultInitOption
}

func SetVersion(ver string) *InitOption {
	DefaultInitOption.version = ver
	return DefaultInitOption
}

func SetCacheService(ok bool) *InitOption {
	DefaultInitOption.isCacheService = ok
	return DefaultInitOption
}

func SetCacheConfig(ok bool) *InitOption {
	DefaultInitOption.isCacheConfig = ok
	return DefaultInitOption
}

func SetCfgMode(m int) *InitOption {
	DefaultInitOption.cfgMode = m
	return DefaultInitOption
}

func (i *InitOption) SetCompanionUrl(url string) *InitOption {
	DefaultInitOption.companionUrl = url
	return DefaultInitOption
}

func (i *InitOption) SetProjectName(projectName string) *InitOption {
	DefaultInitOption.project = projectName
	return DefaultInitOption
}

func (i *InitOption) SetGroup(group string) *InitOption {
	DefaultInitOption.group = group
	return DefaultInitOption
}

func (i *InitOption) SetServiceName(service string) *InitOption {
	DefaultInitOption.service = service
	return DefaultInitOption
}

func (i *InitOption) SetVersion(ver string) *InitOption {
	DefaultInitOption.version = ver
	return DefaultInitOption
}

func (i *InitOption) SetCacheService(ok bool) *InitOption {
	DefaultInitOption.isCacheService = ok
	return DefaultInitOption
}

func (i *InitOption) SetCacheConfig(ok bool) *InitOption {
	DefaultInitOption.isCacheConfig = ok
	return DefaultInitOption
}

func (i *InitOption) SetCfgMode(m int) *InitOption {
	DefaultInitOption.cfgMode = m
	return DefaultInitOption
}

func GetAcfLimits(appid, channel string, funcs []string) (map[string]string, error) {
	if alc != nil && l != nil {
		return alc.getAcfLimits(appid, channel, strings.Join(l.filter(channel, funcs), ";"))
	}
	return nil, errors.New("sdk not init")
}

func GetAcfLimitsMT(appid, cloudid, composeid, serviceid string, funcs []string) (map[string]string, error) {
	channel, err := generateChannel(cloudid, composeid, serviceid)
	if err != nil {
		return nil, err
	}
	return GetAcfLimits(appid, channel, funcs)
}
