package comwrapper

type DataType int
type DataStatus int
type CallBackPtr func(hdl interface{}, resp []WrapperData)
type CustomFuncType int

const (
	DataText  DataType = 0 // 文本数据
	DataAudio DataType = 1 // 音频数据
	DataImage DataType = 2 // 图像数据
	DataVideo DataType = 3 // 视频数据
	DataPer   DataType = 4 // 个性化数据

	DataBegin    DataStatus = 0 // 首数据
	DataContinue DataStatus = 1 // 中间数据
	DataEnd      DataStatus = 2 // 尾数据
	DataOnce     DataStatus = 3 // 非会话单次输入
)

type WrapperData struct {
	Key      string            // 数据标识
	Data     []byte            // 数据实体
	Desc     map[string]string // 数据描述
	Encoding string            // 数据编码
	Type     DataType          // 数据类型
	Status   DataStatus        // 数据状态
}

const (
	FuncTraceLog CustomFuncType = 0
	FunnMeter    CustomFuncType = 1
)
