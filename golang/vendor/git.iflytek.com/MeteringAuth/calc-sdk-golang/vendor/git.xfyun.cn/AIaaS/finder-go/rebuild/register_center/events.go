package register_center

type EventType string



type Event interface {
	Type()EventType
	Data()interface{}
}


