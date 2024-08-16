package licc

type CtrlMode uint16

const (
	//
	CtrlNone = iota

	//
	CtrlDayFlow CtrlMode = 1 << (iota - 1) // 日流控
	//CtrlHourFlow                                  // 小时流控
	CtrlTimeFlow  // 时授
	CtrlCountFlow // 量授
	CtrlUserFlow  // 用户级
	CtrlFreeFlow  // 免费次数
	CtrlSecFlow   // 秒级流控
	CtrlConcFlow  // 并发
	CtrlTail

	// 覆盖所有类型位（用于通过&获取类型位）
	CtrlALL = CtrlDayFlow | CtrlSecFlow | CtrlCountFlow | CtrlConcFlow | CtrlUserFlow | CtrlFreeFlow
)
