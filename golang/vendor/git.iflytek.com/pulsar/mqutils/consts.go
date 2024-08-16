package mqutils

const (
	DEFAULT_LOG_PATH = "/log/server/pulsar.log"
	LOGFILE_IS_DIR   = "logfile is a directory"
)

const (
	ERR_TOPIC_FORMAT     = "pulsar shoud in format (persistent|inpersistent)://tenant/namespace/topic"
	ERR_INVALID_URL      = "invalid mq url"
	ERR_NO_SUCH_CONSUMER = "no such consumer"
	ERR_CONNECT_BROKEN   = "mq net connecttion is broken"
	ERR_DEADLINE         = "context deadline exceed"
)
