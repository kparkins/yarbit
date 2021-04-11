package feed

import (
	"fmt"
	"reflect"
	"sync"
)

type Feed interface {
	Subscribe(chan interface{}) (Subscription, error)
	Send(event interface{}) error
}

type Subscription interface {
	Channel() <-chan interface{}
	Err() <-chan error
}

type feed struct {
	sync.RWMutex
	name        string
	valueType   reflect.Type
	subscribers map[*subscription]bool
}

type subscription struct {
	channel chan interface{}
	err     chan error
}

func NewFeed(name string) Feed {
	return &feed{
		name:        name,
		subscribers: make(map[*subscription]bool),
		valueType:   nil,
	}
}

func (f *feed) Subscribe(channel chan interface{}) (Subscription, error) {
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
	f.subscribers[sub] = true
	return sub, nil
}

func (f *feed) Send(event interface{}) error {
	return nil
}

func (f *feed) checkElementType(t reflect.Type) bool {
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
