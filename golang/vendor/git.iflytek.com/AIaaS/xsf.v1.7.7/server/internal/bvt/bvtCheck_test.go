package bvt

import (
	"fmt"
	"git.iflytek.com/AIaaS/xsf/utils"
	"log"
	"os"
	"testing"
	"time"
)

func Test_bvtVerifier(t *testing.T) {
	var (
		gProject      string = "xsf"
		gGroup        string = "xsf"
		gService      string = "bvt"
		gVersion      string = "1.0.0"
		gCfgFile      string = "bvt.toml"
		gCompanionUrl string = "http://10.1.87.69:6868"
	)
	var (
		timeout        time.Duration = time.Minute
		platformAddr   string        = "http://10.1.87.54:800/api/v1/mission/exec"
		id             string        = "150"
		engIp          string        = "127.0.0.1"
		callback       string        = "http://xxx.xxx.xxx:xxx"
		async          bool          = false
		serviceAddress string        = "http://api.xf-yun.com/v1/private/s67c9c78c"

		licMax  int    = 40
		service string = "fuck"
		namespace string = "789"
	)
	bvtVerifierInst := bvtVerifier{}
	bvtVerifierInst.init(
		gProject,
		gGroup,
		gService,
		gVersion,
		gCfgFile,
		gCompanionUrl,

		timeout,
		platformAddr,
		id,
		engIp,
		callback,
		async,
		serviceAddress,

		licMax,
		service,
		namespace,
	)

	checkErr := bvtVerifierInst.bvtCheck()
	if checkErr != nil {
		log.Fatal(checkErr)
	}
}

func Test_FindManager(t *testing.T) {
	checkErr := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	srvInst, srvInstErr := NewService(CreateCfgOpt(
		utils.WithCfgTick(time.Second),
		utils.WithCfgSessionTimeOut(time.Second),
		utils.WithCfgURL("http://10.1.87.69:6868"),
		utils.WithCfgCachePath("finderCache"),
		utils.WithCfgCacheConfig(true),
		utils.WithCfgCacheService(true),
		utils.WithCfgPrj("xsf"),
		utils.WithCfgGroup("xsf"),
		utils.WithCfgService("bvt"),
		utils.WithCfgVersion("1.0.0"),
	))
	checkErr(srvInstErr)
	cfgContent, cfgContentErr := srvInst.GetRawCfg("bvt.toml")
	checkErr(cfgContentErr)
	fmt.Println(string(cfgContent))
}

func Test_bvtVerifier_checkDeploy(t *testing.T) {

	var (
		gProject      string = "xsf"
		gGroup        string = "xsf"
		gService      string = "bvt"
		gVersion      string = "1.0.0"
		gCfgFile      string = "bvt.toml"
		gCompanionUrl string = "http://10.1.87.69:6868"
	)
	var (
		timeout        time.Duration = time.Minute
		platformAddr   string        = "http://10.1.87.54:800/api/v1/mission/exec"
		id             string        = "150"
		engIp          string        = "127.0.0.1"
		callback       string        = "http://xxx.xxx.xxx:xxx"
		async          bool          = false
		serviceAddress string        = "http://api.xf-yun.com/v1/private/s67c9c78c"

		licMax  int    = 40
		service string = "fuck"
		namespace string = "789"
	)
	bvtVerifierInst := bvtVerifier{}
	bvtVerifierInst.init(
		gProject,
		gGroup,
		gService,
		gVersion,
		gCfgFile,
		gCompanionUrl,

		timeout,
		platformAddr,
		id,
		engIp,
		callback,
		async,
		serviceAddress,

		licMax,
		service,
		namespace,
	)
	t.Log(os.Setenv("DEPLOY_STATUS_API", "172.31.103.99:8085"))
	deployRst := bvtVerifierInst.checkDeploy()
	t.Logf("deployRst:%v\n", deployRst)

}