package event

//事件监听者
type Listener interface {
	Handle(e *Event)
}
