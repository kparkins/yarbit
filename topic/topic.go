package feed

import (
	"fmt"
	"os"
	"reflect"
	"sync"
)

type Topic interface {
	Subscribe(chan interface{}) (Subscription, error)
	Send(event interface{}) error
}

type Subscription interface {
	Channel() <-chan interface{}
	Err() <-chan error
}

type topic struct {
	sync.RWMutex
	name        string
	valueType   reflect.Type
	cases       []reflect.SelectCase
	subscribers map[chan interface{}]*subscription
}

type subscription struct {
	channel chan interface{}
	err     chan error
}

func NewTopic(name string) Topic {
	return &topic{
		name:        name,
		subscribers: make(map[chan interface{}]*subscription),
		cases:       make([]reflect.SelectCase, 0),
		valueType:   nil,
	}
}

func (f *topic) Subscribe(channel chan interface{}) (Subscription, error) {
	chanVal := reflect.ValueOf(channel)
	chanType := chanVal.Type()

	if chanType.Kind() != reflect.Chan {
		return nil, fmt.Errorf("cannot subscribe using non-channel type")
	}
	if chanType.ChanDir() != reflect.BothDir && chanType.ChanDir() != reflect.SendDir {
		return nil, fmt.Errorf("channel used for subscription must allow sending")
	}
	if !f.checkElementType(chanType.Elem()) {
		return nil, fmt.Errorf("incorrect element type given to feed %s", f.name)
	}
	sub := &subscription{
		channel: channel,
		err:     make(chan error, 1),
	}
	f.Lock()
	defer f.Unlock()
	newCase := reflect.SelectCase {
		Dir: reflect.SelectSend, 
		Chan: reflect.ValueOf(channel)
	}
	f.subscribers[channel] = sub
	f.cases = append(f.cases, newCase)
	return sub, nil
}

func (f *topic) Send(event interface{}) error {
	eventVal := reflect.ValueOf(event)
	eventType := eventVal.Type()

	if !f.checkElementType(eventType) {
		return fmt.Errorf("cannot send event with type %v to channel with type %v", eventType, f.valueType)
	}

	cases := make([]reflect.SelectCase, len(f.cases))
	copy(cases, f.cases)
	for len(cases) != 0 {
		index, _, ok := reflect.Select(cases)
		if !ok {
			fmt.Fprintf(os.Stderr, "while sending to case %v", cases[index])
		}
		ready := cases[index].Chan
		ready.Send(reflect.ValueOf(event))
	}

	return nil
}

func (f *topic) checkElementType(t reflect.Type) bool {
	if f.valueType == nil {
		f.valueType = t
		return true
	}
	return f.valueType == t
}

func (s *subscription) Channel() <-chan interface{} {
	return s.channel
}

func (s *subscription) Err() <-chan error {
	return s.err
}
