package utils

import (
	"fmt"
	"testing"
)

func TestNewCfgWithBytes(t *testing.T) {
	cfgData := `[xats] 
finder = 1
port = 5010
report = 1

[local]
max_audio_time = 60

[pprof]
enable = 1
port = 12341

[aires]
HMM_16K = "/msp/resource/sms/xfime/hmm_16k_10ab.bin"
WFST =  "/msp/resource/sms/xfime/wfst_a4ca.bin"
LM = "/msp/resource/sms/xfime/lm_e552.bin"
RLM = "/msp/resource/sms/xfime/rlm_1ca0.bin"
WFST_PGS = "/msp/resource/sms/xfime/wfst_pgs_7126.bin"
HMM_16K_PGS = "/msp/resource/sms/xfime/hmm_16k_pgs_fc56.bin"

[domain]
model1 = "game;WFST_PATCH;/msp/resource/sms/xfime/domain/wfst_patch_game_56cd.bin"
model2 = "game;LM_PERSONAL;/msp/resource/sms/xfime/domain/lm_personal_game_4c88.bin"
model3 = "health;WFST_PATCH;/msp/resource/sms/xfime/domain/wfst_patch_medical_30d1.bin"
model4 = "health;LM_PERSONAL;/msp/resource/sms/xfime/domain/lm_personal_medical_1762.bin"
model5 = "shopping;WFST_PATCH;/msp/resource/sms/xfime/domain/wfst_patch_shopping_41da.bin"
model6 = "shopping;LM_PERSONAL;/msp/resource/sms/xfime/domain/lm_personal_shopping_49f0.bin"
model7 = "trip;WFST_PATCH;/msp/resource/sms/xfime/domain/wfst_patch_travelpart_96f0.bin"
model8 = "trip;LM_PERSONAL;/msp/resource/sms/xfime/domain/lm_personal_travelpart_796e.bin"

[aiges] 
userHotwordCacheSize = 1000
libCodec = "libamr.so;libamr_wb.so;libspeex.so;libico.so;libopus.so;libict.so"
#enablehbase = 0
sessMode = 1
numaNode = 0
elogRemote = 0
elogLocal = 1
elogLocalLog = 1
elogHost = "10.1.86.76"
elogPort = "4545"
elogXml = "seelog.xml"
elogSpill = "/log/server/iatspill"
elogS3ak = "W3SBNK9CAP32EJX2E531"
elogS3sk = "dXaAFrtOfsmsVjmbzUEOe7RrnUkNHBoGClw55WW2"
elogS3ep = "oss-shanghai-js.openstorage.cn"
elogZkhost = "10.101.5.38,10.101.5.32,10.101.5.26,10.101.5.19,10.101.5.15"
elogConsumerNum = 10

[log]
level = "debug"
file = "/log/server/xats/sms_125.log"
size = 100
count = 100
die = 30
async = 0

[trace]
host = "10.1.86.76"
port = 4545 #缺省4545
able = 1
dump = 0
backend = 2
buffer=2000
#是否将日志写入到远端
deliver = 1 #缺省1
#taddrs="iat@10.1.205.151:50051;10.1.205.151:50052,tts@10.1.205.151:50052;10.1.205.151:50051"
loadts = 1000000

[lb] #xrpc loadReporter,#fc的strategy=2时候，这个section无效
able           = 0 #缺省0


[lbv2]
lbname = "lbv2-iat"    #传给服务发现，用来确定使用哪个lb       程序启动时候-m=1表示使用服务发现来寻找Lb
able = 1
sub = "iat"
subsvc = "sms"
apiversion = "1.0.0"
conn-timeout = 100
conn-pool-size = 20        #rpc连接池数量。缺省4
lb-mode= 0  #0禁用lb,2使用lb。缺省0      ats只是上报lb,不使用lb，所以设为0
lb-retry = 2
#taddrs="lbv2@10.1.87.61:9095"
taddrs="lbv2@10.1.87.68:9095"    #不使用服务发现时候使用的默认lb地址


[fc] #xrpc flowControl
able = 1
router = "sessionManager"   	#路由字段，可选项为sessionManager和qpsLimiter
max = 220                  	#会话模式时代表最大的授权量，非会话模式代表间隔时间里的最大请求数
ttl = 10000                  	#会话模式代表会话的超时时间，非会话模式代表有效期（间隔时间）
best = 220                   	#最佳授权数
roll = 5000
strategy = 2                 #可选值为0、1、2（缺省为0），0.代表定时上报(v1)；1.根据授权范围上报(v1)；2.基于hermes（v2）
`
	cfg, cfgErr := NewCfgWithBytes(cfgData)
	if cfgErr != nil {
		panic(cfgErr)
	}
	fmt.Println(cfg.GetInt("fc", "max"))
}
