package main

import (
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"git.iflytek.com/MeteringAuth/janus-sdk-golang/report"
)

var (
	SlaveNodeMap map[int]*Slave = make(map[int]*Slave, SLAVENUM)
	entireUsed   int32
)

type Slave struct {
	sync.RWMutex
	ID       int
	Capacity uint32
	Used     int32
}

func (s *Slave) Run() {
	t := time.NewTicker(time.Duration(reportTime) * time.Millisecond)
	go func() {
		for {
			select {
			case <-t.C:
				{
					// report
					//					zaplogWrap.Logger.Info("report use info" , zap.Int("ID" , s.ID) , zap.Int32("used" , s.Used))
					//AccumulatePoolInst.Report(APPID , s.ID , uint32(s.Used))
					//					zaplogWrap.Logger.Info("report use info" , zap.Int("ID" , s.ID) , zap.Int32("used" , s.Used))
					report.ReportWithAddr(map[string]uint{APPID: uint(s.Used)}, strconv.Itoa(s.ID))

				}
			}
		}
	}()
}

func (s *Slave) MockServe(appid string) {
	//s.Lock()
	//if s.Used >= s.Capacity {
	//	return
	//}
	//s.Unlock()
	atomic.AddInt32(&s.Used, 1)
	atomic.AddInt32(&entireUsed, 1)
	randSeed := rand.New(rand.NewSource(time.Now().UnixNano()))

	index := int(randSeed.Uint32())%int(MockDealTime) + begin
	//s.Used += 1
	//zaplogWrap.Logger.Info("immediately used count" ,zap.Int32("used" , s.Used))
	//time.Sleep(MockDealTime)
	time.Sleep(time.Duration(index) * time.Millisecond)
	atomic.AddInt32(&entireUsed, -1)
	atomic.AddInt32(&s.Used, -1)
}

func checkEntireUsed() bool {
	u := atomic.LoadInt32(&entireUsed)
	if uint32(u) >= THRESHOLD {
		return false
	}
	return true
}

func StartNode() {
	//waitChan := make(chan struct{})
	hostName, _ := os.Hostname()
	addr, _ := net.LookupHost(hostName)
	report.SetCompanionUrl(companionUrl).SetProjectName(projectName).SetGroup(group).SetServiceName(serviceName).SetVersion(version).SetCacheConfig(true).SetCacheService(true).SetCfgMode(1)
	report.Init(CHANNEL, addr[0])

	for i := 0; i < SLAVENUM; i++ {
		s := &Slave{
			ID: i,
		}
		SlaveNodeMap[i] = s
		s.Run()
		time.Sleep(reportLaunch)
	}
	//<-waitChan
	//zaplogWrap.Logger.Info("node finished...")
}
