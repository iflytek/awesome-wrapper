package mock

const (
	// unknown type.
	UNKNOWN int32 = 0
	// for CLIENT_SEND or CLIENT_RECV.
	CLIENT int32 = 1
	// for SERVER_RECV or SERVER_SEND.
	SERVER int32 = 2
	// for MESSAGE_SEND.
	PRODUCER int32 = 3
	// for MESSAGE_RECV.
	CONSUMER int32 = 4
)

var (
	DumpEnable = false
	DumpDir    = ""

	FlushRetryCount = 3
	DeliverEnable   = true
	ForceDeliver    = false

	SpillEnable = true
	SpillDir    = ""

	WatchLogEnable = true

	BuffSize  int32 = 2048
	BatchSize       = 100
	LingerSec       = 5

	Logger interface{}
)

func Init(flumeHost string, flumePort string, num int, serviceIP string, servicePort string, serviceName string) error {
	return nil
}

func Flush(span *Span) error {
	return nil
}

func Fini() {

}
