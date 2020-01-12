package event

//事件结构体
type Event struct {
	//事件名称
	Name string
	//事件包含的数据
	Data interface{}
}

func NewEvent(name string, data interface{}) *Event {
	return &Event{Name: name, Data: data}
}
