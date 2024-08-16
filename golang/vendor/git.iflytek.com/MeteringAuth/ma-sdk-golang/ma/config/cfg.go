package config

import (
	"fmt"
	"os"
	"time"

	"git.iflytek.com/AIaaS/finder-go"
	common "git.iflytek.com/AIaaS/finder-go/common"
	cp "git.iflytek.com/HY_trainee/colorPrinter"
)

const (
	CfgName = "ma-sdk.toml"
	SVC     = "janus"

	// xsfc name
	CliLiccTag = "licc"
	CliLmtTag  = "lmtres"
	CliRepTag  = "rep"

	//xsfc cfg
	CliDurTag    = "duration"
	CliLiccOpTag = "check_opnion"
	CliLmtUpTag  = "update_time"

	//cfg
	CfgCacheDir = "./ma-sdk-cfg"
)

type params struct {
	U string
	P string
	G string
	S string
	V string
	M int
}

var Params = params{}

func InitCfg(url, pro, gro, service, version string, mode int) (err error) {
	switch mode {
	case 0, 2:
		{
			data, err := os.ReadFile(CfgName)
			if err != nil {
				panic(err)
			}
			if err := initCfg(data); err != nil {
				panic(err)
			}
			sp.Println("init", CfgName, "done")
		}
	case 1:
		{
			config := common.BootConfig{
				//companion地址
				CompanionUrl: url,
				//缓存路径
				CachePath: CfgCacheDir,
				//是否缓存服务信息
				CacheService: true,
				//是否缓存配置信息
				CacheConfig:   true,
				ExpireTimeout: 5 * time.Second,
				MeteData: &common.ServiceMeteData{
					Project: pro,
					Group:   gro,
					Service: service,
					Version: version,
				},
			}
			if f, err := finder.NewFinderWithLogger(config, nil); err != nil {
				panic(err)
			} else {
				if cfgdata, err := f.ConfigFinder.UseAndSubscribeConfig([]string{CfgName}, &cfgCB{}); err != nil {
					panic(err)
				} else {
					if err := initCfg(cfgdata[CfgName].File); err != nil {
						panic(err)
					}
					sp.Println("init", CfgName, "done")
				}
			}
		}
	default:
		return fmt.Errorf("mode error")
	}

	Params.U = url
	Params.G = gro
	Params.P = pro
	Params.S = service
	Params.V = version
	Params.M = mode
	sp.Printf("parmas: %#v\n", Params)
	return err
}

type cfgCB struct {
}

func (*cfgCB) OnConfigFileChanged(config *common.Config) bool {
	if err := initCfg(config.File); err != nil {
		sp.Println("update error", err)
		return false
	}
	return true
}

func (*cfgCB) OnConfigFilesAdded(map[string]*common.Config) bool { return true }

func (*cfgCB) OnConfigFilesRemoved([]string) bool { return true }

func (*cfgCB) OnError(common.ConfigErrInfo) {}

var sp = cp.NewctPrinter("ma-sdk-config", cp.Red)
