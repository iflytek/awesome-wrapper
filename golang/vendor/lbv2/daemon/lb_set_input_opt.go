package daemon

type SetInPutOpt func(*SetInPut)

func withSetAddr(addr string) SetInPutOpt {
	return func(in *SetInPut) {
		in.addr = addr
	}
}
func withSetSvc(svc string) SetInPutOpt {
	return func(in *SetInPut) {
		in.svc = svc
	}
}
func withSetSubSvc(subSvc string) SetInPutOpt {
	return func(in *SetInPut) {
		in.subSvc = subSvc
	}
}
func withSetTotal(total int64) SetInPutOpt {
	return func(in *SetInPut) {
		in.total = total
	}
}
func withSetIdle(idle int64) SetInPutOpt {
	return func(in *SetInPut) {
		in.idle = idle
	}
}
func withSetBest(best int64) SetInPutOpt {
	return func(in *SetInPut) {
		in.best = best
	}
}

type SetInPut struct {
	addr              string
	svc               string
	subSvc            string
	total, idle, best int64
}
