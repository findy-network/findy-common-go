package sleeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCount(t *testing.T) {
	myFloor := time.Second * 1.0
	s := New(myFloor)
	assert.Equal(t, 0, s.count)

	var dur time.Duration
	sleep := func(d time.Duration) {
		dur = d
	}

	r := s.Sleep(sleep)
	assert.True(t, r)
	assert.Equal(t, 1, s.count)
	assert.Equal(t, myFloor+time.Second*2, dur)

	r = s.Sleep(sleep)
	assert.True(t, r)
	assert.Equal(t, 2, s.count)
	assert.Equal(t, myFloor+time.Second*4, dur)

	r = s.Sleep(sleep)
	assert.True(t, r)
	assert.Equal(t, 3, s.count)
	assert.Equal(t, myFloor+time.Second*8, dur)

	r = s.Sleep(sleep)
	assert.True(t, r)
	assert.Equal(t, 4, s.count)
	assert.Equal(t, myFloor+time.Second*16, dur)
}
