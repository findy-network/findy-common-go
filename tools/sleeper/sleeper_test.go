package sleeper

import (
	"testing"
	"time"

	"github.com/lainio/err2/assert"
)

func TestCount(t *testing.T) {
	assert.PushTester(t)
	defer assert.PopTester()
	myFloor := time.Second * 1.0
	s := New(myFloor)
	assert.Equal(0, s.count)

	var dur time.Duration
	sleep := func(d time.Duration) {
		dur = d
	}

	r := s.Sleep(sleep)
	assert.That(r)
	assert.Equal(1, s.count)
	assert.Equal(myFloor+time.Second*2, dur)

	r = s.Sleep(sleep)
	assert.That(r)
	assert.Equal(2, s.count)
	assert.Equal(myFloor+time.Second*4, dur)

	r = s.Sleep(sleep)
	assert.That(r)
	assert.Equal(3, s.count)
	assert.Equal(myFloor+time.Second*8, dur)

	r = s.Sleep(sleep)
	assert.That(r)
	assert.Equal(4, s.count)
	assert.Equal(myFloor+time.Second*16, dur)
}
