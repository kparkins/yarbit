package topic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopicSubscribe(t *testing.T) {
	topic := NewTopic("test")
	ch := make(chan int)

	_, err := topic.Subscribe(ch)
	assert.Nil(t, err)

	go topic.Send(1)
	assert.Equal(t, 1, <-ch)
}

func TestTopicUnsubscribe(t *testing.T) {
	topic := NewTopic("test")
	ch := make(chan int)

	_, err := topic.Subscribe(ch)
	assert.Nil(t, err)

	err = topic.Unsubscribe(ch)
	assert.Nil(t, err)

	err = topic.Unsubscribe(ch)
	assert.NotNil(t, err)
}
