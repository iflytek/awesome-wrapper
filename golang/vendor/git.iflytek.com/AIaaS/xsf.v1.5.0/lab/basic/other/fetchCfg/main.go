package main

import (
	"fmt"
	finder "git.iflytek.com/AIaaS/finder-go/common"
	"git.iflytek.com/AIaaS/xsf/utils"
)

type cfg struct {
}

func (c *cfg) OnConfigFileChanged(_ *finder.Config) bool {
	fmt.Println("OnConfigFileChanged")
	return true
}
func (c *cfg) OnError(_ finder.ConfigErrInfo) {
	fmt.Println("OnError")
}

func main() {
	{
		//初始化configurator
		logCfgOpt := &utils.CfgOption{}
		utils.WithCfgVersion("2.2.7")(logCfgOpt)
		utils.WithCfgPrj("guiderAllService")(logCfgOpt)
		utils.WithCfgGroup("gas")(logCfgOpt)
		utils.WithCfgService("lbv2")(logCfgOpt)
		utils.WithCfgName("lbv2.toml")(logCfgOpt)
		utils.WithCfgURL("http://10.1.87.69:6868")(logCfgOpt)
		cfg, err := utils.NewCfg(utils.Centre, logCfgOpt)

		if err != nil {
			panic("xxx")
		}
		fmt.Println(cfg.GetRawCfg())
	}
}