package main

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
	"zaplogWrap"

	"go.uber.org/zap"
)

var (
	AccumulatePoolInst *AccumulatePool = &AccumulatePool{
		updateTicker:  cycleTime,
		countRecord:   make(map[string]*RecordElement, 50),
		historyRecord: make(map[string]*HistoryRecord, 50),
		node:          make(map[int]bool, 10),
	}
	mistakeRefusedNum uint32 = 0
	passNum           uint32 = 0
	isPass            bool   = false
	diffCnt           int32
	baseCnt           uint32
	isSuppression     bool = true
	deviationNum      int32
)

// for pid
var (
	deltaControl float32
	//	pDeviation float32
	pDeviation    int32
	preeDeviation float32
	Kp            float32 = 1
	Ki            float32 = 0
	Kd            float32 = 0
	delta         float32
)

type HistoryRecord struct {
	cursorBegin int
	cursorEnd   int
	d           []uint32
}

type RecordElement struct {
	base uint32
	diff uint32
}

type AccumulatePool struct {
	updateTicker  time.Duration
	historyRecord map[string]*HistoryRecord
	node          map[int]bool
	nodeLock      sync.RWMutex

	countRecord     map[string]*RecordElement
	countRecordLock sync.RWMutex

	mirrorRecord     map[string]*RecordElement
	mirrorRecordLock sync.RWMutex
}

func init() {
	//AccumulatePoolInst.historyRecord[APPID] = &HistoryRecord{
	//	d:         make([]uint32, 100),
	//	cursorEnd: 1,
	//}
	//AccumulatePoolInst.Run()

	// print realUsed
	go func() {
		for {
			//zaplogWrap.Logger.Info("mistakeRefused" , zap.Uint32("mistake" , mistakeRefusedNum) , zap.Uint32("passed" , passNum))
			zaplogWrap.Logger.Info("realUsed", zap.Int32("realUsed", entireUsed))
			time.Sleep(1 * time.Millisecond)
		}

	}()
}

// start routine to updating data to mirror for control module
func (a *AccumulatePool) Run() {
	t := time.NewTicker(a.updateTicker)
	go func() {
		for {
			select {
			case <-t.C:
				{
					//zaplogWrap.Logger.Info("update countRecord")

					a.mirrorRecordLock.Lock()
					a.mirrorRecord = a.countRecord

					// for pid
					//cDeviation := float32(THRESHOLD - a.countRecord[APPID].base )

					//var cDeviation int32= 0
					_, ok := a.countRecord[APPID]
					if ok {
						zaplogWrap.Logger.Info("countRecord data", zap.Uint32("base", a.mirrorRecord[APPID].base))
						//cDeviation = int32(THRESHOLD) - int32(a.countRecord[APPID].base)
					} else {
						zaplogWrap.Logger.Info("nil", zap.Int("nil", 1))
					}
					a.mirrorRecordLock.Unlock()
					//diffCnt = 0
					//cDeviation := float32(int32(THRESHOLD) - int32(entireUsed))
					//baseCnt = uint32(entireUsed)

					// for pid
					//delta = Kp*(cDeviation - pDeviation) + Ki*cDeviation + Kd*(cDeviation-2*pDeviation + preeDeviation)
					//deltaControl += delta
					//zaplogWrap.Logger.Info("delta data" , zap.Int32("deltaControl" , int32(deltaControl)), zap.Int32("delta" , int32(delta)) , zap.Int32("cDeviation" , int32(cDeviation)) , zap.Int32("pDeviation", int32(pDeviation)))
					//preeDeviation = pDeviation

					//	if cDeviation < 0 {
					//		if -cDeviation > deviationNum {
					//			deviationNum = -cDeviation

					//		}
					//	}

					//	if cDeviation < pDeviation {
					//		isSuppression = true
					//		zaplogWrap.Logger.Info("suppersion" , zap.Int32("isSuppression" , deviationNum))
					//	}else {
					//		isSuppression = false
					//		zaplogWrap.Logger.Info("no suppersion")
					//		zaplogWrap.Logger.Info("suppersion" , zap.Int("isSuppression" , 0))
					//	}
					//	pDeviation = cDeviation

					a.countRecordLock.Lock()
					a.countRecord = make(map[string]*RecordElement, 50)
					a.node = make(map[int]bool, 10)
					a.countRecordLock.Unlock()
				}
			}
		}
	}()
}

