package zaplogWrap
import (
	"go.uber.org/zap"
//	"go.uber.org/zap/zapcore"
	//"encoding/json"
)

var (
	Logger *zap.Logger
	cfg zap.Config
)


func init(){
	//rawJSON := []byte(`{
	//	"level": "debug",
	//	"encoding": "json",
	//	"outputPaths": ["stdout", "./log/concurrent.log"],
	//	"errorOutputPaths": ["stderr"],
	//	"encoderConfig": {
	//		"messageKey": "message",
	//		"levelKey": "level",
	//		"levelEncoder": "lowercase",
	//		"timeKey" : "time"
	//	}


	//}`)
	
	//encoderConfig := zapcore.EncoderConfig{
	//	TimeKey:        "time",
	//	LevelKey:       "level",
	//	//NameKey:        "logger",
	//	CallerKey:      "caller",
	//	MessageKey:     "msg",
	//	StacktraceKey:  "stacktrace",
	//	LineEnding:     zapcore.DefaultLineEnding,
	//	EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
	//	EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
	//	EncodeDuration: zapcore.SecondsDurationEncoder,
	//	EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器

	//}

	// 设置日志级别
	atom := zap.NewAtomicLevelAt(zap.DebugLevel)

	cfg:= zap.Config{
		Level:            atom,                                                // 日志级别
		Development:      true,                                                // 开发模式，堆栈跟踪
		Encoding:         "json",                                              // 输出格式 console 或 json
	//	EncoderConfig:    encoderConfig,                                       // 编码器配置
	//	InitialFields:    map[string]interface{}{"serviceName": "concurrent"}, // 初始化字段，如：添加一个服务器名称
		OutputPaths:      []string{"stdout", "./log/concurrent.log"},         // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
		ErrorOutputPaths: []string{"stderr"},

	}

	//var cfg zap.Config
	//if err := json.Unmarshal(rawJSON, &cfg); err != nil {
	//	panic(err)

	//}
	var err error
	Logger , err = cfg.Build()
	if err != nil {
		panic(err)
	}
	//	var err error
	//	Logger , err = zap.NewDevelopment()
	//	if err != nil {
	//		panic(err)
	//	}
}

func Sync(){
	Logger.Sync()
}

