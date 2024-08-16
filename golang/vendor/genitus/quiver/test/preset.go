package test

import (
	"genitus/quiver"
	"fmt"
	"time"
)

func GetEvent() *quiver.EventData {
	sid := fmt.Sprintf("iam56eb0006@lc%x3319210", time.Now().UnixNano()/1000/1000)
	eventSSB := quiver.NewEventWithNamePort(quiver.TYPE_IAT, "jyjhf-51-003", "iat", "9092").
		WithUid("v1042331810").
		WithSid(sid).
		WithSyncId(0).
		WithSub("iat-golang").
		WithName("ssb").
		WithEndpoint("172.26.5.200")

	eventSSB.Tag(quiver.KV{"login_id", "v1042331810@100IME"}).
		Tag(quiver.KV{"is_open", true}).
		Tag(quiver.KV{"socker_id", 8477}).
		Tag(quiver.KV{"params", "sub=iat, auf=audio/L16; rate=16000, ssm=1, cver=5.0.24.1137"}).
		Tag(quiver.KV{"nginx_ip", "172.27.131.11:26559"}).
		Tag(quiver.KV{"lang", ">>>>golang<<<<"}).
		TagDS("lc")

	eventSSB.Desc("req=call `msp_user_ent_get`(`1042331810`, `cantonese16k`)").
		Desc("req-resId:50 resType:WFST resId:S1 resType:HMM_16K resId:S2 resType:HMM_16K").
		Descf("req:%d, res:%s, %f", 50, "123", 16.22)

	eventSSB.Desc("audioLen: 488, audioStatus: 2, errNum: 0").
		Desc("result: {\"sn\":1,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"apd\",\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]}]}").
		Desc("audioLen: 488, audioStatus: 2, errNum: 0").
		Desc("result: {\"sn\":2,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,1],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"1385\"}]}]}").
		Desc("result: {\"sn\":3,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,2],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"13856\"}]}]}").
		Desc("audioLen: 488, audioStatus: 2, errNum: 0").
		Desc("result: {\"sn\":4,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,3],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"138569\"}]}]}").
		Desc("result: {\"sn\":5,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,4],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"1385690\"}]}]}").
		Desc("audioLen: 488, audioStatus: 2, errNum: 0").
		Desc("audioLen: 488, audioStatus: 2, errNum: 0").
		Desc("result: {\"sn\":6,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,5],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"13856901\"}]}]}").
		Desc("result: {\"sn\":7,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,6],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"138569012\"}]}]}").
		Desc("audioLen: 488, audioStatus: 2, errNum: 0").
		Desc("result: {\"sn\":8,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,7],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"1385690123\"}]}]}").
		Desc("result: {\"sn\":9,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,8],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"13856901234\"}]}]}").
		Desc("audioLen: 488, audioStatus: 2, errNum: 0").
		Desc("result: {\"sn\":10,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,9],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"13856901234\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"充\"}]}]}").
		Desc("audioLen: 488, audioStatus: 2, errNum: 0").
		Desc("audioLen: 488, audioStatus: 2, errNum: 0").
		Desc("result: {\"sn\":11,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,10],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"13856901234\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"重\"}]},{\"bg\":0,\"cw\":[{\"sc\":0..00,\"w\":\"六\"}]}]}").
		Desc("audioLen: 488, audioStatus: 2, errNum: 0").
		Desc("audioLen: 488, audioStatus: 2, errNum: 0").
		Desc("result: {\"sn\":12,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,11],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"13856901234\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"充\"}]},{\"bg\":0,\"cw\":[{\"sc\":0..00,\"w\":\"六\"}]}]}").
		Desc("audioLen: 488, audioStatus: 2, errNum: 0").
		Desc("result: {\"sn\":13,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,12],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"13856901234\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"充\"}]},{\"bg\":0,\"cw\":[{\"sc\":0..00,\"w\":\"六\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"千\"}]}]}").
		Desc("result: {\"sn\":14,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,13],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"13856901234\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"充\"}]},{\"bg\":0,\"cw\":[{\"sc\":0..00,\"w\":\"6800\"}]}]}").
		Desc("audioLen: 488, audioStatus: 2, errNum: 0").
		Desc("result: {\"sn\":15,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,14],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"13856901234\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"充\"}]},{\"bg\":0,\"cw\":[{\"sc\":0..00,\"w\":\"6880\"}]}]}").
		Desc("result: {\"sn\":16,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,15],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"13856901234\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"充\"}]},{\"bg\":0,\"cw\":[{\"sc\":0..00,\"w\":\"6888\"}]}]}").
		Desc("audioLen: 488, audioStatus: 2, errNum: 0").
		Desc("result: {\"sn\":17,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,16],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"13856901234\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"充\"}]},{\"bg\":0,\"cw\":[{\"sc\":0..00,\"w\":\"6888\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"点\"}]}]}").
		Desc("result: {\"sn\":18,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,17],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"13856901234\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"充\"}]},{\"bg\":0,\"cw\":[{\"sc\":0..00,\"w\":\"6888\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\".\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"8\"}]}]}").
		Desc("audioLen: 488, audioStatus: 2, errNum: 0").
		Desc("result: {\"sn\":19,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,18],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"13856901234\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"充\"}]},{\"bg\":0,\"cw\":[{\"sc\":0..00,\"w\":\"6888\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\".\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"8\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"元\"}]}]}").
		Desc("audioLen: 488, audioStatus: 2, errNum: 0").
		Desc("result: {\"sn\":20,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,19],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"13856901234\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"充\"}]},{\"bg\":0,\"cw\":[{\"sc\":0..00,\"w\":\"6888\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\".\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"8\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"元\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"话费\"}]}]}").
		Desc("audioLen: 488, audioStatus: 2, errNum: 0").
		Desc("result: {\"sn\":21,\"ls\":false,\"bg\":0,\"ed\":0,\"pgs\":\"rpl\",\"rg\":[1,20],\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"给\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"13856901234\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"充\"}]},{\"bg\":0,\"cw\":[{\"sc\":0..00,\"w\":\"6888\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\".\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"8\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"元\"}]},{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"话费\"}]}]}").
		Desc("result: {\"sn\":22,\"ls\":true,\"bg\":0,\"ed\":0,\"pgs\":\"apd\",\"ws\":[{\"bg\":0,\"cw\":[{\"sc\":0.00,\"w\":\"。\"}]}]}")

	data := []byte("data-123")
	text := []byte("text-123")
	eventSSB.Media("audio/L16", "speex-wb", data).
		Media("audio/L16", "utf-8", text)

	return eventSSB
}
