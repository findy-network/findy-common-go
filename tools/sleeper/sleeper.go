package sleeper

import (
	"math"
	"time"
)

const (
	floor    = 10 * time.Second
	maxCount = 12
)

type Sleeper struct {
	count    int
	floor    time.Duration
	maxCount int
}

func New(f time.Duration) *Sleeper {
	return &Sleeper{count: 0, floor: f, maxCount: maxCount}
}

func (s *Sleeper) Sleep(sleep func(time.Duration)) bool {
	x := 2.0 // base for our exponential function
	if s.count < s.maxCount {
		s.count++
	}
	round := float64(s.count)
	v := time.Duration(math.Pow(x, round))
	sleep(s.floor + v*time.Second)
	return true
}
