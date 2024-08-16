package tool

import (
	xsf "git.iflytek.com/AIaaS/xsf/client"
	"git.iflytek.com/AIaaS/xsf/utils"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/config"
	"sort"
)

func NewClient(cname string) (*xsf.Client, error) {
	cfg := config.Params
	return xsf.InitClient(
		cname,
		utils.CfgMode(cfg.M),
		utils.WithCfgName(config.CfgName),
		utils.WithCfgURL(cfg.U),
		utils.WithCfgPrj(cfg.P),
		utils.WithCfgGroup(cfg.G),
		utils.WithCfgService(cfg.S),
		utils.WithCfgVersion(cfg.V),
		utils.WithCfgCacheConfig(true),
		utils.WithCfgCacheService(true),
		utils.WithCfgCachePath(config.CfgCacheDir),
		utils.WithCfgCB(func(c *utils.Configure) bool {
			return true
		}),
	)
}

func IN(target string, str_array []string) bool {
	sort.Strings(str_array)
	index := sort.SearchStrings(str_array, target)
	//index的取值：[0,len(str_array)]
	if index < len(str_array) && str_array[index] == target { //需要注意此处的判断，先判断 &&左侧的条件，如果不满足则结束此处判断，不会再进行右侧的判断
		return true
	}
	return false
}
