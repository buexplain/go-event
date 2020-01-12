package main

import (
	"github.com/buexplain/go-event"
	"log"
	"time"
)

var bus *event.Bus

func init()  {
	//初始化一个事件调度器
	bus = event.New("test")
	//设置为异步多go程处理，go程数量是3，缓冲通道容量为25
	bus.Async(3, 25)
	//添加一个事件的监听器
	bus.AddListener(TEST_NAME, TestListener{})
}

const TEST_NAME  = "test_event_name"
type TestListener struct {
}
func (this TestListener) Handle(e *event.Event)() {
	<-time.After(1*time.Second)
	log.Printf("%s %+v\n", e.Name, e.Data)
}

func main() {
	sendOk := make(chan struct{})
	go func() {
		defer func() {
			sendOk <- struct{}{}
		}()
		for i:=0; i<50; i++ {
			bus.Append(TEST_NAME, i)
		}
	}()
	<- sendOk
	bus.Close()
}
