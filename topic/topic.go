package topic

import (
	"fmt"
	"os"
	"reflect"
	"sync"
)

type Topic interface {
	Unsubscribe(channel interface{}) error
	Subscribe(channel interface{}) (Subscription, error)
	Send(event interface{}) error
}

type Subscription interface {
	Unsubscribe()
	Channel() interface{}
	Err() <-chan error
}

type topic struct {
	sync.RWMutex
	name        string
	valueType   reflect.Type
	cases       cases
	subscribers map[interface{}]*subscription
}

type subscription struct {
	topic   *topic
	channel interface{}
	err     chan error
}

type cases []reflect.SelectCase

func (c *cases) remove(index int) {
	length := len(*c)
	(*c)[index], (*c)[length-1] = (*c)[length-1], (*c)[index]
	*c = (*c)[:length-1]
}

func (c *cases) find(channel interface{}) int {
	address := reflect.ValueOf(channel)
	for i := 0; i < len(*c); i++ {
		value := (*c)[i].Chan
		if value == address {
			return i
		}
	}
	return -1
}

func (c *cases) setSendValue(value interface{}) {
	length := len(*c)
	v := reflect.ValueOf(value)
	for i := 0; i < length; i++ {
		(*c)[i].Send = v
	}
}

func NewTopic(name string) Topic {
	return &topic{
		name:        name,
		subscribers: make(map[interface{}]*subscription),
		cases:       cases{},
		valueType:   nil,
	}
}

func (t *topic) Subscribe(channel interface{}) (Subscription, error) {
	chanVal := reflect.ValueOf(channel)
	chanType := chanVal.Type()

	if chanType.Kind() != reflect.Chan {
		return nil, fmt.Errorf("cannot subscribe using non-channel type")
	}
	if chanType.ChanDir() != reflect.BothDir && chanType.ChanDir() != reflect.SendDir {
		return nil, fmt.Errorf("channel used for subscription must allow sending")
	}
	if !t.checkElementType(chanType.Elem()) {
		return nil, fmt.Errorf("incorrect element type given to feed %s", t.name)
	}
	sub := &subscription{
		topic:   t,
		channel: channel,
		err:     make(chan error, 1),
	}
	t.Lock()
	defer t.Unlock()
	newCase := reflect.SelectCase{
		Dir:  reflect.SelectSend,
		Chan: reflect.ValueOf(channel),
	}
	t.subscribers[channel] = sub
	t.cases = append(t.cases, newCase)
	return sub, nil
}

func (t *topic) Unsubscribe(channel interface{}) error {
	t.Lock()
	defer t.Unlock()
	delete(t.subscribers, channel)
	index := t.cases.find(channel)
	if index == -1 {
		return fmt.Errorf("unable to remove channel")
	}
	t.cases.remove(index)
	return nil
}

func (t *topic) Send(event interface{}) error {
	value := reflect.ValueOf(event)
	valueType := value.Type()

	if !t.checkElementType(valueType) {
		return fmt.Errorf("cannot send event with type %v to channel with type %v", valueType, t.valueType)
	}

	t.Lock()
	defer t.Unlock()
	cases := t.cases
	cases.setSendValue(event)
	for len(cases) != 0 {
		index, _, ok := reflect.Select(cases)
		if !ok {
			fmt.Fprintf(os.Stderr, "while sending to case %v", cases[index])
		}
		ready := cases[index].Chan
		ready.Send(reflect.ValueOf(event))
		cases.remove(index)
	}
	cases.setSendValue(struct{}{})

	return nil
}

func (t *topic) checkElementType(valueType reflect.Type) bool {
	t.Lock()
	defer t.Unlock()
	if t.valueType == nil {
		t.valueType = valueType
		return true
	}
	return t.valueType == valueType
}

func (s *subscription) Unsubscribe() {
	s.topic.Unsubscribe(s.channel)
}

func (s *subscription) Channel() interface{} {
	return s.channel
}

func (s *subscription) Err() <-chan error {
	return s.err
}