// Do MockEngine
func (a *AccumulatePool) DoMockEngine(appid string) {
	randSeed := rand.New(rand.NewSource(time.Now().UnixNano()))
	index := int(randSeed.Uint32()) % SLAVENUM
    //fmt.Println(index)
	//zaplogWrap.Logger.Info("index", zap.Int("index", index))
	s := SlaveNodeMap[index]
	s.MockServe(appid)

}

// control module
func (a *AccumulatePool) ConcurrentJudge(appid string) bool {
	a.mirrorRecordLock.RLock()
	v, ok := a.mirrorRecord[APPID]
	if !ok {
		a.mirrorRecord[APPID] = &RecordElement{
			base: 0,
			diff: 1,
		}
		//	zaplogWrap.Logger.Info("mirror have no this appid")
		a.mirrorRecordLock.RUnlock()
		return true
	}
	a.mirrorRecordLock.RUnlock()

	baseCnt := v.base
	//baseCnt := entireUsed
	//if baseCnt >= THRESHOLD {
	//	//if checkEntireUsed(){
	//	//	mistakeRefusedNum +=1
	//	//}
	//	return false
	//}
	diffcnt := atomic.LoadUint32(&v.diff)
	floatThres := THRESHOLD
	//	if isSuppression == true {
	//		floatThres -= uint32(deviationNum)
	//	}
	if diffcnt+baseCnt >= floatThres {
		//if checkEntireUsed(){
		//	mistakeRefusedNum +=1
		//}
		return false
	}
	atomic.AddUint32(&v.diff, 1)
	randSeed := rand.New(rand.NewSource(time.Now().UnixNano()))
	index := int(randSeed.Uint32()) % SLAVENUM
	s := SlaveNodeMap[index]
	atomic.AddUint32(&passNum, 1)
	//diffCnt+=1
	s.MockServe(APPID)
	return true
}

// update countRecord multiply
func (a *AccumulatePool) Report(appid string, id int, used uint32) {
	a.countRecordLock.Lock()
	defer a.countRecordLock.Unlock()
	v, ok := a.countRecord[appid]
	//s := a.historyRecord[APPID]
	if !ok {
		a.countRecord[appid] = &RecordElement{
			base: used,
			diff: 0,
		}
		//	s.d[s.cursorEnd] = used
		//	s.cursorEnd += 1
		//	s.cursorBegin += 1
		//	if s.cursorEnd >= 100 {
		//		s.cursorEnd = 0
		//	}
		//	if s.cursorBegin >= 100 {
		//		s.cursorBegin = 0
		//	}

	}

	// FIXME may exceed
	if ok := a.node[id]; !ok {
		a.node[id] = true
		if v != nil {
			atomic.AddUint32(&v.base, used)

			//		s.d[s.cursorEnd] = used
			//		s.cursorEnd += 1
			//		s.cursorBegin += 1
			//		if s.cursorEnd >= 100 {
			//			s.cursorEnd = 0
			//		}
			//		if s.cursorBegin >= 100 {
			//			s.cursorBegin = 0
			//		}
		}
	}
}

func MonotonicCheck(s *HistoryRecord) bool {
	//for i:= 1 ;i < len(s) ; i++ {
	//	if s[i] > s[i-1] {

	//	}
	//}
	if s.d[s.cursorEnd]-s.d[s.cursorBegin] > 0 {
		return true
	}
	return false
}
