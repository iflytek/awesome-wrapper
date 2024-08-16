package log


type Logger interface {
	Errorf(s string,args ...interface{})
	Infof(s string,args ...interface{})
	Warnf(s string,args ...interface{})
}
