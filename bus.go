package event

import (
	"fmt"
	libLog "log"
	"runtime/debug"
	"sync"
	"time"
)

//事件总线
type Bus struct {
	//事件总线名称
	name string
	//事件与监听者容器列表
	savers []*Saver
	//关闭状态
	closed chan struct{}
	//事件异步处理队列
	task chan *Event
	//事件队列处理go程退出等待
	taskWaitGroup *sync.WaitGroup
	//事件队列处理go程退出时候的等待超时时间
	timeout time.Duration
	//关闭锁
	lock *sync.Mutex
}

func New(name string) *Bus {
	return &Bus{
		name:    name,
		savers:  make([]*Saver, 0),
		closed:  make(chan struct{}),
		task:    nil,
		timeout: 2 * time.Second,
		lock:    new(sync.Mutex),
	}
}

//开启异步调度事件
func (this *Bus) Async(worker int, capacity int) {
	if this.task == nil {
		this.task = make(chan *Event, capacity)
		this.taskWaitGroup = &sync.WaitGroup{}
		for i := 0; i < worker; i++ {
			this.taskWaitGroup.Add(1)
			go this.goF()
		}
	}
}

//并发函数
func (this *Bus) goF() {
	defer func() {
		if a := recover(); a != nil {
			//记录错误栈
			var message string
			message = fmt.Sprintf("Events %s uncaught panic: %s", this.name, debug.Stack())
			libLog.Println(message)
			//重启一条go程
			go this.goF()
		} else {
			//退出go程
			this.taskWaitGroup.Done()
		}
	}()
	for {
		select {
		//接收到事件，进行调度
		case e := <-this.task:
			this.dispatch(e)
			break
		//接收到事件队列关闭信号
		case <-this.closed:
			for {
				select {
				case e := <-this.task:
					this.dispatch(e)
					break
				case <-time.After(this.timeout):
					return
				}
			}
		}
	}
}

//关闭事件总线，
func (this *Bus) Close(timeout ...time.Duration) {
	//获取锁
	this.lock.Lock()
	defer this.lock.Unlock()

	//判断是否关闭
	select {
	case <-this.closed:
		return
	default:
		break
	}
	//发出关闭信号
	close(this.closed)
	if this.task != nil {
		if len(timeout) > 0 {
			this.timeout = timeout[0]
		}
		//等待所有事件队列处理go程退出
		this.taskWaitGroup.Wait()
		//彻底关闭事件队列
		close(this.task)
	}
}

//添加某个事件的某个监听者
func (this *Bus) AddListener(name string, l Listener) {
	for _, saver := range this.savers {
		if saver.Name == name {
			saver.Listeners = append(saver.Listeners, l)
			return
		}
	}
	saver := NewSaver(name)
	saver.Listeners = append(saver.Listeners, l)
	this.savers = append(this.savers, saver)
}

//调度一个事件
func (this *Bus) dispatch(event *Event) {
	for _, saver := range this.savers {
		if saver.Name != event.Name {
			continue
		}
		for _, l := range saver.Listeners {
			l.Handle(event)
		}
		break
	}
}

//触发一个事件
func (this *Bus) Trigger(event *Event) {
	//判断事件总线关闭状态
	select {
	case <-this.closed:
		//已经关闭，不在处理事件
		return
	default:
		break
	}
	if this.task != nil {
		//发送事件到事件队列
		this.task <- event
	} else {
		//同步处理事件，直接调度
		this.dispatch(event)
	}
}

//新增事件并触发它
func (this *Bus) Append(name string, data interface{}) {
	this.Trigger(NewEvent(name, data))
}
