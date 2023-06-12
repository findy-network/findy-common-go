package x

import "sync"

// RWSlice is a type for a thread-safe Go slice. It tries to be short and simple.
// Tip: It's useful to create a type alias (it allows it):
//
//	type testersSlice = [int]testing.TB
//
// Which shortens and makes easier to read its usage:
//
//	x.TxSlice(testers, func(s testersSlice) {
//	    s[goid()] = val
//	})
type RWSlice[S ~[]T, T any] struct {
	sync.RWMutex
	s S
}

// NewRWSlice creates a new thread-safe slice that's as simple as possible. The
// first version had only two functions Tx and Rx to allow interact with the
// slice.
func NewRWSlice[S ~[]T, T any](size ...int) *RWSlice[S, T] {
	// build in make() have to deal by us
	switch len(size) {
	case 2:
		return &RWSlice[S, T]{s: make([]T, size[0], size[1])}
	default:
		return &RWSlice[S, T]{s: make([]T, size[0])}
	}
}

// Tx executes a update transaction to the thread-safe slice.
func (s *RWSlice[S, T]) Tx(f func(s S)) {
	s.Lock()
	defer s.Unlock()
	f(s.s)
}

// Add adds a value to the thread-safe slice.
func (s *RWSlice[S, T]) Add(val T) T {
	s.Lock()
	defer s.Unlock()
	s.s = append(s.s, val)
	return val
}

// Set sets a value to the slice index in a thread-safe way.
func (s *RWSlice[S, T]) Set(index int, val T) T {
	s.Lock()
	defer s.Unlock()
	s.s[index] = val
	return val
}

// Rx executes thread-safe read transaction i.e. critical section for the
// RWSlice.
func (s *RWSlice[S, T]) Rx(f func(s S)) {
	s.RLock()
	defer s.RUnlock()
	f(s.s)
}

// Get reads thread-safely value from index index.
func (s *RWSlice[S, T]) Get(index int) T {
	s.RLock()
	defer s.RUnlock()
	return s.s[index]
}
