package daemon

type SubSvcItem struct {
	cas       int64 //事务性 0：可修改 1：修改中
	dead      int64 //0：alive 1：dead
	addr      string
	ttl       int64 //纳秒,定时清除无效节点的时间间隔
	timestamp int64 //纳秒,节点最近一次上报的时间
	bestInst  int64
	idleInst  int64
	totalInst int64
}

func (s *SubSvcItem) FInit() *SubSvcItem {
	s.cas = 0
	s.addr = ""
	s.bestInst = 0
	s.idleInst = 0
	s.totalInst = 0
	return s
}

type SubSvcItemSlice []*SubSvcItem

func (l SubSvcItemSlice) Len() int {
	return len(l)
}
func (l SubSvcItemSlice) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (l SubSvcItemSlice) Less(i, j int) bool {
	return l[i].idleInst < l[j].idleInst
}
