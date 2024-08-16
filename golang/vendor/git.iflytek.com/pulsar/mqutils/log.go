package mqutils

import (
	utils "git.iflytek.com/AIaaS/xsf/utils"
	"github.com/apache/pulsar-client-go/pulsar/log"
)

type nopEntry struct{}

func (e nopEntry) WithFields(fields log.Fields) log.Entry             { return nopEntry{} }
func (e nopEntry) WithField(name string, value interface{}) log.Entry { return nopEntry{} }

func (e nopEntry) Debug(args ...interface{})                 {}
func (e nopEntry) Info(args ...interface{})                  {}
func (e nopEntry) Warn(args ...interface{})                  {}
func (e nopEntry) Error(args ...interface{})                 {}
func (e nopEntry) Debugf(format string, args ...interface{}) {}
func (e nopEntry) Infof(format string, args ...interface{})  {}
func (e nopEntry) Warnf(format string, args ...interface{})  {}
func (e nopEntry) Errorf(format string, args ...interface{}) {}

type LogWrapper struct {
	log *utils.Logger
}

func (log *LogWrapper) SubLogger(fields log.Fields) log.Logger {
	//log.log.Errorf()
	return log
}
func (log *LogWrapper) WithFields(fields log.Fields) log.Entry {
	return nopEntry{}
}
func (log *LogWrapper) WithField(name string, value interface{}) log.Entry {
	return nopEntry{}
}
func (log *LogWrapper) WithError(err error) log.Entry {
	return nopEntry{}
}

func (log *LogWrapper) Debug(args ...interface{}) {
	log.log.Debugf("", args)
}
func (log *LogWrapper) Info(args ...interface{}) {
	log.log.Infof("", args)
}
func (log *LogWrapper) Warn(args ...interface{}) {
	log.log.Warnf("", args)
}
func (log *LogWrapper) Error(args ...interface{}) {
	log.log.Errorf("", args)
}

func (log *LogWrapper) Debugf(format string, args ...interface{}) {
	log.log.Debugf(format, args)
}
func (log *LogWrapper) Infof(format string, args ...interface{}) {
	log.log.Infof(format, args)
}
func (log *LogWrapper) Warnf(format string, args ...interface{}) {
	log.log.Warnf(format, args)
}
func (log *LogWrapper) Errorf(format string, args ...interface{}) {
	log.log.Errorf(format, args)
}
