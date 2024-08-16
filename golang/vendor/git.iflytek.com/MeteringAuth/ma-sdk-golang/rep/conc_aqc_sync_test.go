package rep

import (
	xsf "git.iflytek.com/AIaaS/xsf/client"
	"git.iflytek.com/AIaaS/xsf/utils"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/config"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/syncproto"
	"sync"
	"testing"
)

func TestAqcCounterSyncManager_sendToLksByXsf(t *testing.T) {
	xsfClient, err := xsf.InitClient(
		"test",
		utils.CfgMode(1),
		utils.WithCfgName(config.CfgName),
		utils.WithCfgURL("http://10.1.87.70:6868"),
		utils.WithCfgPrj("AIPaaS"),
		utils.WithCfgGroup("calcaqc002"),
		utils.WithCfgService("janus-client"),
		utils.WithCfgVersion("3.0.0"),
		utils.WithCfgCacheConfig(false),
		utils.WithCfgCacheService(false),
		utils.WithCfgCachePath(config.CfgCacheDir),
		utils.WithCfgCB(func(c *utils.Configure) bool {
			return true
		}),
	)
	if err != nil {
		t.Error(err)
		return
	}
	testm := NewAqcCounterSyncManager(1000, 10000, xsfClient, 1000, "", 1)

	reqs := make(map[string]*syncproto.AqcRequest)
	reqs["test"] = &syncproto.AqcRequest{
		Data: []*syncproto.AqcMetadata{
			{
				Addr: "127.0.0.1",
				Tuple: &syncproto.MetaTuple{
					AppId:    "test1",
					Channel:  "test1",
					Function: "test1",
				},
			},
		},
	}
	reqs["test2"] = &syncproto.AqcRequest{
		Data: []*syncproto.AqcMetadata{
			{
				Addr: "127.0.0.2",
				Tuple: &syncproto.MetaTuple{
					AppId:    "test2",
					Channel:  "test2",
					Function: "test2",
				},
			},
		},
	}
	testm.sendToLksByXsf(reqs, "test")
}

func TestNewAqcCounterSyncManager(t *testing.T) {
	xsfClient, err := xsf.InitClient(
		"test",
		utils.CfgMode(1),
		utils.WithCfgName(config.CfgName),
		utils.WithCfgURL("http://10.1.87.70:6868"),
		utils.WithCfgPrj("AIPaaS"),
		utils.WithCfgGroup("calcaqc002"),
		utils.WithCfgService("janus-client"),
		utils.WithCfgVersion("3.0.0"),
		utils.WithCfgCacheConfig(false),
		utils.WithCfgCacheService(false),
		utils.WithCfgCachePath(config.CfgCacheDir),
		utils.WithCfgCB(func(c *utils.Configure) bool {
			return true
		}),
	)
	if err != nil {
		t.Error(err)
		return
	}
	testm := NewAqcCounterSyncManager(1000, 10000, xsfClient, 1000, "", 5)
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		go func() {
			wg.Add(1)
			for {
				testm.Add(aQcSyncCounterKey{
					counterKey: counterKey{
						appId:    "test",
						channel:  "test",
						function: "test",
					},
					addr: "127.0.0.1",
				}, 1)
			}
			//wg.Done()
		}()
	}
	//time.Sleep(1000 * time.Second)
	wg.Wait()
}
