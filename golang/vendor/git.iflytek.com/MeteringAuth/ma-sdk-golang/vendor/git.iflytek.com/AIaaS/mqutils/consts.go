package mqutils

import "errors"

var (
	ErrTopicFormat      = errors.New("pulsar shoud in format (persistent|inpersistent)://tenant/namespace/topic")
	ErrInvalidUrl       = errors.New("invalid mq url")
	ErrInvalidChannel   = errors.New("invalid amqp.channel")
	ErrNoSuchConsumer   = errors.New("no such consumer")
	ErrConnectionBroken = errors.New("mq network connection is broken")
	ErrDeadline         = errors.New("context deadline exceed")
	ErrNotSupport       = errors.New("operation not support")
)
