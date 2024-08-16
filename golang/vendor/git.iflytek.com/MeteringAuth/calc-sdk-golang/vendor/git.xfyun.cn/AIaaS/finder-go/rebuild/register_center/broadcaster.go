package register_center

type Listener interface {
	Type() EventType
	OnMessage(t EventType,data interface{})
}

type BroadCaster interface {
	RegisterListener(ls Listener)
	RemoveListener(ls Listener)
	SendBroadCast(e Event)
}