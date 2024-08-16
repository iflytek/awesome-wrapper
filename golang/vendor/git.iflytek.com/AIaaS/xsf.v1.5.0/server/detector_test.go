package xsf

import (
	"fmt"
	"git.iflytek.com/AIaaS/xsf/utils"
	"testing"
)

func Test_detector_getAll(t *testing.T) {
	logger, err := utils.NewLocalLog(
		utils.SetCaller(true),
		utils.SetLevel("debug"),
		utils.SetFileName("test.log"),
		utils.SetMaxSize(3),
		utils.SetMaxBackups(3),
		utils.SetMaxAge(3),
		utils.SetAsync(false),
		utils.SetCacheMaxCount(30000),
		utils.SetBatchSize(1024))
	if err != nil {
		panic(err)
	}

	detectorInst, detectorInstErr := newDetector(
		"http://10.1.87.69:6868",
		"guiderAllService",
		"gas",
		"atmos-iat",
		logger,
	)
	if detectorInstErr != nil {
		panic(detectorInstErr)
	}
	fmt.Println(detectorInst.getAll())

}
