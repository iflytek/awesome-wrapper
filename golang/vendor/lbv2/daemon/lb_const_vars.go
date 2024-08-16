package daemon

const (
	NBESTTAG = "nbest"
	ALL      = "all"
	REPORTER = "set"
	CLIENT   = "get"
	UID      = "uid"
	EXPARAM  = "exparam"
	/*-------指定组件的版本号--------*/
	LBVERSION = "0.1.0"
)
const (
	SEED      = 10000 //节点的分段基数
	LBTAG     = "lbv2"
	SVC       = "svc"
	THRESHOLD = "threshold"
	TICKER    = "ticker"
	SUBSVC    = "subsvc"
	TTL       = "ttl"
	STRATEGY  = "strategy"
	BO        = "bo"
	RMQADDRS  = "rmqaddrs"
	RMQTOPIC  = "rmqtopic"
	RMQGROUP  = "rmqgroup"
	DEBUG     = "debug"
)
const (
	DB          = "db"
	DBBASEURL   = "baseurl"
	DBCALLER    = "caller"
	DBCALLERKEY = "callerkey"
	DBTIMEOUT   = "timeout"
	DBTOKEN     = "token"
	DBVERSION   = "version"
	DBIDC       = "idc"
	DBSCHEMA    = "schema"
	DBTABLE     = "table"
)

//worker
const (
	LICADDR   = "addr"
	LICSVC    = "svc"
	LICSUBSVC = "subsvc"
	LICTOTAL  = "total"
	LICIDLE   = "idle"
	LICBEST   = "best"
)

const (
	//some vars about rmq
	RmqUid        = "common.uid"
	rmqSvcName    = "svc_name"
	rmqSubSvcName = "sub_svc_name"
	notifyOp      = "sync" //todo 后续和珍松沟通，确定
)

type StrategyClassify int

const (
	lic   StrategyClassify = iota
	licEx  //arm分段策略
)

func (s StrategyClassify) String() string {
	switch s {
	case lic:
		return "lic"
	case licEx:
		return "licEx"
	default:
		return "Unknown"
	}
}
