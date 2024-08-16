package report

import (
	"errors"
	"time"

	"git.iflytek.com/MeteringAuth/janus-sdk-golang/ver"
)

var (
	CfgName           = "janus-report.toml"
	ClientName        = "janus-report"
	r                 *ReportManager
	start             bool
	DefaultInitOption = &InitOption{
		companionUrl:   "http://companion.xfyun.iflytek:6868",
		project:        "AIaaS",
		group:          "dx",
		service:        "janus-sdk",
		version:        ver.Version,
		isCacheService: false,
		isCacheConfig:  false,
		cachePath:      "./janus-sdk-cache",
		cfgMode:        1,
	}
)

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

/*addr ip:port*/
func Init(addr string) (err error) {
	if addr == "" {
		r.xsfc.Log.Errorf("addr is nil")
		return
	}
	r, err = newReportManger(addr)
	if err != nil {
		return
	}
	r.xsfc.Log.Infof("init janus sdk success\n")
	start = true
	return

}

func Fini() {
	if !start {
		return
	}
	start = false
	time.Sleep(time.Second)
	r.fini()
}

/*appid,auth count*/
func Report(channel string, concInfo map[string]uint) (err error) {
	if !start {
		return errors.New("Report have not inited")
	}
	r.report(channel, concInfo, false)
	return
}

// func ReportEnt(channel string, concInfo map[string]uint) (err error) {
// 	if !start {
// 		return errors.New("Report have not inited")
// 	}

// 	r.report(channel, concInfo, true)
// 	return
// }

func Sync(concInfo map[string]string, channel, endpoint string) (err error) {
	if !start {
		return errors.New("Report have not inited")
	}
	r.sync(concInfo, channel, endpoint, false)
	return
}

// func SyncEnt(concInfo map[string]string, channel, endpoint string) (err error) {
// 	if !start {
// 		return errors.New("Report have not inited")
// 	}
// 	r.sync(concInfo, channel, endpoint, true)
// 	return
// }

func ReportWithAddr(channel string, concInfo map[string]uint, addr string) (err error) {
	if !start {
		return errors.New("Report have not inited")
	}
	r.reportWithAddr(channel, concInfo, addr, false)
	return
}

// func ReportWithAddrEnt(channel string, concInfo map[string]uint, addr string) (err error) {
// 	if !start {
// 		return errors.New("Report have not inited")
// 	}
// 	r.reportWithAddr(channel, concInfo, addr, true)
// 	return
// }

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
