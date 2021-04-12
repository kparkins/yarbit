package feed

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFeedSubscribe(t *testing.T) {
	feed := NewTopic("test")
	ch := make(chan<- int, 0)
	val := reflect.ValueOf(ch)
	ctype := val.Type()
	otype := ctype.Elem()
	dir := ctype.ChanDir()
	fmt.Printf("%v\n", val)
	fmt.Printf("%v\n", ctype)
	fmt.Printf("%v\n", otype)
	if dir == reflect.BothDir {
		fmt.Println("bopth dir")
	} else if dir == reflect.SendDir {
		fmt.Println("send dir")
	}
	assert.NotNil(t, feed)
}
