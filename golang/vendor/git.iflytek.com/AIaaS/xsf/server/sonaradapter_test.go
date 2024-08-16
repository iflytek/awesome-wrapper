package xsf

import (
	"github.com/cihub/seelog"
	"log"
	"testing"
)

func Test_SonarWithLog(t *testing.T) {
	logcfg := `<seelog>
    <outputs formatid="main">
        <filter levels="info,debug,critical,error">
            <console />
        </filter>
        <filter levels="debug">
            <file path="debug.txt" />
        </filter>
    </outputs>
    <formats>
        <format id="main" format="%Date/%Time [%LEV] %Msg%n"/>
    </formats>
</seelog>`
	logger, logErr := seelog.LoggerFromConfigAsBytes([]byte(logcfg))
	if logErr != nil {
		seelog.Critical("err parsing config log file", logErr)
	}
	s := SonarAdapter{}
	InitSonarErr := s.initSonar(
		WithSonarAdapterAble(false),
		WithSonarAdapterMetricEndpoint("ds"),
		WithSonarAdapterMetricServiceName("127.0.0.1"),
		WithSonarAdapterMetricPort("9090"),
		WithSonarAdapterLogger(logger),
		WithSonarAdapterSonarDumpEnable(true),
		WithSonarAdapterSonarDeliverEnable(false),
		WithSonarAdapterSonarHost("127.0.0.1"),
		WithSonarAdapterSonarPort("9090"),
		WithSonarAdapterSonarBackend(4))
	if InitSonarErr != nil {
		log.Panic(InitSonarErr)
	}
	s.NewMetric("mem", 1234567, KV{"is_open", true})
}
