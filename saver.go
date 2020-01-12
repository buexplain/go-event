package event

//事件与监听者容器
type Saver struct {
	Name      string
	Listeners []Listener
}

func NewSaver(name string) *Saver {
	return &Saver{Name: name, Listeners: make([]Listener, 0)}
}
